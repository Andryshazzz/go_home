package profileservice

import (
	"context"
	"go_home/internal/domain/entity"

	"github.com/google/uuid"
)

type ProfileRepository interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (entity.User, error)
	ListUsers(ctx context.Context, filter entity.ListUsersRequest) ([]entity.User, int, error)
	UpdateUser(ctx context.Context, req entity.UpdateProfileRequest) (entity.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type ProfileService struct {
	profileRepository ProfileRepository
}

func NewProfileService(profileRepository ProfileRepository) *ProfileService {
	return &ProfileService{
		profileRepository: profileRepository,
	}
}
