package authservice

import (
	"context"
	entity "go_home/internal/domain/entity"
	apperrors "go_home/pkg/errors"
)

func (s *AuthService) Register(
	ctx context.Context,
	req entity.RegisterUserRequest,
) (*entity.RegisterUserResponse, error) {
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

	passwordHash, err := hashPassword(req.Password)
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
		return nil, apperrors.ErrCreateUser
	}

	tokenPair, err := s.jwtService.GenerateTokenPair(
		createdUser.ID.String(),
	)
	if err != nil {
		return nil, apperrors.ErrGenerateTokens
	}

	return &entity.RegisterUserResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}
