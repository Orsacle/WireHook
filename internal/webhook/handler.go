package webhook

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"github.com/Orsacle/wirehook/internal/config"
	"github.com/Orsacle/wirehook/internal/provider"
)

func Handler(route config.Route) http.HandlerFunc {
	p, ok := provider.Get(route.Provider)
	if !ok {
		log.Fatalf("unknown provider %q for route %q", route.Provider, route.Path)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "cannot read body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if err := p.Verify(r, body, route.Secret); err != nil {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}

		event, err := p.Parse(r, body)
		if err != nil {
			http.Error(w, "cannot parse event", http.StatusBadRequest)
			return
		}

		forward(route.Targets, event.Payload)
		w.WriteHeader(http.StatusAccepted)
	}
}

func forward(targets []string, payload []byte) {
	for _, target := range targets {
		resp, err := http.Post(target, "application/json", bytes.NewReader(payload))
		if err != nil {
			log.Printf("forward to %s failed: %v", target, err)
			continue
		}
		resp.Body.Close()
	}
}
