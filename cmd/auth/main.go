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
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/service/auth"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/storage/postgres"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/storage/redis"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.Load()
	log.Printf("[AUTH] Starting Auth Service on port %s", cfg.AuthPort)

	db, err := postgres.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[AUTH] Database connection failed: %v", err)
	}
	defer db.Close()

	redisClient, err := redis.Connect(cfg.RedisURL)
	if err != nil {
		log.Fatalf("[AUTH] Redis connection failed: %v", err)
	}
	defer redisClient.Close()

	userRepo := postgres.NewUserRepo(db)

	mailService := auth.NewMailService(cfg, redisClient)
	authService := auth.NewAuthService(cfg, userRepo, redisClient, mailService)
	discordService := auth.NewDiscordService(cfg, userRepo)
	googleService := auth.NewGoogleService(cfg, userRepo)

	authHandler := handler.NewAuthHandler(authService, discordService, googleService)
	userHandler := handler.NewUserHandler(userRepo)

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.CORS(cfg.CorsOrigin))
	r.Use(middleware.Throttle(redisClient, 100, 1*time.Minute))
	r.Use(middleware.OptionalJWTAuth(cfg.JWTSecret, userRepo))

	r.Mount("/auth", authHandler.Routes())
	r.Mount("/user", userHandler.Routes())

	srv := &http.Server{
		Addr:         ":" + cfg.AuthPort,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[AUTH] Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("[AUTH] Shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Println("[AUTH] Stopped")
}
