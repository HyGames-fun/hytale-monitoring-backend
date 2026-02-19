package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/HyGames-fun/hytale-monitoring-backend/internal/domain"
	goredis "github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *goredis.Client
}

func Connect(redisURL string) (*Client, error) {
	opt, err := goredis.ParseURL(redisURL)
	if err != nil {
		opt = &goredis.Options{
			Addr: "localhost:6379",
		}
	}

	rdb := goredis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return &Client{rdb: rdb}, nil
	}

	return &Client{rdb: rdb}, nil
}

func (c *Client) Close() error {
	return c.rdb.Close()
}

func (c *Client) GetGuestLike(ctx context.Context, ip string, serverID int) (bool, error) {
	key := fmt.Sprintf("guest_like:%s:%d", ip, serverID)
	val, err := c.rdb.Get(ctx, key).Result()
	if errors.Is(err, goredis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return val == "true" || val == "1", nil
}

func (c *Client) SetGuestLike(ctx context.Context, ip string, serverID int, ttl time.Duration) error {
	key := fmt.Sprintf("guest_like:%s:%d", ip, serverID)
	return c.rdb.Set(ctx, key, "true", ttl).Err()
}

func (c *Client) DelGuestLike(ctx context.Context, ip string, serverID int) error {
	key := fmt.Sprintf("guest_like:%s:%d", ip, serverID)
	return c.rdb.Del(ctx, key).Err()
}

type EmailCodeData struct {
	Code  string `json:"code"`
	Count int    `json:"count"`
}

func (c *Client) GetEmailCode(ctx context.Context, email string) (*EmailCodeData, time.Duration, error) {
	key := fmt.Sprintf("email_code:%s", email)
	val, err := c.rdb.Get(ctx, key).Result()
	if errors.Is(err, goredis.Nil) {
		return nil, 0, nil
	}
	if err != nil {
		return nil, 0, err
	}

	ttl, _ := c.rdb.TTL(ctx, key).Result()

	var data EmailCodeData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, 0, err
	}

	return &data, ttl, nil
}

func (c *Client) SaveEmailCode(ctx context.Context, email string, code string, count int, ttl time.Duration) error {
	key := fmt.Sprintf("email_code:%s", email)
	data := EmailCodeData{
		Code:  code,
		Count: count,
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, key, string(bytes), ttl).Err()
}

func (c *Client) DelEmailCode(ctx context.Context, email string) error {
	key := fmt.Sprintf("email_code:%s", email)
	return c.rdb.Del(ctx, key).Err()
}

func (c *Client) SetLeaderboard(ctx context.Context, items []domain.LeaderboardItem, ttl time.Duration) error {
	data, err := json.Marshal(items)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, "server:leaderboard", data, ttl).Err()
}

func (c *Client) GetLeaderboard(ctx context.Context) ([]domain.LeaderboardItem, error) {
	val, err := c.rdb.Get(ctx, "server:leaderboard").Result()
	if errors.Is(err, goredis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var items []domain.LeaderboardItem
	if err := json.Unmarshal([]byte(val), &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (c *Client) CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, int, error) {
	rkey := fmt.Sprintf("ratelimit:%s", key)
	count, err := c.rdb.Incr(ctx, rkey).Result()
	if err != nil {
		return true, limit, nil
	}

	if count == 1 {
		c.rdb.Expire(ctx, rkey, window)
	}

	if count > int64(limit) {
		return false, 0, nil
	}

	remaining := limit - int(count)
	return true, remaining, nil
}
