package profileservice

import (
	"context"
	"fmt"
	"go_home/internal/domain/entity"
	apperrors "go_home/pkg/errors"
)

const (
	defaultLimit = 20
	maxLimit     = 100
	maxLoginLen  = 64
)

func (s *ProfileService) ListUsers(
	ctx context.Context,
	req entity.ListUsersRequest,
) (*entity.ListUsersResponse, error) {
	if req.Limit <= 0 {
		req.Limit = defaultLimit
	}
	if req.Limit > maxLimit {
		return nil, apperrors.ErrInvalidArgument
	}
	if req.Offset < 0 {
		return nil, apperrors.ErrInvalidArgument
	}
	if len([]rune(req.Login)) > maxLoginLen {
		return nil, apperrors.ErrInvalidArgument
	}

	users, total, err := s.profileRepository.ListUsers(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%w: list users: %w", apperrors.ErrInternal, err)
	}

	return &entity.ListUsersResponse{
		Users: users,
		Total: total,
	}, nil
}
