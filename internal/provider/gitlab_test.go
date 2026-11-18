package provider

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGitlabVerify(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{"valid", "s3cret", false},
		{"missing header", "", true},
		{"wrong token", "other", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/hooks/gitlab", nil)
			if tt.token != "" {
				r.Header.Set("X-Gitlab-Token", tt.token)
			}
			err := gitlabProvider{}.Verify(r, nil, "s3cret")
			if (err != nil) != tt.wantErr {
				t.Fatalf("Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGitlabParse(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/hooks/gitlab", nil)
	r.Header.Set("X-Gitlab-Event", "Push Hook")
	ev, err := gitlabProvider{}.Parse(r, []byte(`{}`))
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}
	if ev.Provider != "gitlab" || ev.Type != "Push Hook" || string(ev.Payload) != "{}" {
		t.Errorf("Parse() = %+v", ev)
	}

	r = httptest.NewRequest(http.MethodPost, "/hooks/gitlab", nil)
	if _, err := (gitlabProvider{}).Parse(r, nil); err == nil {
		t.Error("Parse() without event header: want error, got nil")
	}
}

func TestGet(t *testing.T) {
	for _, name := range []string{"github", "gitlab"} {
		if _, ok := Get(name); !ok {
			t.Errorf("Get(%q) not found", name)
		}
	}
	if _, ok := Get("bitbucket"); ok {
		t.Error("Get(\"bitbucket\") found, want not found")
	}
}
