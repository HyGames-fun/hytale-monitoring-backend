package agent

import (
	"context"
	"sync"
	"time"

	"github.com/HyGames-fun/hytale-monitoring-backend/internal/domain"
	"github.com/HyGames-fun/hytale-monitoring-backend/pkg/agentproto"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ServerRepo interface {
	GetServerByID(id int) (*domain.Server, error)
}

type MetricsRepo interface {
	SaveMetric(m *domain.ServerMetric) error
}

type AgentSession struct {
	SessionID string
	ServerID  int
	CreatedAt time.Time
	LastSeen  time.Time
}

type Server struct {
	agentproto.UnimplementedHytaleMonitoringAgentServer
	serverRepo  ServerRepo
	metricsRepo MetricsRepo
	sessions    map[string]*AgentSession
	mu          sync.RWMutex
}

func NewServer(serverRepo ServerRepo, metricsRepo MetricsRepo) *Server {
	return &Server{
		serverRepo:  serverRepo,
		metricsRepo: metricsRepo,
		sessions:    make(map[string]*AgentSession),
	}
}

func (s *Server) RegisterAgent(ctx context.Context, req *RegisterAgentRequestProto) (*agentproto.RegisterAgentResponse, error) {
	if req.ServerId <= 0 {
		return &agentproto.RegisterAgentResponse{
			Success: false,
			Message: "invalid server_id",
		}, nil
	}

	server, err := s.serverRepo.GetServerByID(int(req.ServerId))
	if err != nil || server == nil {
		return &agentproto.RegisterAgentResponse{
			Success: false,
			Message: "server not found",
		}, nil
	}

	if server.VerificationToken != nil && *server.VerificationToken != "" {
		if req.AgentToken != *server.VerificationToken {
			return &agentproto.RegisterAgentResponse{
				Success: false,
				Message: "invalid agent_token",
			}, nil
		}
	}

	sessionID := uuid.New().String()
	s.mu.Lock()
	s.sessions[sessionID] = &AgentSession{
		SessionID: sessionID,
		ServerID:  int(req.ServerId),
		CreatedAt: time.Now(),
		LastSeen:  time.Now(),
	}
	s.mu.Unlock()

	return &agentproto.RegisterAgentResponse{
		Success:               true,
		Message:               "registered successfully",
		SessionId:             sessionID,
		ReportIntervalSeconds: 15,
	}, nil
}

type RegisterAgentRequestProto = agentproto.RegisterAgentRequest

func (s *Server) ReportMetrics(ctx context.Context, req *agentproto.MetricsReportRequest) (*agentproto.MetricsReportResponse, error) {
	s.mu.Lock()
	session, exists := s.sessions[req.SessionId]
	if !exists {
		s.mu.Unlock()
		return nil, status.Errorf(codes.Unauthenticated, "invalid or expired session_id")
	}
	session.LastSeen = time.Now()
	serverID := session.ServerID
	s.mu.Unlock()

	metric := &domain.ServerMetric{
		ServerID:      serverID,
		IsOnline:      true,
		PlayersOnline: int(req.OnlinePlayers),
		PlayersMax:    int(req.MaxPlayers),
		TPS:           req.Tps,
		RAMUsedMB:     int(req.RamUsedBytes / (1024 * 1024)),
		LatencyMS:     1,
		RecordedAt:    time.Now(),
	}

	if err := s.metricsRepo.SaveMetric(metric); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save metric: %v", err)
	}

	return &agentproto.MetricsReportResponse{
		Acknowledged: true,
		ServerTime:   time.Now().Unix(),
	}, nil
}

func (s *Server) Ping(ctx context.Context, req *agentproto.PingRequest) (*agentproto.PingResponse, error) {
	s.mu.Lock()
	session, exists := s.sessions[req.SessionId]
	if exists {
		session.LastSeen = time.Now()
	}
	s.mu.Unlock()

	return &agentproto.PingResponse{
		ServerTimestamp: time.Now().UnixMilli(),
		IsActive:        exists,
	}, nil
}
