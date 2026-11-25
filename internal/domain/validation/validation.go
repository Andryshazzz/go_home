package validation

import (
	"net/mail"
	"strings"
	"unicode/utf8"

	apperrors "go_home/pkg/errors"
)

const (
	LoginMinLen      = 3
	LoginMaxLen      = 64
	EmailMaxLen      = 255
	AgeMin           = 1
	AgeMax           = 120
	BioMaxLen        = 1000
	PasswordMinLen   = 8
	PasswordMaxBytes = 72
)

func invalid(field, reason string) error {
	return apperrors.ErrInvalidArgument.WithDetails(apperrors.FieldError{Field: field, Reason: reason})
}

func Login(raw string) (string, error) {
	login := strings.TrimSpace(raw)

	switch n := utf8.RuneCountInString(login); {
	case n < LoginMinLen:
		return "", invalid("login", "min")
	case n > LoginMaxLen:
		return "", invalid("login", "max")
	}

	return login, nil
}

func Email(raw string) (string, error) {
	email := strings.TrimSpace(raw)

	if utf8.RuneCountInString(email) > EmailMaxLen {
		return "", invalid("email", "max")
	}

	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return "", invalid("email", "email")
	}

	return email, nil
}

func Age(age int) error {
	if age < AgeMin || age > AgeMax {
		return invalid("age", "range")
	}

	return nil
}

func Bio(bio string) error {
	if utf8.RuneCountInString(bio) > BioMaxLen {
		return invalid("bio", "max")
	}

	return nil
}

func Password(password string) error {
	switch n := len(password); {
	case n < PasswordMinLen:
		return invalid("password", "min")
	case n > PasswordMaxBytes:
		return invalid("password", "max")
	}

	return nil
}
