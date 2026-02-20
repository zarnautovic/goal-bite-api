package auth

import (
	"errors"
	"net/mail"
	"strings"
)

var ErrInvalidEmail = errors.New("invalid email")

func NormalizeEmail(raw string) (string, error) {
	addr, err := mail.ParseAddress(strings.TrimSpace(raw))
	if err != nil {
		return "", ErrInvalidEmail
	}
	parts := strings.Split(strings.ToLower(strings.TrimSpace(addr.Address)), "@")
	if len(parts) != 2 {
		return "", ErrInvalidEmail
	}
	local := parts[0]
	domain := parts[1]
	if local == "" || domain == "" {
		return "", ErrInvalidEmail
	}

	switch domain {
	case "gmail.com", "googlemail.com":
		domain = "gmail.com"
		local = strings.Split(local, "+")[0]
		local = strings.ReplaceAll(local, ".", "")
		if local == "" {
			return "", ErrInvalidEmail
		}
	}

	return local + "@" + domain, nil
}
