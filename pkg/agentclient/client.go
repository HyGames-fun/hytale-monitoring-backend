package agentclient

import (
	"context"
	"fmt"
	"time"

	"github.com/HyGames-fun/hytale-monitoring-backend/pkg/agentproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	TargetAddress string
	ServerID      int32
	AgentToken    string
	ServerVersion string
	ModVersion    string
}

type Client struct {
	cfg       Config
	conn      *grpc.ClientConn
	client    agentproto.HytaleMonitoringAgentClient
	sessionID string
	interval  time.Duration
}

type MetricsData struct {
	OnlinePlayers int32
	MaxPlayers    int32
	TPS           float64
	RAMUsedBytes  int64
	RAMMaxBytes   int64
	UptimeSeconds int64
	PlayerNames   []string
}

func Connect(ctx context.Context, cfg Config) (*Client, error) {
	conn, err := grpc.DialContext(
		ctx,
		cfg.TargetAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to monitoring agent server: %w", err)
	}

	protoClient := agentproto.NewHytaleMonitoringAgentClient(conn)

	regResp, err := protoClient.RegisterAgent(ctx, &agentproto.RegisterAgentRequest{
		ServerId:      cfg.ServerID,
		AgentToken:    cfg.AgentToken,
		ServerVersion: cfg.ServerVersion,
		ModVersion:    cfg.ModVersion,
	})
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("agent registration failed: %w", err)
	}

	if !regResp.Success {
		_ = conn.Close()
		return nil, fmt.Errorf("agent registration rejected: %s", regResp.Message)
	}

	interval := time.Duration(regResp.ReportIntervalSeconds) * time.Second
	if interval <= 0 {
		interval = 15 * time.Second
	}

	return &Client{
		cfg:       cfg,
		conn:      conn,
		client:    protoClient,
		sessionID: regResp.SessionId,
		interval:  interval,
	}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) Report(ctx context.Context, data MetricsData) error {
	resp, err := c.client.ReportMetrics(ctx, &agentproto.MetricsReportRequest{
		SessionId:     c.sessionID,
		ServerId:      c.cfg.ServerID,
		OnlinePlayers: data.OnlinePlayers,
		MaxPlayers:    data.MaxPlayers,
		Tps:           data.TPS,
		RamUsedBytes:  data.RAMUsedBytes,
		RamMaxBytes:   data.RAMMaxBytes,
		UptimeSeconds: data.UptimeSeconds,
		PlayerNames:   data.PlayerNames,
		Timestamp:     time.Now().Unix(),
	})
	if err != nil {
		return fmt.Errorf("failed to report telemetry: %w", err)
	}
	if !resp.Acknowledged {
		return fmt.Errorf("telemetry report not acknowledged by monitoring server")
	}
	return nil
}

func (c *Client) Ping(ctx context.Context) (time.Duration, error) {
	start := time.Now()
	resp, err := c.client.Ping(ctx, &agentproto.PingRequest{
		SessionId:       c.sessionID,
		ClientTimestamp: start.UnixMilli(),
	})
	if err != nil {
		return 0, err
	}
	if !resp.IsActive {
		return 0, fmt.Errorf("agent session is no longer active")
	}
	return time.Since(start), nil
}

func (c *Client) Interval() time.Duration {
	return c.interval
}
