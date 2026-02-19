package domain

import (
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var kebabCaseRegex = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

type CreateServerDTO struct {
	IP          *string  `json:"ip,omitempty"`
	Domain      *string  `json:"domain,omitempty"`
	Description string   `json:"description"`
	Name        string   `json:"name"`
	NameID      string   `json:"nameId"`
	Poster      *string  `json:"poster,omitempty"`
	Tags        []string `json:"tags"`
	Region      string   `json:"region"`
}

func (dto *CreateServerDTO) Validate() map[string][]string {
	errs := make(map[string][]string)

	hasIP := dto.IP != nil && strings.TrimSpace(*dto.IP) != ""
	hasDomain := dto.Domain != nil && strings.TrimSpace(*dto.Domain) != ""
	if !hasIP && !hasDomain {
		errs["_atLeastOne"] = []string{"At least one of the following properties must be provided: ip, domain"}
	}

	if hasIP {
		ipVal := strings.TrimSpace(*dto.IP)
		if !isValidHostPort(ipVal) {
			errs["ip"] = []string{fmt.Sprintf("host or port is not reachable in %s", ipVal)}
		}
	}

	if hasDomain {
		domainVal := strings.TrimSpace(*dto.Domain)
		if !isValidDomain(domainVal) {
			errs["domain"] = []string{fmt.Sprintf("domain %s does not exist or incorrect format", domainVal)}
		}
	}

	nameTrimmed := strings.TrimSpace(dto.Name)
	if len(dto.Name) != len(nameTrimmed) {
		errs["name"] = append(errs["name"], "name must be trimmed")
	}
	if len(nameTrimmed) < 2 || len(nameTrimmed) > 30 {
		errs["name"] = append(errs["name"], "name must be longer than or equal to 2 and shorter than or equal to 30 characters")
	}

	nameIDTrimmed := strings.TrimSpace(dto.NameID)
	if len(dto.NameID) != len(nameIDTrimmed) {
		errs["nameId"] = append(errs["nameId"], "nameId must be trimmed")
	}
	if len(nameIDTrimmed) < 2 || len(nameIDTrimmed) > 30 {
		errs["nameId"] = append(errs["nameId"], "nameId must be longer than or equal to 2 and shorter than or equal to 30 characters")
	}
	if !kebabCaseRegex.MatchString(nameIDTrimmed) {
		errs["nameId"] = append(errs["nameId"], fmt.Sprintf("%s is not kebab-case", nameIDTrimmed))
	}

	descTrimmed := strings.TrimSpace(dto.Description)
	if len(dto.Description) != len(descTrimmed) {
		errs["description"] = append(errs["description"], "description must be trimmed")
	}
	if len(descTrimmed) < 10 || len(descTrimmed) > 300 {
		errs["description"] = append(errs["description"], "description must be longer than or equal to 10 and shorter than or equal to 300 characters")
	}

	if len(dto.Tags) == 0 {
		errs["tags"] = []string{"tags should not be empty"}
	} else {
		for _, tag := range dto.Tags {
			if !ValidTags[Tag(tag)] {
				errs["tags"] = []string{"each value in tags must be one of the following values: PVP, PVE, RP, RPG, VANILLA, SURVIVAL, MINIGAMES"}
				break
			}
		}
	}

	if !ValidRegions[Region(dto.Region)] {
		errs["region"] = []string{"region must be one of the following values: CIS, EUROPE, ASIA, AFRICA, NORTHAMERICA, SOUTHAMERICA"}
	}

	if dto.Poster != nil && strings.TrimSpace(*dto.Poster) != "" {
		if _, err := url.ParseRequestURI(*dto.Poster); err != nil {
			errs["poster"] = []string{"poster must be an URL address"}
		}
	}

	return errs
}

type UpdateServerDTO struct {
	IP          *string   `json:"ip,omitempty"`
	Domain      *string   `json:"domain,omitempty"`
	Description *string   `json:"description,omitempty"`
	Name        *string   `json:"name,omitempty"`
	NameID      *string   `json:"nameId,omitempty"`
	Poster      *string   `json:"poster,omitempty"`
	Tags        *[]string `json:"tags,omitempty"`
	Region      *string   `json:"region,omitempty"`
}

func (dto *UpdateServerDTO) Validate() map[string][]string {
	errs := make(map[string][]string)

	if dto.IP != nil && strings.TrimSpace(*dto.IP) != "" {
		if !isValidHostPort(*dto.IP) {
			errs["ip"] = []string{fmt.Sprintf("host or port is not reachable in %s", *dto.IP)}
		}
	}

	if dto.Domain != nil && strings.TrimSpace(*dto.Domain) != "" {
		if !isValidDomain(*dto.Domain) {
			errs["domain"] = []string{fmt.Sprintf("domain %s does not exist or incorrect format", *dto.Domain)}
		}
	}

	if dto.Name != nil {
		t := strings.TrimSpace(*dto.Name)
		if len(*dto.Name) != len(t) {
			errs["name"] = append(errs["name"], "name must be trimmed")
		}
		if len(t) < 2 || len(t) > 30 {
			errs["name"] = append(errs["name"], "name must be longer than or equal to 2 and shorter than or equal to 30 characters")
		}
	}

	if dto.NameID != nil {
		t := strings.TrimSpace(*dto.NameID)
		if len(*dto.NameID) != len(t) {
			errs["nameId"] = append(errs["nameId"], "nameId must be trimmed")
		}
		if len(t) < 2 || len(t) > 30 {
			errs["nameId"] = append(errs["nameId"], "nameId must be longer than or equal to 2 and shorter than or equal to 30 characters")
		}
		if !kebabCaseRegex.MatchString(t) {
			errs["nameId"] = append(errs["nameId"], fmt.Sprintf("%s is not kebab-case", t))
		}
	}

	if dto.Description != nil {
		t := strings.TrimSpace(*dto.Description)
		if len(*dto.Description) != len(t) {
			errs["description"] = append(errs["description"], "description must be trimmed")
		}
		if len(t) < 10 || len(t) > 300 {
			errs["description"] = append(errs["description"], "description must be longer than or equal to 10 and shorter than or equal to 300 characters")
		}
	}

	if dto.Tags != nil {
		for _, tag := range *dto.Tags {
			if !ValidTags[Tag(tag)] {
				errs["tags"] = []string{"each value in tags must be one of the following values: PVP, PVE, RP, RPG, VANILLA, SURVIVAL, MINIGAMES"}
				break
			}
		}
	}

	if dto.Region != nil && !ValidRegions[Region(*dto.Region)] {
		errs["region"] = []string{"region must be one of the following values: CIS, EUROPE, ASIA, AFRICA, NORTHAMERICA, SOUTHAMERICA"}
	}

	return errs
}

type FiltersDTO struct {
	Tags   []string `json:"tags,omitempty"`
	Region *string  `json:"region,omitempty"`
}

type OrderDTO struct {
	Likes     *string `json:"likes,omitempty"`
	CreatedAt *string `json:"createdAt,omitempty"`
}

type FindPageDTO struct {
	Page     int         `json:"page"`
	Quantity int         `json:"quantity"`
	Filters  *FiltersDTO `json:"filters,omitempty"`
	Order    *OrderDTO   `json:"order,omitempty"`
}

func (dto *FindPageDTO) Validate() map[string][]string {
	errs := make(map[string][]string)
	if dto.Page < 0 {
		errs["page"] = []string{"page must not be less than 0"}
	}
	if dto.Quantity < 1 || dto.Quantity > 50 {
		errs["quantity"] = []string{"quantity must not be less than 1 and not greater than 50"}
	}
	return errs
}

type ServerResponse struct {
	ID            int      `json:"id"`
	Domain        *string  `json:"domain,omitempty"`
	IP            *string  `json:"ip,omitempty"`
	Description   string   `json:"description"`
	Name          string   `json:"name"`
	NameID        string   `json:"nameId"`
	Tags          []string `json:"tags"`
	Poster        *string  `json:"poster,omitempty"`
	LikesQuantity int      `json:"likesQuantity"`
	Liked         bool     `json:"liked"`
	Players       int      `json:"players"`
	IsVerified    bool     `json:"isVerified"`
}

type ServerStatusResponse struct {
	ID       int   `json:"id"`
	IsOnline *bool `json:"isOnline"`
}

type RegisterDTO struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (dto *RegisterDTO) Validate() map[string][]string {
	errs := make(map[string][]string)

	nameTrimmed := strings.TrimSpace(dto.Name)
	if len(dto.Name) != len(nameTrimmed) {
		errs["name"] = append(errs["name"], "name must be trimmed")
	}
	if len(nameTrimmed) < 3 {
		errs["name"] = append(errs["name"], "name must be longer than or equal to 3 characters")
	}

	dto.Email = strings.ToLower(strings.TrimSpace(dto.Email))
	if _, err := mail.ParseAddress(dto.Email); err != nil || !strings.Contains(dto.Email, "@") {
		errs["email"] = []string{"email must be an email"}
	}

	if len(dto.Password) < 6 {
		errs["password"] = []string{"password must be longer than or equal to 6 characters"}
	}

	return errs
}

type LoginDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (dto *LoginDTO) Validate() map[string][]string {
	errs := make(map[string][]string)
	dto.Email = strings.ToLower(strings.TrimSpace(dto.Email))
	if _, err := mail.ParseAddress(dto.Email); err != nil || !strings.Contains(dto.Email, "@") {
		errs["email"] = []string{"email must be an email"}
	}
	if len(dto.Password) == 0 {
		errs["password"] = []string{"password should not be empty"}
	}
	return errs
}

type LeaderboardItem struct {
	ServerID      int     `json:"serverId"`
	NameID        string  `json:"nameId"`
	Name          string  `json:"name"`
	LikesQuantity int     `json:"likesQuantity"`
	PlayersOnline int     `json:"playersOnline"`
	Rank          int     `json:"rank"`
	IsOnline      bool    `json:"isOnline"`
	Score         float64 `json:"score"`
}

type InitVerificationResponse struct {
	ServerID       int    `json:"serverId"`
	Domain         string `json:"domain"`
	ChallengeType  string `json:"challengeType"`
	ChallengeValue string `json:"challengeValue"`
	AgentToken     string `json:"agentToken"`
	Instructions   string `json:"instructions"`
}

func isValidHostPort(val string) bool {
	if !strings.Contains(val, ":") {
		return false
	}
	host, portStr, err := net.SplitHostPort(val)
	if err != nil {
		return false
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return false
	}
	return host != ""
}

func isValidDomain(val string) bool {
	if strings.Contains(val, ":") {
		host, portStr, err := net.SplitHostPort(val)
		if err != nil {
			return false
		}
		port, err := strconv.Atoi(portStr)
		if err != nil || port < 1 || port > 65535 {
			return false
		}
		val = host
	}
	return len(val) > 0 && len(val) < 255 && !strings.Contains(val, " ")
}
