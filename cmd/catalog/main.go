package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HyGames-fun/hytale-monitoring-backend/internal/config"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/handler"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/middleware"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/service/catalog"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/storage/postgres"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/storage/redis"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/turnstile"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.Load()
	log.Printf("[CATALOG] Starting Catalog API on port %s", cfg.Port)

	db, err := postgres.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[CATALOG] Database connection failed: %v", err)
	}
	defer db.Close()

	redisClient, err := redis.Connect(cfg.RedisURL)
	if err != nil {
		log.Fatalf("[CATALOG] Redis connection failed: %v", err)
	}
	defer redisClient.Close()

	userRepo := postgres.NewUserRepo(db)
	serverRepo := postgres.NewServerRepo(db)
	metricsRepo := postgres.NewMetricsRepo(db)
	verificationRepo := postgres.NewVerificationRepo(db)

	turnstileVerifier := turnstile.NewVerifier(cfg.TurnstileSecretKey, cfg.IsDev)

	serverService := catalog.NewServerService(serverRepo, metricsRepo, redisClient, turnstileVerifier)
	verificationService := catalog.NewVerificationService(serverRepo, verificationRepo)

	catalogHandler := handler.NewCatalogHandler(serverService, verificationService)

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.CORS(cfg.CorsOrigin))
	r.Use(middleware.Throttle(redisClient, 100, 1*time.Minute))
	r.Use(middleware.OptionalJWTAuth(cfg.JWTSecret, userRepo))

	r.Mount("/", catalogHandler.Routes())

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[CATALOG] Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("[CATALOG] Shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Println("[CATALOG] Stopped")
}
