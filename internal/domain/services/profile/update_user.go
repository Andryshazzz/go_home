package profileservice

import (
	"context"
	"errors"
	"fmt"
	"go_home/internal/domain/entity"
	apperrors "go_home/pkg/errors"
	"strings"
)

func (s *ProfileService) UpdateMyProfile(
	ctx context.Context,
	req entity.UpdateProfileRequest,
) (entity.User, error) {
	if req.Login == nil && req.Email == nil && req.Age == nil && req.Bio == nil {
		return entity.User{}, apperrors.ErrInvalidArgument
	}

	if req.Login != nil {
		login := strings.TrimSpace(*req.Login)
		if login == "" || len([]rune(login)) > 64 {
			return entity.User{}, apperrors.ErrInvalidArgument
		}
		req.Login = &login
	}

	if req.Email != nil {
		email := strings.TrimSpace(*req.Email)
		if email == "" || len([]rune(email)) > 255 {
			return entity.User{}, apperrors.ErrInvalidArgument
		}
		req.Email = &email
	}

	if req.Age != nil && (*req.Age <= 0 || *req.Age > 120) {
		return entity.User{}, apperrors.ErrInvalidArgument
	}

	if req.Bio != nil && len([]rune(*req.Bio)) > 1000 {
		return entity.User{}, apperrors.ErrInvalidArgument
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
