package authservice

import (
	"context"
	"errors"
	"fmt"
	"go_home/internal/domain/entity"
	"go_home/internal/domain/validation"
	apperrors "go_home/pkg/errors"
	"go_home/pkg/password"
)

func (s *AuthService) Register(
	ctx context.Context,
	req entity.RegisterUserRequest,
) (*entity.TokenPairResponse, error) {
	login, err := validation.Login(req.Login)
	if err != nil {
		return nil, err
	}

	email, err := validation.Email(req.Email)
	if err != nil {
		return nil, err
	}

	if err := validation.Password(req.Password); err != nil {
		return nil, err
	}

	if err := validation.Age(req.Age); err != nil {
		return nil, err
	}

	if err := validation.Bio(req.Bio); err != nil {
		return nil, err
	}

	passwordHash, err := password.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", apperrors.ErrHashPassword, err)
	}

	user := entity.NewUser(login, email, req.Age, passwordHash, req.Bio)

	pair, session, err := s.generateTokens(user.ID)
	if err != nil {
		return nil, err
	}

	if err := s.authRepository.CreateUserWithSession(ctx, user, session); err != nil {
		if errors.Is(err, apperrors.ErrUserAlreadyExists) {
			return nil, err
		}

		return nil, fmt.Errorf("%w: %w", apperrors.ErrCreateUser, err)
	}

	return &entity.TokenPairResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	}, nil
}
