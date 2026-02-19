package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/HyGames-fun/hytale-monitoring-backend/internal/config"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/domain"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/storage/postgres"
)

type DiscordService struct {
	cfg        *config.Config
	userRepo   *postgres.UserRepo
	httpClient *http.Client
}

func NewDiscordService(cfg *config.Config, userRepo *postgres.UserRepo) *DiscordService {
	return &DiscordService{
		cfg:      cfg,
		userRepo: userRepo,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *DiscordService) GetAuthURL() string {
	clientID := s.cfg.DiscordClientID
	if clientID == "" {
		clientID = "1475225594282381513"
	}
	redirectURI := url.QueryEscape(s.cfg.DiscordRedirectURI)
	return fmt.Sprintf("https://discord.com/oauth2/authorize?client_id=%s&response_type=code&redirect_uri=%s&scope=identify", clientID, redirectURI)
}

type discordTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type discordUserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

func (s *DiscordService) HandleCallback(ctx context.Context, code string) (*domain.User, string, time.Duration, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", s.cfg.DiscordRedirectURI)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://discord.com/api/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, "", 0, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(s.cfg.DiscordClientID, s.cfg.DiscordClientSecret)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, "", 0, errors.New("discord oauth request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", 0, fmt.Errorf("discord oauth failed with code %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp discordTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, "", 0, err
	}

	userReq, err := http.NewRequestWithContext(ctx, "GET", "https://discord.com/api/users/@me", nil)
	if err != nil {
		return nil, "", 0, err
	}
	userReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)

	userResp, err := s.httpClient.Do(userReq)
	if err != nil {
		return nil, "", 0, errors.New("failed to fetch discord user")
	}
	defer userResp.Body.Close()

	var dsUser discordUserResponse
	if err := json.NewDecoder(userResp.Body).Decode(&dsUser); err != nil {
		return nil, "", 0, err
	}

	existingUser, err := s.userRepo.FindUserByDiscordID(dsUser.ID)
	if err != nil {
		return nil, "", 0, err
	}

	expiresIn := time.Duration(tokenResp.ExpiresIn) * time.Second
	if expiresIn <= 0 {
		expiresIn = 7 * 24 * time.Hour
	}

	if existingUser != nil {
		return existingUser, tokenResp.RefreshToken, expiresIn, nil
	}

	avatarURL := fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", dsUser.ID, dsUser.Avatar)
	newUser := &domain.User{
		Name:      dsUser.Username,
		DiscordID: &dsUser.ID,
		Avatar:    &avatarURL,
	}

	if err := s.userRepo.CreateUser(newUser); err != nil {
		return nil, "", 0, fmt.Errorf("failed to register discord user: %w", err)
	}

	return newUser, tokenResp.RefreshToken, expiresIn, nil
}
