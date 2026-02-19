package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/HyGames-fun/hytale-monitoring-backend/internal/response"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/storage/redis"
)

func Throttle(redisClient *redis.Client, limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if redisClient == nil {
				next.ServeHTTP(w, r)
				return
			}

			ip := GetClientIP(r)
			key := "req:" + ip + ":" + r.URL.Path

			ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
			defer cancel()

			allowed, _, err := redisClient.CheckRateLimit(ctx, key, limit, window)
			if err == nil && !allowed {
				response.Error(w, http.StatusTooManyRequests, "ThrottlerException: Too Many Requests")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func GetClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return strings.TrimSpace(realIP)
	}
	remoteAddr := r.RemoteAddr
	if idx := strings.LastIndex(remoteAddr, ":"); idx != -1 {
		return remoteAddr[:idx]
	}
	return remoteAddr
}
