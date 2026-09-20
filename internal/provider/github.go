package provider

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
)

type githubProvider struct{}

func (githubProvider) Verify(r *http.Request, body []byte, secret string) error {
	sig := r.Header.Get("X-Hub-Signature-256")
	if sig == "" {
		return fmt.Errorf("missing signature header")
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(strings.TrimSpace(sig))) {
		return fmt.Errorf("signature mismatch")
	}
	return nil
}

func (githubProvider) Parse(r *http.Request, body []byte) (Event, error) {
	eventType := r.Header.Get("X-GitHub-Event")
	if eventType == "" {
		return Event{}, fmt.Errorf("missing event type header")
	}
	return Event{Provider: "github", Type: eventType, Payload: body}, nil
}
