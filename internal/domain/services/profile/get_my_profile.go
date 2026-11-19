package profileservice

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go_home/internal/domain/entity"
	apperrors "go_home/pkg/errors"
)

func (s *ProfileService) GetMyProfile(ctx context.Context, userID uuid.UUID) (entity.User, error) {
	user, err := s.profileRepository.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFoundUser) {
			return entity.User{}, err
		}

		return entity.User{}, fmt.Errorf("%w: %w", apperrors.ErrInternal, err)
	}

	return user, nil
}
