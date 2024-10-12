package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	ll "github.com/ewen-lbh/label-logger-go"
)

func redactURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}

	if u.User == nil {
		return u.String()
	}

	u.User = url.UserPassword(u.User.Username(), "REDACTED")
	return u.String()
}

func decodeRequest(w http.ResponseWriter, r *http.Request, v any) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		ll.ErrorDisplay("could not read request body: %w", err)
		http.Error(w, "could not read request body", http.StatusBadRequest)
	}

	err = json.Unmarshal(body, &v)
	if err != nil {
		ll.ErrorDisplay("could not decode json", err)
		http.Error(w, "could not decode json", http.StatusBadRequest)
	}
}
