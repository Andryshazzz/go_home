package dto

import (
	apperrors "go_home/pkg/errors"
	"time"
)

type TokenPairResponse struct {
	AccessToken  string `json:"access_token"  example:"eyJhbGciOiJIUzI1NiIs..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIs..."`
}

type LoginResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}

type UserResponse struct {
	ID    string `json:"id"`
	Login string `json:"login"`
	Email string `json:"email"`
	Age   int    `json:"age"`
	Bio   string `json:"bio"`
}

type ListUsersResponse struct {
	Items  []UserResponse `json:"items"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

type AvatarResponse struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}

type ErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Details []apperrors.FieldError `json:"details,omitempty"`
}
