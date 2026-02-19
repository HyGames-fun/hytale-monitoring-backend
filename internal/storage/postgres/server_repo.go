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

type ServerRepo struct {
	db *sqlx.DB
}

func NewServerRepo(db *sqlx.DB) *ServerRepo {
	return &ServerRepo{db: db}
}

func (r *ServerRepo) CreateServer(s *domain.Server) error {
	now := time.Now()
	s.CreatedAt = now
	s.UpdatedAt = now

	query := `
	INSERT INTO "server" (domain, real_ip, description, name, name_id, tags, region, poster, likes_quantity, user_id, is_verified, verification_token, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	RETURNING id;
	`
	err := r.db.QueryRowx(
		query,
		s.Domain,
		s.RealIP,
		s.Description,
		s.Name,
		s.NameID,
		s.Tags,
		s.Region,
		s.Poster,
		s.LikesQuantity,
		s.UserID,
		s.IsVerified,
		s.VerificationToken,
		s.CreatedAt,
		s.UpdatedAt,
	).Scan(&s.ID)

	if err != nil {
		return err
	}
	return nil
}

func (r *ServerRepo) GetServerByID(id int) (*domain.Server, error) {
	var s domain.Server
	query := `SELECT id, domain, real_ip, description, name, name_id, tags, region, poster, likes_quantity, user_id, is_verified, verification_token, created_at, updated_at FROM "server" WHERE id = $1`
	err := r.db.Get(&s, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get server by id: %w", err)
	}
	return &s, nil
}

func (r *ServerRepo) GetServerByNameID(nameID string) (*domain.Server, error) {
	var s domain.Server
	query := `SELECT id, domain, real_ip, description, name, name_id, tags, region, poster, likes_quantity, user_id, is_verified, verification_token, created_at, updated_at FROM "server" WHERE name_id = $1`
	err := r.db.Get(&s, query, nameID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get server by nameId: %w", err)
	}
	return &s, nil
}

func (r *ServerRepo) UpdateServer(s *domain.Server) error {
	s.UpdatedAt = time.Now()
	query := `
	UPDATE "server"
	SET domain = $1, real_ip = $2, description = $3, name = $4, name_id = $5, tags = $6, region = $7, poster = $8, updated_at = $9
	WHERE id = $10
	`
	res, err := r.db.Exec(query, s.Domain, s.RealIP, s.Description, s.Name, s.NameID, s.Tags, s.Region, s.Poster, s.UpdatedAt, s.ID)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *ServerRepo) DeleteServer(id int) error {
	query := `DELETE FROM "server" WHERE id = $1`
	res, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *ServerRepo) CountServers() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM "server"`
	err := r.db.Get(&count, query)
	return count, err
}

func (r *ServerRepo) FindPage(dto *domain.FindPageDTO) ([]domain.Server, error) {
	var conditions []string
	var args []any
	argIdx := 1

	if dto.Filters != nil {
		if dto.Filters.Region != nil && *dto.Filters.Region != "" {
			conditions = append(conditions, fmt.Sprintf("region = $%d", argIdx))
			args = append(args, *dto.Filters.Region)
			argIdx++
		}
		if len(dto.Filters.Tags) > 0 {
			var tagChecks []string
			for _, tag := range dto.Filters.Tags {
				tagChecks = append(tagChecks, fmt.Sprintf("tags @> $%d::jsonb", argIdx))
				args = append(args, fmt.Sprintf("[\"%s\"]", tag))
				argIdx++
			}
			conditions = append(conditions, "("+strings.Join(tagChecks, " OR ")+")")
		}
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	orderClause := "ORDER BY likes_quantity DESC, created_at DESC"
	if dto.Order != nil {
		var orders []string
		if dto.Order.Likes != nil {
			dir := "DESC"
			if strings.ToLower(*dto.Order.Likes) == "asc" {
				dir = "ASC"
			}
			orders = append(orders, fmt.Sprintf("likes_quantity %s", dir))
		}
		if dto.Order.CreatedAt != nil {
			dir := "DESC"
			if strings.ToLower(*dto.Order.CreatedAt) == "asc" {
				dir = "ASC"
			}
			orders = append(orders, fmt.Sprintf("created_at %s", dir))
		}
		if len(orders) > 0 {
			orderClause = "ORDER BY " + strings.Join(orders, ", ")
		}
	}

	limit := dto.Quantity
	if limit <= 0 {
		limit = 10
	}
	offset := dto.Page * limit

	query := fmt.Sprintf(`
		SELECT id, domain, real_ip, description, name, name_id, tags, region, poster, likes_quantity, user_id, is_verified, verification_token, created_at, updated_at
		FROM "server"
		%s
		%s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderClause, argIdx, argIdx+1)

	args = append(args, limit, offset)

	var servers []domain.Server
	err := r.db.Select(&servers, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find page: %w", err)
	}
	return servers, nil
}

