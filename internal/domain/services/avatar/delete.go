package avatarservice

import (
	"context"
	"errors"
	"fmt"

	apperrors "go_home/pkg/errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *AvatarService) Delete(ctx context.Context, userID, avatarID uuid.UUID) error {
	if err := s.avatarRepository.SoftDeleteAvatar(ctx, userID, avatarID); err != nil {
		if errors.Is(err, apperrors.ErrNotFoundAvatar) {
			return err
		}

		return fmt.Errorf("%w: delete avatar: %w", apperrors.ErrInternal, err)
	}

	s.logFor(ctx).Info("avatar deleted",
		zap.String("user_id", userID.String()),
		zap.String("avatar_id", avatarID.String()),
	)

	return nil
}
