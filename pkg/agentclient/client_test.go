package agentclient

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/HyGames-fun/hytale-monitoring-backend/internal/domain"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/service/agent"
	"github.com/HyGames-fun/hytale-monitoring-backend/pkg/agentproto"
	"google.golang.org/grpc"
)

type mockServerRepo struct{}

func (m *mockServerRepo) GetServerByID(id int) (*domain.Server, error) {
	tok := "secret-test-token"
	return &domain.Server{
		ID:                id,
		Name:              "Mock Server",
		VerificationToken: &tok,
	}, nil
}

type mockMetricsRepo struct {
	lastSaved *domain.ServerMetric
}

func (m *mockMetricsRepo) SaveMetric(metric *domain.ServerMetric) error {
	m.lastSaved = metric
	return nil
}

func TestAgentClient_EndToEnd(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer lis.Close()

	grpcServer := grpc.NewServer()
	sRepo := &mockServerRepo{}
	mRepo := &mockMetricsRepo{}
	srv := agent.NewServer(sRepo, mRepo)
	agentproto.RegisterHytaleMonitoringAgentServer(grpcServer, srv)

	go func() {
		_ = grpcServer.Serve(lis)
	}()
	defer grpcServer.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := Connect(ctx, Config{
		TargetAddress: lis.Addr().String(),
		ServerID:      42,
		AgentToken:    "secret-test-token",
		ServerVersion: "1.0.0",
		ModVersion:    "0.1.0",
	})
	if err != nil {
		t.Fatalf("failed to connect agent client: %v", err)
	}
	defer client.Close()

	latency, err := client.Ping(ctx)
	if err != nil {
		t.Fatalf("ping failed: %v", err)
	}
	if latency < 0 {
		t.Fatalf("invalid latency: %v", latency)
	}

	err = client.Report(ctx, MetricsData{
		OnlinePlayers: 25,
		MaxPlayers:    100,
		TPS:           19.9,
		RAMUsedBytes:  1024 * 1024 * 512,
		RAMMaxBytes:   1024 * 1024 * 1024,
		UptimeSeconds: 3600,
		PlayerNames:   []string{"Player1", "Player2"},
	})
	if err != nil {
		t.Fatalf("report failed: %v", err)
	}

	if mRepo.lastSaved == nil || mRepo.lastSaved.PlayersOnline != 25 {
		t.Fatalf("expected 25 players reported, got %+v", mRepo.lastSaved)
	}
}
