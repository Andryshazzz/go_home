package profileservice

import (
	"context"
	"fmt"
	"time"

	"go_home/internal/domain/entity"
	"go_home/internal/domain/validation"
	apperrors "go_home/pkg/errors"
)

const (
	activeUserMinAvatars = 3
	avatarURLTTL         = time.Hour
)

func (s *ProfileService) ListActiveUsers(
	ctx context.Context,
	req entity.ListActiveUsersRequest,
) ([]entity.ActiveUser, error) {
	if err := validation.AgeRange(req.MinAge, req.MaxAge); err != nil {
		return nil, err
	}

	if req.Limit <= 0 {
		req.Limit = defaultLimit
	}
	if req.Limit > maxLimit {
		return nil, apperrors.ErrInvalidArgument.WithDetails(
			apperrors.FieldError{Field: "limit", Reason: "max"},
		)
	}
	if req.Offset < 0 {
		return nil, apperrors.ErrInvalidArgument.WithDetails(
			apperrors.FieldError{Field: "offset", Reason: "min"},
		)
	}

	users, err := s.profileRepository.ListActiveUsers(ctx, req, activeUserMinAvatars)
	if err != nil {
		return nil, fmt.Errorf("%w: list active users: %w", apperrors.ErrInternal, err)
	}

	for i := range users {
		url, err := s.avatarURLs.PresignedGetURL(ctx, users[i].LastAvatar.ObjectKey, avatarURLTTL)
		if err != nil {
			return nil, fmt.Errorf("%w: avatar url for user %s: %w", apperrors.ErrInternal, users[i].User.ID, err)
		}

		users[i].LastAvatar.URL = url
	}

	return users, nil
}
