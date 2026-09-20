package provider

import "net/http"

type Event struct {
	Provider string
	Type     string
	Payload  []byte
}

type Provider interface {
	Verify(r *http.Request, body []byte, secret string) error
	Parse(r *http.Request, body []byte) (Event, error)
}

func Get(name string) (Provider, bool) {
	p, ok := registry[name]
	return p, ok
}

var registry = map[string]Provider{
	"github": githubProvider{},
	"gitlab": gitlabProvider{},
}
