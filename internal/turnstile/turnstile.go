package turnstile

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

type Verifier struct {
	secretKey  string
	httpClient *http.Client
	isDev      bool
}

func NewVerifier(secretKey string, isDev bool) *Verifier {
	return &Verifier{
		secretKey: secretKey,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		isDev: isDev,
	}
}

type siteVerifyRequest struct {
	Secret   string `json:"secret"`
	Response string `json:"response"`
	RemoteIP string `json:"remoteip,omitempty"`
}

type siteVerifyResponse struct {
	Success bool `json:"success"`
}

func (v *Verifier) Verify(remoteIP, token string) bool {
	if token == "" {
		return false
	}
	if v.secretKey == "" || v.isDev {
		if token == "mock-token" || token == "test" || token == "XXXX.DUMMY.TOKEN.XXXX" {
			return true
		}
	}

	reqBody, err := json.Marshal(siteVerifyRequest{
		Secret:   v.secretKey,
		Response: token,
		RemoteIP: remoteIP,
	})
	if err != nil {
		return false
	}

	resp, err := v.httpClient.Post(
		"https://challenges.cloudflare.com/turnstile/v0/siteverify",
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	var verifyResp siteVerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&verifyResp); err != nil {
		return false
	}
	return verifyResp.Success
}
