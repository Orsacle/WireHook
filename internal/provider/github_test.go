package provider

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
)

func githubSig(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestGithubVerify(t *testing.T) {
	body := []byte(`{"action":"opened"}`)
	tests := []struct {
		name    string
		sig     string
		wantErr bool
	}{
		{"valid", githubSig("s3cret", body), false},
		{"valid with surrounding whitespace", " " + githubSig("s3cret", body) + " ", false},
		{"missing header", "", true},
		{"wrong secret", githubSig("other", body), true},
		{"missing prefix", githubSig("s3cret", body)[len("sha256="):], true},
		{"tampered body", githubSig("s3cret", []byte("tampered")), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/hooks/github", nil)
			if tt.sig != "" {
				r.Header.Set("X-Hub-Signature-256", tt.sig)
			}
			err := githubProvider{}.Verify(r, body, "s3cret")
			if (err != nil) != tt.wantErr {
				t.Fatalf("Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGithubParse(t *testing.T) {
	body := []byte(`{}`)

	r := httptest.NewRequest(http.MethodPost, "/hooks/github", nil)
	r.Header.Set("X-GitHub-Event", "push")
	ev, err := githubProvider{}.Parse(r, body)
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}
	if ev.Provider != "github" || ev.Type != "push" || string(ev.Payload) != "{}" {
		t.Errorf("Parse() = %+v", ev)
	}

	r = httptest.NewRequest(http.MethodPost, "/hooks/github", nil)
	if _, err := (githubProvider{}).Parse(r, body); err == nil {
		t.Error("Parse() without event header: want error, got nil")
	}
}
