package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/Orsacle/wirehook/internal/config"
)

type recorder struct {
	mu     sync.Mutex
	bodies []string
	types  []string
}

func newTarget(t *testing.T) (*httptest.Server, *recorder) {
	t.Helper()
	rec := &recorder{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		rec.mu.Lock()
		rec.bodies = append(rec.bodies, string(b))
		rec.types = append(rec.types, r.Header.Get("Content-Type"))
		rec.mu.Unlock()
	}))
	t.Cleanup(srv.Close)
	return srv, rec
}

func githubRequest(body, secret, event string) *http.Request {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	r := httptest.NewRequest(http.MethodPost, "/hooks/github", strings.NewReader(body))
	r.Header.Set("X-Hub-Signature-256", "sha256="+hex.EncodeToString(mac.Sum(nil)))
	if event != "" {
		r.Header.Set("X-GitHub-Event", event)
	}
	return r
}

func TestHandlerForwardsValidGithubEvent(t *testing.T) {
	t1, rec1 := newTarget(t)
	t2, rec2 := newTarget(t)
	h := Handler(config.Route{
		Path:     "/hooks/github",
		Provider: "github",
		Secret:   "s3cret",
		Targets:  []string{t1.URL, t2.URL},
	})

	body := `{"action":"opened"}`
	w := httptest.NewRecorder()
	h(w, githubRequest(body, "s3cret", "pull_request"))

	if w.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusAccepted)
	}
	for i, rec := range []*recorder{rec1, rec2} {
		if len(rec.bodies) != 1 || rec.bodies[0] != body {
			t.Errorf("target %d got %v, want [%s]", i+1, rec.bodies, body)
		}
		if rec.types[0] != "application/json" {
			t.Errorf("target %d Content-Type = %q", i+1, rec.types[0])
		}
	}
}

func TestHandlerRejectsBadSignature(t *testing.T) {
	target, rec := newTarget(t)
	h := Handler(config.Route{Provider: "github", Secret: "s3cret", Targets: []string{target.URL}})

	w := httptest.NewRecorder()
	h(w, githubRequest(`{}`, "wrong", "push"))

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
	if len(rec.bodies) != 0 {
		t.Errorf("payload was forwarded despite bad signature: %v", rec.bodies)
	}
}

func TestHandlerRejectsMissingEventType(t *testing.T) {
	target, rec := newTarget(t)
	h := Handler(config.Route{Provider: "github", Secret: "s3cret", Targets: []string{target.URL}})

	w := httptest.NewRecorder()
	h(w, githubRequest(`{}`, "s3cret", ""))

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	if len(rec.bodies) != 0 {
		t.Errorf("payload was forwarded despite parse error: %v", rec.bodies)
	}
}

func TestHandlerGitlab(t *testing.T) {
	target, rec := newTarget(t)
	h := Handler(config.Route{Provider: "gitlab", Secret: "s3cret", Targets: []string{target.URL}})

	r := httptest.NewRequest(http.MethodPost, "/hooks/gitlab", strings.NewReader(`{"ok":true}`))
	r.Header.Set("X-Gitlab-Token", "s3cret")
	r.Header.Set("X-Gitlab-Event", "Push Hook")
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusAccepted)
	}
	if len(rec.bodies) != 1 || rec.bodies[0] != `{"ok":true}` {
		t.Errorf("target got %v", rec.bodies)
	}
}

func TestHandlerContinuesAfterFailingTarget(t *testing.T) {
	dead := httptest.NewServer(http.NotFoundHandler())
	deadURL := dead.URL
	dead.Close()
	good, rec := newTarget(t)

	h := Handler(config.Route{Provider: "github", Secret: "s3cret", Targets: []string{deadURL, good.URL}})
	w := httptest.NewRecorder()
	h(w, githubRequest(`{}`, "s3cret", "push"))

	if w.Code != http.StatusAccepted {
		t.Errorf("status = %d, want %d", w.Code, http.StatusAccepted)
	}
	if len(rec.bodies) != 1 {
		t.Errorf("good target got %d deliveries, want 1", len(rec.bodies))
	}
}
