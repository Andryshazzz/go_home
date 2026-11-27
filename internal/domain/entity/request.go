package entity

import (
	"io"
	"time"

	"github.com/google/uuid"
)

type RegisterUserRequest struct {
	Login    string
	Email    string
	Password string
	Age      int
	Bio      string
}

type LoginUserRequest struct {
	Login    string
	Password string
}

type LogoutUserRequest struct {
	RefreshToken string
}

type ListUsersRequest struct {
	Login  string
	Limit  int
	Offset int
}

type UpdateProfileRequest struct {
	UserID uuid.UUID
	Login  *string
	Email  *string
	Age    *int
	Bio    *string
}

type RefreshTokenRequest struct {
	RefreshToken string
}

type RotateSessionRequest struct {
	SessionID    uuid.UUID
	UserID       uuid.UUID
	OldTokenHash string
	NewTokenHash string
	ExpiresAt    time.Time
}

type UploadAvatarRequest struct {
	UserID uuid.UUID
	File   io.ReadSeeker
	Size   int64
}
