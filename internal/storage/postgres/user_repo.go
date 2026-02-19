package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/HyGames-fun/hytale-monitoring-backend/internal/domain"
	"github.com/jmoiron/sqlx"
)

type UserRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) CreateUser(u *domain.User) error {
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now

	query := `
	INSERT INTO "user" (name, email, avatar, password, discord_id, google_id, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	RETURNING id;
	`
	err := r.db.QueryRowx(query, u.Name, u.Email, u.Avatar, u.Password, u.DiscordID, u.GoogleID, u.CreatedAt, u.UpdatedAt).Scan(&u.ID)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}
	return nil
}

func (r *UserRepo) FindUserByID(id int) (*domain.User, error) {
	var u domain.User
	query := `SELECT id, name, email, avatar, password, discord_id, google_id, created_at, updated_at FROM "user" WHERE id = $1`
	err := r.db.Get(&u, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}
	return &u, nil
}

func (r *UserRepo) FindUserByEmail(email string) (*domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	var u domain.User
	query := `SELECT id, name, email, avatar, password, discord_id, google_id, created_at, updated_at FROM "user" WHERE LOWER(email) = LOWER($1)`
	err := r.db.Get(&u, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}
	return &u, nil
}

func (r *UserRepo) FindUserByDiscordID(discordID string) (*domain.User, error) {
	var u domain.User
	query := `SELECT id, name, email, avatar, password, discord_id, google_id, created_at, updated_at FROM "user" WHERE discord_id = $1`
	err := r.db.Get(&u, query, discordID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user by discord id: %w", err)
	}
	return &u, nil
}

func (r *UserRepo) FindUserByGoogleID(googleID string) (*domain.User, error) {
	var u domain.User
	query := `SELECT id, name, email, avatar, password, discord_id, google_id, created_at, updated_at FROM "user" WHERE google_id = $1`
	err := r.db.Get(&u, query, googleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user by google id: %w", err)
	}
	return &u, nil
}

func (r *UserRepo) SetAvatar(userID int, avatarURL *string) error {
	query := `UPDATE "user" SET avatar = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(query, avatarURL, userID)
	return err
}
