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

type GoogleService struct {
	cfg        *config.Config
	userRepo   *postgres.UserRepo
	httpClient *http.Client
}

func NewGoogleService(cfg *config.Config, userRepo *postgres.UserRepo) *GoogleService {
	return &GoogleService{
		cfg:      cfg,
		userRepo: userRepo,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *GoogleService) GetAuthURL() string {
	clientID := s.cfg.GoogleClientID
	redirectURI := url.QueryEscape(s.cfg.GoogleRedirectURI)
	return fmt.Sprintf("https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=openid%%20profile%%20email&access_type=offline", clientID, redirectURI)
}

type googleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type googleUserResponse struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func (s *GoogleService) HandleCallback(ctx context.Context, code string) (*domain.User, string, time.Duration, error) {
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", s.cfg.GoogleClientID)
	data.Set("client_secret", s.cfg.GoogleClientSecret)
	data.Set("redirect_uri", s.cfg.GoogleRedirectURI)
	data.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, "POST", "https://oauth2.googleapis.com/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, "", 0, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, "", 0, errors.New("google oauth token request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", 0, fmt.Errorf("google oauth failed with code %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp googleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, "", 0, err
	}

	userReq, err := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, "", 0, err
	}
	userReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)

	userResp, err := s.httpClient.Do(userReq)
	if err != nil {
		return nil, "", 0, errors.New("failed to fetch google user profile")
	}
	defer userResp.Body.Close()

	var gUser googleUserResponse
	if err := json.NewDecoder(userResp.Body).Decode(&gUser); err != nil {
		return nil, "", 0, err
	}

	existingUser, err := s.userRepo.FindUserByGoogleID(gUser.ID)
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

	if gUser.Email != "" {
		if byEmail, err := s.userRepo.FindUserByEmail(gUser.Email); err == nil && byEmail != nil {
			byEmail.GoogleID = &gUser.ID
			return byEmail, tokenResp.RefreshToken, expiresIn, nil
		}
	}

	name := gUser.Name
	if name == "" {
		name = "User"
	}
	email := gUser.Email
	avatar := gUser.Picture

	newUser := &domain.User{
		Name:     name,
		Email:    &email,
		GoogleID: &gUser.ID,
		Avatar:   &avatar,
	}

	if err := s.userRepo.CreateUser(newUser); err != nil {
		return nil, "", 0, fmt.Errorf("failed to register google user: %w", err)
	}

	return newUser, tokenResp.RefreshToken, expiresIn, nil
}
