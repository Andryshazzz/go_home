package profileservice

import (
	"context"
	"errors"
	"fmt"
	"go_home/internal/domain/entity"
	apperrors "go_home/pkg/errors"
	"strings"
	"unicode/utf8"
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
		if utf8.RuneCountInString(login) < 3 {
			return entity.User{}, apperrors.ErrInvalidArgument
		}
		req.Login = &login
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
