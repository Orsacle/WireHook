package router

import (
	"net/http"

	"github.com/Orsacle/wirehook/internal/config"
	"github.com/Orsacle/wirehook/internal/webhook"
)

func New(cfg *config.Config) *http.ServeMux {
	mux := http.NewServeMux()

	for _, route := range cfg.Routes {
		mux.HandleFunc(route.Path, webhook.Handler(route))
	}

	return mux
}
