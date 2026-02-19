package postgres

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func Connect(databaseURL string) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	if err := InitSchema(db); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}

func InitSchema(db *sqlx.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS "user" (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		email VARCHAR(255) UNIQUE,
		avatar VARCHAR(1024),
		password VARCHAR(255),
		discord_id VARCHAR(255) UNIQUE,
		google_id VARCHAR(255) UNIQUE,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS "server" (
		id SERIAL PRIMARY KEY,
		domain VARCHAR(255) UNIQUE,
		real_ip VARCHAR(255) UNIQUE,
		description TEXT NOT NULL,
		name VARCHAR(255) NOT NULL,
		name_id VARCHAR(255) UNIQUE NOT NULL,
		tags JSONB NOT NULL DEFAULT '[]'::jsonb,
		region VARCHAR(64) NOT NULL,
		poster VARCHAR(1024),
		likes_quantity INT NOT NULL DEFAULT 0,
		user_id INT UNIQUE REFERENCES "user"(id) ON DELETE SET NULL,
		is_verified BOOLEAN NOT NULL DEFAULT FALSE,
		verification_token VARCHAR(255),
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS "like" (
		user_id INT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
		server_id INT NOT NULL REFERENCES "server"(id) ON DELETE CASCADE,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		PRIMARY KEY (user_id, server_id)
	);

	CREATE TABLE IF NOT EXISTS "server_metrics" (
		id BIGSERIAL PRIMARY KEY,
		server_id INT NOT NULL REFERENCES "server"(id) ON DELETE CASCADE,
		is_online BOOLEAN NOT NULL,
		players_online INT NOT NULL DEFAULT 0,
		players_max INT NOT NULL DEFAULT 0,
		tps DOUBLE PRECISION NOT NULL DEFAULT 20.0,
		ram_used_mb INT NOT NULL DEFAULT 0,
		latency_ms INT NOT NULL DEFAULT 0,
		recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS "server_verifications" (
		id SERIAL PRIMARY KEY,
		server_id INT NOT NULL REFERENCES "server"(id) ON DELETE CASCADE,
		user_id INT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
		challenge_type VARCHAR(64) NOT NULL,
		challenge_value VARCHAR(255) NOT NULL,
		status VARCHAR(64) NOT NULL DEFAULT 'PENDING',
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		verified_at TIMESTAMPTZ
	);

	CREATE INDEX IF NOT EXISTS idx_server_name_id ON "server"(name_id);
	CREATE INDEX IF NOT EXISTS idx_server_likes_qty ON "server"(likes_quantity DESC);
	CREATE INDEX IF NOT EXISTS idx_server_metrics_server_id ON "server_metrics"(server_id, recorded_at DESC);
	`

	_, err := db.Exec(schema)
	return err
}
