package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HyGames-fun/hytale-monitoring-backend/internal/config"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/handler"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/middleware"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/service/agent"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/service/auth"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/service/catalog"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/service/collector"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/storage/postgres"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/storage/redis"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/turnstile"
	"github.com/HyGames-fun/hytale-monitoring-backend/pkg/agentproto"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc"
)

// @title           Hytale Monitoring API
// @version         1.0
// @description     Hytale Server Monitoring, Catalog and Telemetry Service.

// @host      localhost:4242
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token.

func main() {
	cfg := config.Load()
	log.Printf("[ALL] Starting unified Hytale Monitoring services (HTTP port %s, gRPC port %s)", cfg.Port, cfg.GRPCPort)

	db, err := postgres.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[ALL] Database connection failed: %v", err)
	}
	defer db.Close()

	redisClient, err := redis.Connect(cfg.RedisURL)
	if err != nil {
		log.Fatalf("[ALL] Redis connection failed: %v", err)
	}
	defer redisClient.Close()

	userRepo := postgres.NewUserRepo(db)
	serverRepo := postgres.NewServerRepo(db)
	metricsRepo := postgres.NewMetricsRepo(db)
	verificationRepo := postgres.NewVerificationRepo(db)

	turnstileVerifier := turnstile.NewVerifier(cfg.TurnstileSecretKey, cfg.IsDev)

	serverService := catalog.NewServerService(serverRepo, metricsRepo, redisClient, turnstileVerifier)
	verificationService := catalog.NewVerificationService(serverRepo, verificationRepo)
	mailService := auth.NewMailService(cfg, redisClient)
	authService := auth.NewAuthService(cfg, userRepo, redisClient, mailService)
	discordService := auth.NewDiscordService(cfg, userRepo)
	googleService := auth.NewGoogleService(cfg, userRepo)

	catalogHandler := handler.NewCatalogHandler(serverService, verificationService)
	authHandler := handler.NewAuthHandler(authService, discordService, googleService)
	userHandler := handler.NewUserHandler(userRepo)

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.CORS(cfg.CorsOrigin))
	r.Use(middleware.Throttle(redisClient, 100, 1*time.Minute))
	r.Use(middleware.OptionalJWTAuth(cfg.JWTSecret, userRepo))

	r.Mount("/", catalogHandler.Routes())
	r.Mount("/auth", authHandler.Routes())
	r.Mount("/user", userHandler.Routes())

	httpSrv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("[ALL] HTTP Gateway listening on :%s", cfg.Port)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[ALL] HTTP server error: %v", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	poller := collector.NewPoller(serverRepo, metricsRepo, redisClient, 10, 30*time.Second)
	go poller.Start(ctx)

	grpcListener, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("[ALL] Failed to listen on gRPC port %s: %v", cfg.GRPCPort, err)
	}

	grpcServer := grpc.NewServer()
	agentService := agent.NewServer(serverRepo, metricsRepo)
	agentproto.RegisterHytaleMonitoringAgentServer(grpcServer, agentService)

	go func() {
		log.Printf("[ALL] gRPC Telemetry Server listening on :%s", cfg.GRPCPort)
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("[ALL] gRPC server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("[ALL] Shutting down gracefully...")
	cancel()
	grpcServer.GracefulStop()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = httpSrv.Shutdown(shutdownCtx)
	log.Println("[ALL] All services stopped.")
}