func (r *ServerRepo) GetUserLike(userID, serverID int) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM "like" WHERE user_id = $1 AND server_id = $2`
	err := r.db.Get(&count, query, userID, serverID)
	return count > 0, err
}

func (r *ServerRepo) ToggleUserLike(userID, serverID int) (bool, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()

	var count int
	err = tx.Get(&count, `SELECT COUNT(*) FROM "like" WHERE user_id = $1 AND server_id = $2`, userID, serverID)
	if err != nil {
		return false, err
	}

	isLikedNow := false
	if count > 0 {
		_, err = tx.Exec(`DELETE FROM "like" WHERE user_id = $1 AND server_id = $2`, userID, serverID)
		if err != nil {
			return false, err
		}
		_, err = tx.Exec(`UPDATE "server" SET likes_quantity = GREATEST(likes_quantity - 1, 0) WHERE id = $1`, serverID)
		if err != nil {
			return false, err
		}
		isLikedNow = false
	} else {
		_, err = tx.Exec(`INSERT INTO "like" (user_id, server_id, created_at) VALUES ($1, $2, NOW())`, userID, serverID)
		if err != nil {
			return false, err
		}
		_, err = tx.Exec(`UPDATE "server" SET likes_quantity = likes_quantity + 1 WHERE id = $1`, serverID)
		if err != nil {
			return false, err
		}
		isLikedNow = true
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}
	return isLikedNow, nil
}

func (r *ServerRepo) UpdateLikesQuantity(serverID int, delta int) error {
	var query string
	if delta >= 0 {
		query = `UPDATE "server" SET likes_quantity = likes_quantity + $1 WHERE id = $2`
	} else {
		query = `UPDATE "server" SET likes_quantity = GREATEST(likes_quantity + $1, 0) WHERE id = $2`
	}
	_, err := r.db.Exec(query, delta, serverID)
	return err
}

func (r *ServerRepo) SearchServers(q string, page, limit int) ([]domain.Server, error) {
	words := strings.Fields(strings.ToLower(q))
	if len(words) == 0 {
		return []domain.Server{}, nil
	}

	var conditions []string
	var args []any
	argIdx := 1

	for _, word := range words {
		pattern := "%" + word + "%"
		tagCandidate := strings.ToUpper(word)
		isTag := domain.ValidTags[domain.Tag(tagCandidate)]

		if isTag {
			conditions = append(conditions, fmt.Sprintf("(LOWER(name) LIKE $%d OR LOWER(description) LIKE $%d OR tags @> $%d::jsonb)", argIdx, argIdx+1, argIdx+2))
			args = append(args, pattern, pattern, fmt.Sprintf("[\"%s\"]", tagCandidate))
			argIdx += 3
		} else {
			conditions = append(conditions, fmt.Sprintf("(LOWER(name) LIKE $%d OR LOWER(description) LIKE $%d)", argIdx, argIdx+1))
			args = append(args, pattern, pattern)
			argIdx += 2
		}
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")
	offset := page * limit

	query := fmt.Sprintf(`
		SELECT id, domain, real_ip, description, name, name_id, tags, region, poster, likes_quantity, user_id, is_verified, verification_token, created_at, updated_at
		FROM "server"
		%s
		ORDER BY likes_quantity DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	args = append(args, limit, offset)

	var servers []domain.Server
	err := r.db.Select(&servers, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search servers: %w", err)
	}
	return servers, nil
}

func (r *ServerRepo) GetAllActiveServers() ([]domain.Server, error) {
	var servers []domain.Server
	query := `SELECT id, domain, real_ip, description, name, name_id, tags, region, poster, likes_quantity, user_id, is_verified, verification_token, created_at, updated_at FROM "server"`
	err := r.db.Select(&servers, query)
	return servers, err
}

func (r *ServerRepo) SetVerified(serverID int, userID int) error {
	query := `UPDATE "server" SET is_verified = TRUE, user_id = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(query, userID, serverID)
	return err
}
