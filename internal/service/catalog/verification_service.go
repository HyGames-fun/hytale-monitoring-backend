package catalog

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/HyGames-fun/hytale-monitoring-backend/internal/domain"
	"github.com/google/uuid"
)

type VerificationServerRepo interface {
	GetServerByID(id int) (*domain.Server, error)
	SetVerified(serverID int, userID int) error
}

type VerificationRepo interface {
	CreateVerification(v *domain.ServerVerification) error
	GetPendingVerification(serverID int, challengeType domain.ChallengeType) (*domain.ServerVerification, error)
	UpdateVerificationStatus(id int, status domain.VerificationStatus) error
}

type VerificationService struct {
	serverRepo       VerificationServerRepo
	verificationRepo VerificationRepo
}

func NewVerificationService(serverRepo VerificationServerRepo, verificationRepo VerificationRepo) *VerificationService {
	return &VerificationService{
		serverRepo:       serverRepo,
		verificationRepo: verificationRepo,
	}
}

func (s *VerificationService) InitVerification(serverID int, userID int) (*domain.InitVerificationResponse, error) {
	server, err := s.serverRepo.GetServerByID(serverID)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, errors.New("server not found")
	}

	randomBytes := make([]byte, 16)
	_, _ = rand.Read(randomBytes)
	tokenSuffix := hex.EncodeToString(randomBytes)

	challengeValue := fmt.Sprintf("hytale-verify-%s-%s", server.NameID, tokenSuffix)
	agentToken := uuid.New().String()

	vDNS := &domain.ServerVerification{
		ServerID:       serverID,
		UserID:         userID,
		ChallengeType:  domain.ChallengeDNS,
		ChallengeValue: challengeValue,
	}
	if err := s.verificationRepo.CreateVerification(vDNS); err != nil {
		return nil, fmt.Errorf("failed to create DNS verification challenge: %w", err)
	}

	vAgent := &domain.ServerVerification{
		ServerID:       serverID,
		UserID:         userID,
		ChallengeType:  domain.ChallengeAgent,
		ChallengeValue: agentToken,
	}
	if err := s.verificationRepo.CreateVerification(vAgent); err != nil {
		return nil, fmt.Errorf("failed to create agent verification challenge: %w", err)
	}

	domainName := ""
	if server.Domain != nil {
		domainName = *server.Domain
	}

	instructions := fmt.Sprintf(
		"Option 1 (DNS): Create a TXT record for %s with value '%s'.\nOption 2 (Agent): Install the monitoring mod and set agent_token to '%s'.",
		domainName, challengeValue, agentToken,
	)

	return &domain.InitVerificationResponse{
		ServerID:       serverID,
		Domain:         domainName,
		ChallengeType:  "DNS_TXT / AGENT_TOKEN",
		ChallengeValue: challengeValue,
		AgentToken:     agentToken,
		Instructions:   instructions,
	}, nil
}

func (s *VerificationService) ConfirmDNSVerification(serverID int, userID int) (bool, string, error) {
	server, err := s.serverRepo.GetServerByID(serverID)
	if err != nil || server == nil {
		return false, "server not found", nil
	}

	if server.Domain == nil || *server.Domain == "" {
		return false, "server does not have a registered domain name", nil
	}

	pending, err := s.verificationRepo.GetPendingVerification(serverID, domain.ChallengeDNS)
	if err != nil || pending == nil {
		return false, "no pending DNS verification challenge found", nil
	}

	domainToLookup := *server.Domain
	if strings.Contains(domainToLookup, ":") {
		host, _, err := net.SplitHostPort(domainToLookup)
		if err == nil {
			domainToLookup = host
		}
	}

	txtRecords, err := net.LookupTXT(domainToLookup)
	if err != nil {
		return false, fmt.Sprintf("DNS lookup failed: %v", err), nil
	}

	matched := false
	for _, record := range txtRecords {
		if strings.TrimSpace(record) == pending.ChallengeValue || strings.Contains(record, pending.ChallengeValue) {
			matched = true
			break
		}
	}

	if !matched {
		return false, fmt.Sprintf("DNS TXT challenge value not found in records for %s", domainToLookup), nil
	}

	if err := s.verificationRepo.UpdateVerificationStatus(pending.ID, domain.StatusVerified); err != nil {
		return false, "failed to update verification status", err
	}

	if err := s.serverRepo.SetVerified(serverID, userID); err != nil {
		return false, "failed to set server verified", err
	}

	return true, "server ownership successfully verified via DNS TXT record", nil
}
