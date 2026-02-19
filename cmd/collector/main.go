package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HyGames-fun/hytale-monitoring-backend/internal/config"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/service/agent"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/service/collector"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/storage/postgres"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/storage/redis"
	"github.com/HyGames-fun/hytale-monitoring-backend/pkg/agentproto"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.Load()
	log.Println("[COLLECTOR] Starting Background Metrics Collector and Telemetry Agent Server")

	db, err := postgres.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[COLLECTOR] Database connection failed: %v", err)
	}
	defer db.Close()

	redisClient, err := redis.Connect(cfg.RedisURL)
	if err != nil {
		log.Fatalf("[COLLECTOR] Redis connection failed: %v", err)
	}
	defer redisClient.Close()

	serverRepo := postgres.NewServerRepo(db)
	metricsRepo := postgres.NewMetricsRepo(db)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	poller := collector.NewPoller(serverRepo, metricsRepo, redisClient, 10, 30*time.Second)
	go poller.Start(ctx)
	log.Println("[COLLECTOR] Periodic node polling worker pool started (10 workers, 30s interval)")

	grpcListener, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("[COLLECTOR] Failed to listen on gRPC port %s: %v", cfg.GRPCPort, err)
	}

	grpcServer := grpc.NewServer()
	agentService := agent.NewServer(serverRepo, metricsRepo)
	agentproto.RegisterHytaleMonitoringAgentServer(grpcServer, agentService)

	go func() {
		log.Printf("[COLLECTOR] gRPC Agent Telemetry Server listening on port %s", cfg.GRPCPort)
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("[COLLECTOR] gRPC server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("[COLLECTOR] Shutting down...")
	cancel()
	grpcServer.GracefulStop()
	log.Println("[COLLECTOR] Stopped")
}
