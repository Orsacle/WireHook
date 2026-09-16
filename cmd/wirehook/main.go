package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Orsacle/WireHook/internal/config"
	"github.com/Orsacle/WireHook/internal/router"
)

func main() {
	cfgPath := os.Getenv("WIREHOOK_CONFIG")
	if cfgPath == "" {
		cfgPath = "configs/config.yaml"
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	mux := router.New(cfg)

	log.Printf("wirehook listening on %s", cfg.ListenAddr)
	if err := http.ListenAndServe(cfg.ListenAddr, mux); err != nil {
		log.Fatalf("server: %v", err)
	}
}
