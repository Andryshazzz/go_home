package authservice

import (
	"context"
	"fmt"
	"go_home/internal/domain/entity"
	apperrors "go_home/pkg/errors"
	"go_home/pkg/jwt"
	"go_home/pkg/tokenhash"

	"github.com/google/uuid"
)

type AuthRepository interface {
	CreateUserWithSession(ctx context.Context, user entity.User, session entity.Session) error
	GetUserByLogin(ctx context.Context, login string) (entity.User, error)
	CreateSession(ctx context.Context, session entity.Session) error
	GetSession(ctx context.Context, id uuid.UUID) (entity.Session, error)
	DeleteSession(ctx context.Context, id uuid.UUID) error
	RotateSession(ctx context.Context, req entity.RotateSessionRequest) error
}

type AuthService struct {
	authRepository AuthRepository
	jwtService     *jwt.Service
}

func NewAuthService(authRepository AuthRepository, jwtService *jwt.Service) *AuthService {
	return &AuthService{
		authRepository: authRepository,
		jwtService:     jwtService,
	}
}

func (s *AuthService) issueTokens(ctx context.Context, userID uuid.UUID) (*jwt.TokenPair, error) {
	pair, session, err := s.generateTokens(userID)
	if err != nil {
		return nil, err
	}

	if err := s.authRepository.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("%w: create session: %w", apperrors.ErrInternal, err)
	}

	return pair, nil
}

func (s *AuthService) generateTokens(userID uuid.UUID) (*jwt.TokenPair, entity.Session, error) {
	sessionID := uuid.New()

	pair, err := s.jwtService.GenerateTokenPair(userID.String(), sessionID.String())
	if err != nil {
		return nil, entity.Session{}, fmt.Errorf("%w: %w", apperrors.ErrGenerateTokens, err)
	}

	return pair, entity.Session{
		ID:        sessionID,
		UserID:    userID,
		TokenHash: tokenhash.Hash(pair.RefreshToken),
		ExpiresAt: pair.RefreshExpiresAt,
	}, nil
}
