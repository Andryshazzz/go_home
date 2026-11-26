package profileservice

import (
	"context"
	"errors"
	"fmt"
	apperrors "go_home/pkg/errors"

	"github.com/google/uuid"
)

func (s *ProfileService) DeleteMyProfile(ctx context.Context, userID uuid.UUID) error {
	if err := s.profileRepository.DeleteUser(ctx, userID); err != nil {
		if errors.Is(err, apperrors.ErrNotFoundUser) {
			return err
		}

		return fmt.Errorf("%w: delete user: %w", apperrors.ErrInternal, err)
	}

	return nil
}
