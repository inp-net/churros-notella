package main

import "net/url"

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
