package provider

import (
	"fmt"
	"net/http"
)

type gitlabProvider struct{}

func (gitlabProvider) Verify(r *http.Request, body []byte, secret string) error {
	token := r.Header.Get("X-Gitlab-Token")
	if token != secret {
		return fmt.Errorf("token mismatch")
	}
	return nil
}

func (gitlabProvider) Parse(r *http.Request, body []byte) (Event, error) {
	eventType := r.Header.Get("X-Gitlab-Event")
	if eventType == "" {
		return Event{}, fmt.Errorf("missing event type header")
	}
	return Event{Provider: "gitlab", Type: eventType, Payload: body}, nil
}
