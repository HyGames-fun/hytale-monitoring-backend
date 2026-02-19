package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/smtp"

	"github.com/HyGames-fun/hytale-monitoring-backend/internal/config"
	"github.com/HyGames-fun/hytale-monitoring-backend/internal/storage/redis"
)

type MailService struct {
	cfg         *config.Config
	redisClient *redis.Client
}

func NewMailService(cfg *config.Config, redisClient *redis.Client) *MailService {
	return &MailService{
		cfg:         cfg,
		redisClient: redisClient,
	}
}

func (s *MailService) CheckResendCooldown(ctx context.Context, email string) error {
	exist, ttl, err := s.redisClient.GetEmailCode(ctx, email)
	if err != nil || exist == nil {
		return nil
	}

	if exist.Count >= 3 {
		return errors.New("maximum verification attempts reached, please wait before requesting again")
	}

	resendAfter := s.cfg.SMTPVerificationTTL - s.cfg.SMTPVerificationResendDelay
	if exist.Count > 0 && ttl > resendAfter {
		return errors.New("can not resend: please wait before requesting another code")
	}

	return nil
}

func (s *MailService) SendVerificationCode(ctx context.Context, email string) (string, error) {
	if err := s.CheckResendCooldown(ctx, email); err != nil {
		return "", err
	}

	code := s.generateCode()

	exist, _, _ := s.redisClient.GetEmailCode(ctx, email)
	count := 1
	if exist != nil {
		count = exist.Count + 1
	}

	if err := s.redisClient.SaveEmailCode(ctx, email, code, count, s.cfg.SMTPVerificationTTL); err != nil {
		return "", fmt.Errorf("failed to save code in cache: %w", err)
	}

	if s.cfg.SMTPHost != "" && s.cfg.SMTPUser != "" {
		go func() {
			if err := s.sendEmail(email, code); err != nil {
				log.Printf("[MAIL] Failed to send email to %s: %v", email, err)
			}
		}()
	} else {
		log.Printf("[MAIL-MOCK] Verification code for %s is: %s (attempt %d)", email, code, count)
	}

	return code, nil
}

func (s *MailService) generateCode() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(900000))
	return fmt.Sprintf("%06d", n.Int64()+100000)
}

func (s *MailService) sendEmail(toEmail, code string) error {
	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPassword, s.cfg.SMTPHost)
	to := []string{toEmail}
	msg := []byte(fmt.Sprintf(
		"From: %s <%s>\r\n"+
			"To: %s\r\n"+
			"Subject: Verification code\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n\r\n"+
			"<b>%s</b>\r\n",
		s.cfg.SMTPName, s.cfg.SMTPEmail, toEmail, code,
	))

	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)
	return smtp.SendMail(addr, auth, s.cfg.SMTPEmail, to, msg)
}
