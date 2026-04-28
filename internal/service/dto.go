package service

import (
	"finance_tracker/internal/apperrors"
	"net/mail"
	"strings"
)

func IsValidEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return apperrors.ErrInvalidEmail
	}

	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address == "" {
		return apperrors.ErrInvalidEmail
	}

	// ParseAddress accepts "Name <a@b.com>", but we want a plain email string.
	if addr.Address != email {
		return apperrors.ErrInvalidEmail
	}

	return nil
}

