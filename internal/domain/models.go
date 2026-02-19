package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type Region string

const (
	RegionCIS          Region = "CIS"
	RegionEurope       Region = "EUROPE"
	RegionAsia         Region = "ASIA"
	RegionAfrica       Region = "AFRICA"
	RegionNorthAmerica Region = "NORTHAMERICA"
	RegionSouthAmerica Region = "SOUTHAMERICA"
)

var ValidRegions = map[Region]bool{
	RegionCIS:          true,
	RegionEurope:       true,
	RegionAsia:         true,
	RegionAfrica:       true,
	RegionNorthAmerica: true,
	RegionSouthAmerica: true,
}

type Tag string

const (
	TagPVP       Tag = "PVP"
	TagPVE       Tag = "PVE"
	TagRP        Tag = "RP"
	TagRPG       Tag = "RPG"
	TagVanilla   Tag = "VANILLA"
	TagSurvival  Tag = "SURVIVAL"
	TagMinigames Tag = "MINIGAMES"
)

var ValidTags = map[Tag]bool{
	TagPVP:       true,
	TagPVE:       true,
	TagRP:        true,
	TagRPG:       true,
	TagVanilla:   true,
	TagSurvival:  true,
	TagMinigames: true,
}

type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	return json.Marshal(s)
}

func (s *StringSlice) Scan(src any) error {
	if src == nil {
		*s = []string{}
		return nil
	}
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		return errors.New("unsupported type for StringSlice scan")
	}
}

type User struct {
	ID        int        `db:"id" json:"id"`
	Name      string     `db:"name" json:"name"`
	Email     *string    `db:"email" json:"email,omitempty"`
	Avatar    *string    `db:"avatar" json:"avatar,omitempty"`
	Password  *string    `db:"password" json:"-"`
	DiscordID *string    `db:"discord_id" json:"discordId,omitempty"`
	GoogleID  *string    `db:"google_id" json:"googleId,omitempty"`
	CreatedAt time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time  `db:"updated_at" json:"updatedAt"`
}

type Server struct {
	ID                int         `db:"id" json:"id"`
	Domain            *string     `db:"domain" json:"domain,omitempty"`
	RealIP            *string     `db:"real_ip" json:"ip,omitempty"`
	Description       string      `db:"description" json:"description"`
	Name              string      `db:"name" json:"name"`
	NameID            string      `db:"name_id" json:"nameId"`
	Tags              StringSlice `db:"tags" json:"tags"`
	Region            string      `db:"region" json:"region"`
	Poster            *string     `db:"poster" json:"poster,omitempty"`
	LikesQuantity     int         `db:"likes_quantity" json:"likesQuantity"`
	UserID            *int        `db:"user_id" json:"userId,omitempty"`
	IsVerified        bool        `db:"is_verified" json:"isVerified"`
	VerificationToken *string     `db:"verification_token" json:"-"`
	CreatedAt         time.Time   `db:"created_at" json:"createdAt"`
	UpdatedAt         time.Time   `db:"updated_at" json:"updatedAt"`
}

type Like struct {
	UserID    int       `db:"user_id" json:"userId"`
	ServerID  int       `db:"server_id" json:"serverId"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

type ServerMetric struct {
	ID            int64     `db:"id" json:"id"`
	ServerID      int       `db:"server_id" json:"serverId"`
	IsOnline      bool      `db:"is_online" json:"isOnline"`
	PlayersOnline int       `db:"players_online" json:"playersOnline"`
	PlayersMax    int       `db:"players_max" json:"playersMax"`
	TPS           float64   `db:"tps" json:"tps"`
	RAMUsedMB     int       `db:"ram_used_mb" json:"ramUsedMb"`
	LatencyMS     int       `db:"latency_ms" json:"latencyMs"`
	RecordedAt    time.Time `db:"recorded_at" json:"recordedAt"`
}

type ChallengeType string

const (
	ChallengeDNS   ChallengeType = "DNS_TXT"
	ChallengeAgent ChallengeType = "AGENT_TOKEN"
)

type VerificationStatus string

const (
	StatusPending  VerificationStatus = "PENDING"
	StatusVerified VerificationStatus = "VERIFIED"
	StatusFailed   VerificationStatus = "FAILED"
)

type ServerVerification struct {
	ID             int                `db:"id" json:"id"`
	ServerID       int                `db:"server_id" json:"serverId"`
	UserID         int                `db:"user_id" json:"userId"`
	ChallengeType  ChallengeType      `db:"challenge_type" json:"challengeType"`
	ChallengeValue string             `db:"challenge_value" json:"challengeValue"`
	Status         VerificationStatus `db:"status" json:"status"`
	CreatedAt      time.Time          `db:"created_at" json:"createdAt"`
	VerifiedAt     *time.Time         `db:"verified_at" json:"verifiedAt,omitempty"`
}
