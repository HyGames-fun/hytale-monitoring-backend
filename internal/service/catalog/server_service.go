package catalog

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/HyGames-fun/hytale-monitoring-backend/internal/domain"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/storage/postgres"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/storage/redis"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/turnstile"
)

type ServerService struct {
	serverRepo  *postgres.ServerRepo
	metricsRepo *postgres.MetricsRepo
	redisClient *redis.Client
	turnstile   *turnstile.Verifier
}

func NewServerService(
	serverRepo *postgres.ServerRepo,
	metricsRepo *postgres.MetricsRepo,
	redisClient *redis.Client,
	turnstile *turnstile.Verifier,
) *ServerService {
	return &ServerService{
		serverRepo:  serverRepo,
		metricsRepo: metricsRepo,
		redisClient: redisClient,
		turnstile:   turnstile,
	}
}

func (s *ServerService) Create(dto *domain.CreateServerDTO) (*domain.Server, error) {
	server := &domain.Server{
		Domain:        dto.Domain,
		RealIP:        dto.IP,
		Description:   dto.Description,
		Name:          dto.Name,
		NameID:        dto.NameID,
		Tags:          dto.Tags,
		Region:        dto.Region,
		Poster:        dto.Poster,
		LikesQuantity: 0,
	}

	if err := s.serverRepo.CreateServer(server); err != nil {
		return nil, err
	}
	return server, nil
}

func (s *ServerService) Like(serverID int, user *domain.User, token string, ip string) (bool, error) {
	server, err := s.serverRepo.GetServerByID(serverID)
	if err != nil {
		return false, err
	}
	if server == nil {
		return false, errors.New("server not found")
	}

	if user == nil {
		if token == "" {
			return false, errors.New("invalid turnstile token!")
		}
		if !s.turnstile.Verify(ip, token) {
			return false, errors.New("invalid turnstile token!")
		}
		return s.toggleGuestLike(serverID, ip)
	}

	return s.serverRepo.ToggleUserLike(user.ID, serverID)
}

