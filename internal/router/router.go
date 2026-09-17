package router

import (
	"net/http"

	"github.com/Orsacle/WireHook/internal/config"
	"github.com/Orsacle/WireHook/internal/webhook"
)

func New(cfg *config.Config) *http.ServeMux {
	mux := http.NewServeMux()

	for _, route := range cfg.Routes {
		mux.HandleFunc(route.Path, webhook.Handler(route))
	}

	return mux
}
