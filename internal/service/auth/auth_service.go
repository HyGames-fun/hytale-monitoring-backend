package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/HyGames-fun/hytale-monitoring-backend/internal/config"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/domain"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/storage/postgres"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/storage/redis"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	cfg         *config.Config
	userRepo    *postgres.UserRepo
	redisClient *redis.Client
	mailService *MailService
}

func NewAuthService(
	cfg *config.Config,
	userRepo *postgres.UserRepo,
	redisClient *redis.Client,
	mailService *MailService,
) *AuthService {
	return &AuthService{
		cfg:         cfg,
		userRepo:    userRepo,
		redisClient: redisClient,
		mailService: mailService,
	}
}

type AuthResult struct {
	AccessToken string `json:"accessToken"`
}

func (s *AuthService) SignTokens(userID int) (accessToken, refreshToken string, err error) {
	accessClaims := jwt.MapClaims{
		"id":  userID,
		"exp": time.Now().Add(s.cfg.JWTAccessTokenTTL).Unix(),
		"iat": time.Now().Unix(),
	}
	accessTok := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = accessTok.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", "", err
	}

	refreshClaims := jwt.MapClaims{
		"id":  userID,
		"exp": time.Now().Add(s.cfg.JWTRefreshTokenTTL).Unix(),
		"iat": time.Now().Unix(),
	}
	refreshTok := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshTok.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) SetCookie(w http.ResponseWriter, name, value string, expires time.Time) {
	sameSite := http.SameSiteLaxMode
	secure := !s.cfg.IsDev
	if secure {
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Domain:   s.cfg.CookieDomain,
		Expires:  expires,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})
}

func (s *AuthService) ClearCookie(w http.ResponseWriter, name string) {
	s.SetCookie(w, name, "", time.Unix(0, 0))
}

func (s *AuthService) SetAuthResponse(w http.ResponseWriter, userID int) (*AuthResult, error) {
	access, refresh, err := s.SignTokens(userID)
	if err != nil {
		return nil, err
	}

	s.SetCookie(w, "refreshToken", refresh, time.Now().Add(s.cfg.JWTRefreshTokenTTL))
	return &AuthResult{AccessToken: access}, nil
}

func (s *AuthService) InitRegister(w http.ResponseWriter, dto *domain.RegisterDTO) error {
	existing, err := s.userRepo.FindUserByEmail(dto.Email)
	if err != nil {
		return err
	}
	if existing != nil {
		return fmt.Errorf("user with email %s already exists", dto.Email)
	}

	regClaims := jwt.MapClaims{
		"name":     dto.Name,
		"email":    dto.Email,
		"password": dto.Password,
		"exp":      time.Now().Add(s.cfg.RegisterTokenTTL).Unix(),
		"iat":      time.Now().Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, regClaims)
	regToken, err := tok.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return err
	}

	s.SetCookie(w, "registerToken", regToken, time.Now().Add(s.cfg.RegisterTokenTTL))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = s.mailService.SendVerificationCode(ctx, dto.Email)
	return err
}

func (s *AuthService) VerifyRegisterToken(r *http.Request) (*domain.RegisterDTO, error) {
	cookie, err := r.Cookie("registerToken")
	if err != nil || cookie.Value == "" {
		return nil, errors.New("register token not found")
	}

	token, err := jwt.Parse(cookie.Value, func(t *jwt.Token) (any, error) {
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid register token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid register token")
	}

	name, _ := claims["name"].(string)
	email, _ := claims["email"].(string)
	password, _ := claims["password"].(string)

	return &domain.RegisterDTO{
		Name:     name,
		Email:    email,
		Password: password,
	}, nil
}

func (s *AuthService) Register(w http.ResponseWriter, r *http.Request, code int) (*AuthResult, error) {
	dto, err := s.VerifyRegisterToken(r)
	if err != nil {
		s.ClearCookie(w, "registerToken")
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cachedCode, _, err := s.redisClient.GetEmailCode(ctx, dto.Email)
	if err != nil || cachedCode == nil {
		return nil, errors.New("code not found")
	}

	if cachedCode.Code != strconv.Itoa(code) {
		return nil, errors.New("invalid code")
	}

	existing, err := s.userRepo.FindUserByEmail(dto.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("user with email %s already exists", dto.Email)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	pwStr := string(hashedPassword)
	emailLower := strings.ToLower(dto.Email)
	newUser := &domain.User{
		Name:     dto.Name,
		Email:    &emailLower,
		Password: &pwStr,
	}

	if err := s.userRepo.CreateUser(newUser); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.ClearCookie(w, "registerToken")
	_ = s.redisClient.DelEmailCode(ctx, dto.Email)

	return s.SetAuthResponse(w, newUser.ID)
}

func (s *AuthService) ResendEmail(w http.ResponseWriter, r *http.Request) error {
	dto, err := s.VerifyRegisterToken(r)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = s.mailService.SendVerificationCode(ctx, dto.Email)
	return err
}

func (s *AuthService) StopRegisterSession(w http.ResponseWriter) {
	s.ClearCookie(w, "registerToken")
}

func (s *AuthService) Login(w http.ResponseWriter, dto *domain.LoginDTO) (*AuthResult, error) {
	user, err := s.userRepo.FindUserByEmail(dto.Email)
	if err != nil || user == nil || user.Password == nil {
		return nil, errors.New("user not found!")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(dto.Password)); err != nil {
		return nil, errors.New("user not found!")
	}

	return s.SetAuthResponse(w, user.ID)
}

func (s *AuthService) Refresh(w http.ResponseWriter, r *http.Request) (*AuthResult, error) {
	cookie, err := r.Cookie("refreshToken")
	if err != nil || cookie.Value == "" {
		return nil, errors.New("refresh token must be updated!")
	}

	token, err := jwt.Parse(cookie.Value, func(t *jwt.Token) (any, error) {
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid refresh token")
	}

	idFloat, ok := claims["id"].(float64)
	if !ok {
		return nil, errors.New("invalid refresh token")
	}

	userID := int(idFloat)
	user, err := s.userRepo.FindUserByID(userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found!")
	}

	return s.SetAuthResponse(w, user.ID)
}

func (s *AuthService) Logout(w http.ResponseWriter) {
	s.ClearCookie(w, "refreshToken")
}
