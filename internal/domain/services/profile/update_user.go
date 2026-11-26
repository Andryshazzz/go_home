package profileservice

import (
	"context"
	"errors"
	"fmt"
	"go_home/internal/domain/entity"
	"go_home/internal/domain/validation"
	apperrors "go_home/pkg/errors"
)

func (s *ProfileService) UpdateMyProfile(
	ctx context.Context,
	req entity.UpdateProfileRequest,
) (entity.User, error) {
	if req.Login == nil && req.Email == nil && req.Age == nil && req.Bio == nil {
		return entity.User{}, apperrors.ErrInvalidArgument
	}

	if req.Login != nil {
		login, err := validation.Login(*req.Login)
		if err != nil {
			return entity.User{}, err
		}
		req.Login = &login
	}

	if req.Email != nil {
		email, err := validation.Email(*req.Email)
		if err != nil {
			return entity.User{}, err
		}
		req.Email = &email
	}

	if req.Age != nil {
		if err := validation.Age(*req.Age); err != nil {
			return entity.User{}, err
		}
	}

	if req.Bio != nil {
		if err := validation.Bio(*req.Bio); err != nil {
			return entity.User{}, err
		}
	}

	user, err := s.profileRepository.UpdateUser(ctx, req)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFoundUser) ||
			errors.Is(err, apperrors.ErrUserAlreadyExists) {
			return entity.User{}, err
		}

		return entity.User{}, fmt.Errorf("%w: update user: %w", apperrors.ErrInternal, err)
	}

	return user, nil
}
