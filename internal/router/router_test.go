package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Orsacle/wirehook/internal/config"
)

func TestNewRoutesByPath(t *testing.T) {
	mux := New(&config.Config{
		Routes: []config.Route{
			{Path: "/hooks/github", Provider: "github", Secret: "s"},
			{Path: "/hooks/gitlab", Provider: "gitlab", Secret: "s"},
		},
	})

	tests := []struct {
		name string
		path string
		want int
	}{
		{"github route", "/hooks/github", http.StatusUnauthorized},
		{"gitlab route", "/hooks/gitlab", http.StatusUnauthorized},
		{"unknown path", "/hooks/other", http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, httptest.NewRequest(http.MethodPost, tt.path, strings.NewReader("{}")))
			if w.Code != tt.want {
				t.Errorf("status = %d, want %d", w.Code, tt.want)
			}
		})
	}
}

func TestNewNoRoutes(t *testing.T) {
	mux := New(&config.Config{})
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/anything", nil))
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}
