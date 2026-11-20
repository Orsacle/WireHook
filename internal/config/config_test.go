package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoad(t *testing.T) {
	path := writeConfig(t, `
listen_addr: ":9090"
routes:
  - path: /hooks/github
    provider: github
    secret: "s3cret"
    targets:
      - "http://localhost:9000/ingest"
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.ListenAddr != ":9090" {
		t.Errorf("ListenAddr = %q, want :9090", cfg.ListenAddr)
	}
	if len(cfg.Routes) != 1 {
		t.Fatalf("len(Routes) = %d, want 1", len(cfg.Routes))
	}
	want := Route{
		Path:     "/hooks/github",
		Provider: "github",
		Secret:   "s3cret",
		Targets:  []string{"http://localhost:9000/ingest"},
	}
	got := cfg.Routes[0]
	if got.Path != want.Path || got.Provider != want.Provider || got.Secret != want.Secret ||
		len(got.Targets) != 1 || got.Targets[0] != want.Targets[0] {
		t.Errorf("Routes[0] = %+v, want %+v", got, want)
	}
}

func TestLoadDefaultListenAddr(t *testing.T) {
	cfg, err := Load(writeConfig(t, "routes: []\n"))
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.ListenAddr != ":8080" {
		t.Errorf("ListenAddr = %q, want :8080", cfg.ListenAddr)
	}
}

func TestLoadErrors(t *testing.T) {
	t.Run("missing file", func(t *testing.T) {
		if _, err := Load(filepath.Join(t.TempDir(), "nope.yaml")); err == nil {
			t.Error("want error, got nil")
		}
	})
	t.Run("invalid yaml", func(t *testing.T) {
		if _, err := Load(writeConfig(t, "routes: [unclosed")); err == nil {
			t.Error("want error, got nil")
		}
	})
}

func TestLoadExampleConfig(t *testing.T) {
	cfg, err := Load(filepath.Join("..", "..", "configs", "config.example.yaml"))
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if len(cfg.Routes) == 0 {
		t.Error("example config has no routes")
	}
}
