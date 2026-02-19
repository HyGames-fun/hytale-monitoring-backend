package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port                        string
	AuthPort                    string
	CollectorPort               string
	GRPCPort                    string
	DatabaseURL                 string
	RedisURL                    string
	JWTSecret                   string
	JWTAccessTokenTTL           time.Duration
	JWTRefreshTokenTTL          time.Duration
	RegisterTokenTTL            time.Duration
	CookieDomain                string
	CorsOrigin                  string
	TurnstileSecretKey          string
	DiscordClientID             string
	DiscordClientSecret         string
	DiscordRedirectURI          string
	GoogleClientID              string
	GoogleClientSecret          string
	GoogleRedirectURI           string
	SMTPHost                    string
	SMTPPort                    int
	SMTPUser                    string
	SMTPPassword                string
	SMTPEmail                   string
	SMTPName                    string
	SMTPVerificationTTL         time.Duration
	SMTPVerificationResendDelay time.Duration
	SearchMethod                string
	IsDev                       bool
}

func Load() *Config {
	env := getEnv("NODE_ENV", getEnv("ENV", "development"))
	isDev := env == "development" || env == "dev"

	return &Config{
		Port:                        getEnv("PORT", "4242"),
		AuthPort:                    getEnv("AUTH_PORT", "4243"),
		CollectorPort:               getEnv("COLLECTOR_PORT", "4244"),
		GRPCPort:                    getEnv("GRPC_PORT", "50051"),
		DatabaseURL:                 getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/hytale?sslmode=disable"),
		RedisURL:                    getEnv("REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:                   getEnv("JWT_SECRET", "super-secret-jwt-key-2026"),
		JWTAccessTokenTTL:           parseDuration(getEnv("JWT_ACCESS_TOKEN_TTL", "15m"), 15*time.Minute),
		JWTRefreshTokenTTL:          parseDuration(getEnv("JWT_REFRESH_TOKEN_TTL", "720h"), 30*24*time.Hour),
		RegisterTokenTTL:            parseDuration(getEnv("REGISTER_TOKEN_TTL", "15m"), 15*time.Minute),
		CookieDomain:                getEnv("COOKIE_DOMAIN", "localhost"),
		CorsOrigin:                  getEnv("CORS_ORIGIN", "*"),
		TurnstileSecretKey:          getEnv("TURNSTILE_SECRET_KEY", ""),
		DiscordClientID:             getEnv("DISCORD_CLIENT_ID", ""),
		DiscordClientSecret:         getEnv("DISCORD_CLIENT_SECRET", ""),
		DiscordRedirectURI:          getEnv("DISCORD_REDIRECT_URI", "http://localhost:4242/auth/discord/callback"),
		GoogleClientID:              getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:          getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURI:           getEnv("GOOGLE_REDIRECT_URI", "http://localhost:4242/auth/google/callback"),
		SMTPHost:                    getEnv("SMTP_HOST", ""),
		SMTPPort:                    parseInt(getEnv("SMTP_PORT", "465"), 465),
		SMTPUser:                    getEnv("SMTP_USER", ""),
		SMTPPassword:                getEnv("SMTP_PASSWORD", ""),
		SMTPEmail:                   getEnv("SMTP_EMAIL", "noreply@hytale-monitoring.fun"),
		SMTPName:                    getEnv("SMTP_NAME", "Hytale Monitoring"),
		SMTPVerificationTTL:         parseDuration(getEnv("SMTP_VERIFICATION_TTL", "15m"), 15*time.Minute),
		SMTPVerificationResendDelay: parseDuration(getEnv("SMTP_VERIFICATION_RESEND_DELAY", "1m"), 1*time.Minute),
		SearchMethod:                getEnv("SEARCH_METHOD", "database"),
		IsDev:                       isDev,
	}
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}

func parseInt(val string, defaultVal int) int {
	if i, err := strconv.Atoi(val); err == nil {
		return i
	}
	return defaultVal
}

func parseDuration(val string, defaultVal time.Duration) time.Duration {
	if d, err := time.ParseDuration(val); err == nil {
		return d
	}
	return defaultVal
}