func (s *ServerService) toggleGuestLike(serverID int, ip string) (bool, error) {
	if ip == "" {
		return false, errors.New("guest IP not found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	isLiked, err := s.redisClient.GetGuestLike(ctx, ip, serverID)
	if err != nil {
		return false, err
	}

	if !isLiked {
		if err := s.redisClient.SetGuestLike(ctx, ip, serverID, 24*time.Hour); err != nil {
			return false, err
		}
		_ = s.serverRepo.UpdateLikesQuantity(serverID, 1)
		return true, nil
	} else {
		if err := s.redisClient.DelGuestLike(ctx, ip, serverID); err != nil {
			return false, err
		}
		_ = s.serverRepo.UpdateLikesQuantity(serverID, -1)
		return false, nil
	}
}

func (s *ServerService) FindPage(dto *domain.FindPageDTO, user *domain.User, ip string) ([]domain.ServerResponse, error) {
	servers, err := s.serverRepo.FindPage(dto)
	if err != nil {
		return nil, err
	}

	var ids []int
	for _, srv := range servers {
		ids = append(ids, srv.ID)
	}

	latestMetrics, _ := s.metricsRepo.GetLatestMetricsForServers(ids)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	res := make([]domain.ServerResponse, 0, len(servers))
	for _, srv := range servers {
		var liked bool
		if user != nil {
			liked, _ = s.serverRepo.GetUserLike(user.ID, srv.ID)
		} else if ip != "" {
			liked, _ = s.redisClient.GetGuestLike(ctx, ip, srv.ID)
		}

		players := 10
		if m, ok := latestMetrics[srv.ID]; ok && m != nil {
			players = m.PlayersOnline
		}

		res = append(res, domain.ServerResponse{
			ID:            srv.ID,
			Domain:        srv.Domain,
			IP:            srv.RealIP,
			Description:   srv.Description,
			Name:          srv.Name,
			NameID:        srv.NameID,
			Tags:          srv.Tags,
			Poster:        srv.Poster,
			LikesQuantity: srv.LikesQuantity,
			Liked:         liked,
			Players:       players,
			IsVerified:    srv.IsVerified,
		})
	}

	return res, nil
}

func (s *ServerService) FindOne(nameID string, user *domain.User, ip string) (*domain.ServerResponse, error) {
	server, err := s.serverRepo.GetServerByNameID(nameID)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, errors.New("server not found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var liked bool
	if user != nil {
		liked, _ = s.serverRepo.GetUserLike(user.ID, server.ID)
	} else if ip != "" {
		liked, _ = s.redisClient.GetGuestLike(ctx, ip, server.ID)
	}

	players := 10
	if m, err := s.metricsRepo.GetLatestMetric(server.ID); err == nil && m != nil {
		players = m.PlayersOnline
	}

	return &domain.ServerResponse{
		ID:            server.ID,
		Domain:        server.Domain,
		IP:            server.RealIP,
		Description:   server.Description,
		Name:          server.Name,
		NameID:        server.NameID,
		Tags:          server.Tags,
		Poster:        server.Poster,
		LikesQuantity: server.LikesQuantity,
		Liked:         liked,
		Players:       players,
		IsVerified:    server.IsVerified,
	}, nil
}

func (s *ServerService) FindQuantity() (int, error) {
	return s.serverRepo.CountServers()
}

func (s *ServerService) GetOnlineStatus(serverID int) (*domain.ServerStatusResponse, error) {
	server, err := s.serverRepo.GetServerByID(serverID)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, errors.New("server not found")
	}

	var isOnline *bool
	target := ""
	if server.RealIP != nil && *server.RealIP != "" {
		target = *server.RealIP
	} else if server.Domain != nil && *server.Domain != "" {
		target = *server.Domain
	}

	if target != "" {
		online := CheckPort(target, 3*time.Second)
		isOnline = &online
	}

	return &domain.ServerStatusResponse{
		ID:       server.ID,
		IsOnline: isOnline,
	}, nil
}

func (s *ServerService) GetOnlineStatuses(ids []int) ([]domain.ServerStatusResponse, error) {
	if len(ids) == 0 {
		return nil, errors.New("servers not found")
	}

	var res []domain.ServerStatusResponse
	for _, id := range ids {
		status, err := s.GetOnlineStatus(id)
		if err != nil {
			continue
		}
		res = append(res, *status)
	}

	if len(res) == 0 {
		return nil, errors.New("servers not found")
	}

	return res, nil
}

func (s *ServerService) Search(query string, page, limit int) ([]domain.Server, error) {
	if strings.TrimSpace(query) == "" {
		return nil, errors.New("query is empty")
	}
	if limit < 1 {
		return nil, errors.New("limit is less than 1")
	}
	if limit > 50 {
		return nil, errors.New("limit is more than 50")
	}
	if page < 0 {
		return nil, errors.New("page is less than 0")
	}

	return s.serverRepo.SearchServers(query, page, limit)
}

func (s *ServerService) Update(id int, dto *domain.UpdateServerDTO) (*domain.Server, error) {
	server, err := s.serverRepo.GetServerByID(id)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, errors.New("server not found")
	}

	if dto.Domain != nil {
		server.Domain = dto.Domain
	}
	if dto.IP != nil {
		server.RealIP = dto.IP
	}
	if dto.Name != nil {
		server.Name = *dto.Name
	}
	if dto.NameID != nil {
		server.NameID = *dto.NameID
	}
	if dto.Description != nil {
		server.Description = *dto.Description
	}
	if dto.Poster != nil {
		server.Poster = dto.Poster
	}
	if dto.Tags != nil {
		server.Tags = *dto.Tags
	}
	if dto.Region != nil {
		server.Region = *dto.Region
	}

	if err := s.serverRepo.UpdateServer(server); err != nil {
		return nil, err
	}
	return server, nil
}

func (s *ServerService) Remove(id int) (*domain.Server, error) {
	server, err := s.serverRepo.GetServerByID(id)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, errors.New("server not found")
	}

	if err := s.serverRepo.DeleteServer(id); err != nil {
		return nil, err
	}
	return server, nil
}

func (s *ServerService) GetLeaderboard() ([]domain.LeaderboardItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if cached, err := s.redisClient.GetLeaderboard(ctx); err == nil && len(cached) > 0 {
		return cached, nil
	}

	servers, err := s.serverRepo.FindPage(&domain.FindPageDTO{Page: 0, Quantity: 50})
	if err != nil {
		return nil, err
	}

	items := make([]domain.LeaderboardItem, 0, len(servers))
	for idx, srv := range servers {
		items = append(items, domain.LeaderboardItem{
			ServerID:      srv.ID,
			NameID:        srv.NameID,
			Name:          srv.Name,
			LikesQuantity: srv.LikesQuantity,
			PlayersOnline: 10,
			Rank:          idx + 1,
			IsOnline:      true,
			Score:         float64(srv.LikesQuantity*10 + 10),
		})
	}

	_ = s.redisClient.SetLeaderboard(ctx, items, 60*time.Second)
	return items, nil
}

func CheckPort(value string, timeout time.Duration) bool {
	if !strings.Contains(value, ":") {
		value = value + ":25565"
	}
	conn, err := net.DialTimeout("tcp", value, timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}
