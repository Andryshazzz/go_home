package profileservice

import (
	"context"
	"go_home/internal/domain/entity"
	"time"

	"github.com/google/uuid"
)

type ProfileRepository interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (entity.User, error)
	ListUsers(ctx context.Context, filter entity.ListUsersRequest) ([]entity.User, int, error)
	UpdateUser(ctx context.Context, req entity.UpdateProfileRequest) (entity.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	ListActiveUsers(ctx context.Context, req entity.ListActiveUsersRequest, minAvatars int) ([]entity.ActiveUser, error)
}

type AvatarURLs interface {
	PresignedGetURL(ctx context.Context, key string, ttl time.Duration) (string, error)
}

type ProfileService struct {
	profileRepository ProfileRepository
	avatarURLs        AvatarURLs
}

func NewProfileService(profileRepository ProfileRepository, avatarURLs AvatarURLs) *ProfileService {
	return &ProfileService{
		profileRepository: profileRepository,
		avatarURLs:        avatarURLs,
	}
}
