package authservice

import (
	"context"
	"errors"
	"fmt"
	entity "go_home/internal/domain/entity"
	apperrors "go_home/pkg/errors"
	"go_home/pkg/password"
)

func (s *AuthService) Register(
	ctx context.Context,
	req entity.RegisterUserRequest,
) (*entity.TokenPairResponse, error) {
	if req.Login == "" {
		return nil, apperrors.ErrInvalidArgument
	}
	if req.Email == "" {
		return nil, apperrors.ErrInvalidArgument
	}
	if len(req.Password) < 8 || len(req.Password) > 72 {
		return nil, apperrors.ErrInvalidArgument
	}
	if req.Age <= 0 || req.Age > 120 {
		return nil, apperrors.ErrInvalidArgument
	}
	if len([]rune(req.Bio)) > 1000 {
		return nil, apperrors.ErrInvalidArgument
	}

	passwordHash, err := password.Hash(req.Password)
	if err != nil {
		return nil, apperrors.ErrHashPassword
	}

	user := entity.NewUser(
		req.Login,
		req.Email,
		req.Age,
		string(passwordHash),
		req.Bio,
	)

	createdUser, err := s.authRepository.CreateUser(ctx, user)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserAlreadyExists) {
			return nil, err
		}

		return nil, fmt.Errorf("%w: %w", apperrors.ErrCreateUser, err)
	}

	tokenPair, err := s.issueTokens(ctx, createdUser.ID)
	if err != nil {
		return nil, err
	}

	return &entity.TokenPairResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}
