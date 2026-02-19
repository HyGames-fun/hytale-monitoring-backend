package postgres

import (
	"database/sql"
	"errors"
	"time"

	"github.com/HyGames-fun/hytale-monitoring-backend/internal/domain"
	"github.com/jmoiron/sqlx"
)

type VerificationRepo struct {
	db *sqlx.DB
}

func NewVerificationRepo(db *sqlx.DB) *VerificationRepo {
	return &VerificationRepo{db: db}
}

func (r *VerificationRepo) CreateVerification(v *domain.ServerVerification) error {
	v.CreatedAt = time.Now()
	v.Status = domain.StatusPending

	query := `
	INSERT INTO "server_verifications" (server_id, user_id, challenge_type, challenge_value, status, created_at)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id;
	`
	return r.db.QueryRowx(query, v.ServerID, v.UserID, v.ChallengeType, v.ChallengeValue, v.Status, v.CreatedAt).Scan(&v.ID)
}

func (r *VerificationRepo) GetPendingVerification(serverID int, challengeType domain.ChallengeType) (*domain.ServerVerification, error) {
	var v domain.ServerVerification
	query := `
	SELECT id, server_id, user_id, challenge_type, challenge_value, status, created_at, verified_at
	FROM "server_verifications"
	WHERE server_id = $1 AND challenge_type = $2 AND status = 'PENDING'
	ORDER BY created_at DESC
	LIMIT 1
	`
	err := r.db.Get(&v, query, serverID, challengeType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *VerificationRepo) UpdateVerificationStatus(id int, status domain.VerificationStatus) error {
	now := time.Now()
	query := `UPDATE "server_verifications" SET status = $1, verified_at = $2 WHERE id = $3`
	_, err := r.db.Exec(query, status, now, id)
	return err
}
