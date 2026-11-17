package authservice

import (
	"context"
	"errors"
	"fmt"
	"go_home/internal/domain/entity"
	apperrors "go_home/pkg/errors"

	"golang.org/x/crypto/bcrypt"
)

func (s *AuthService) Login(ctx context.Context, req entity.LoginUserRequest) (*entity.LoginUserResponse, error) {
	user, err := s.authRepository.GetUserByLogin(ctx, req.Login)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFoundUser) {
			return nil, apperrors.ErrInvalidCredentials
		}

		return nil, fmt.Errorf("%w: %w", apperrors.ErrInternal, err)
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(req.Password),
	); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, apperrors.ErrInvalidCredentials
		}

		return nil, fmt.Errorf("%w: compare password: %w", apperrors.ErrInternal, err)
	}

	tokenPair, err := s.jwtService.GenerateTokenPair(user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", apperrors.ErrGenerateTokens, err)
	}

	return &entity.LoginUserResponse{
		User:         user,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}
