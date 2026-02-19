package collector

import (
	"context"
	"log"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/HyGames-fun/hytale-monitoring-backend/internal/domain"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/storage/postgres"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/storage/redis"
)

type Poller struct {
	serverRepo  *postgres.ServerRepo
	metricsRepo *postgres.MetricsRepo
	redisClient *redis.Client
	pool        *WorkerPool
	interval    time.Duration
}

func NewPoller(
	serverRepo *postgres.ServerRepo,
	metricsRepo *postgres.MetricsRepo,
	redisClient *redis.Client,
	workerCount int,
	interval time.Duration,
) *Poller {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return &Poller{
		serverRepo:  serverRepo,
		metricsRepo: metricsRepo,
		redisClient: redisClient,
		pool:        NewWorkerPool(workerCount, 1000),
		interval:    interval,
	}
}

func (p *Poller) Start(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	p.pollAll(ctx)

	for {
		select {
		case <-ctx.Done():
			p.pool.Stop()
			return
		case <-ticker.C:
			p.pollAll(ctx)
		}
	}
}

func (p *Poller) pollAll(ctx context.Context) {
	servers, err := p.serverRepo.GetAllActiveServers()
	if err != nil {
		log.Printf("[POLLER] Failed to fetch active servers: %v", err)
		return
	}

	if len(servers) == 0 {
		return
	}

	var wg sync.WaitGroup
	type polledResult struct {
		server  domain.Server
		online  bool
		latency int
	}
	resultsChan := make(chan polledResult, len(servers))

	for _, srv := range servers {
		wg.Add(1)
		targetServer := srv
		p.pool.Submit(func(tCtx context.Context) {
			defer wg.Done()
			online, latency := checkServerReachability(targetServer)

			metric := &domain.ServerMetric{
				ServerID:      targetServer.ID,
				IsOnline:      online,
				PlayersOnline: 10,
				PlayersMax:    100,
				TPS:           20.0,
				RAMUsedMB:     2048,
				LatencyMS:     latency,
				RecordedAt:    time.Now(),
			}

			_ = p.metricsRepo.SaveMetric(metric)

			resultsChan <- polledResult{
				server:  targetServer,
				online:  online,
				latency: latency,
			}
		})
	}

	wg.Wait()
	close(resultsChan)

	var leaderboard []domain.LeaderboardItem
	for res := range resultsChan {
		score := float64(res.server.LikesQuantity*10) + 10
		if !res.online {
			score /= 2
		}

		leaderboard = append(leaderboard, domain.LeaderboardItem{
			ServerID:      res.server.ID,
			NameID:        res.server.NameID,
			Name:          res.server.Name,
			LikesQuantity: res.server.LikesQuantity,
			PlayersOnline: 10,
			IsOnline:      res.online,
			Score:         score,
		})
	}

	sort.Slice(leaderboard, func(i, j int) bool {
		return leaderboard[i].Score > leaderboard[j].Score
	})

	for i := range leaderboard {
		leaderboard[i].Rank = i + 1
	}

	if err := p.redisClient.SetLeaderboard(ctx, leaderboard, p.interval*2); err != nil {
		log.Printf("[POLLER] Failed to cache leaderboard in Redis: %v", err)
	}
}

func checkServerReachability(s domain.Server) (bool, int) {
	target := ""
	if s.RealIP != nil && *s.RealIP != "" {
		target = *s.RealIP
	} else if s.Domain != nil && *s.Domain != "" {
		target = *s.Domain
	}

	if target == "" {
		return false, 0
	}

	if !strings.Contains(target, ":") {
		target += ":25565"
	}

	start := time.Now()
	conn, err := net.DialTimeout("tcp", target, 2*time.Second)
	if err != nil {
		return false, 0
	}
	_ = conn.Close()
	return true, int(time.Since(start).Milliseconds())
}
