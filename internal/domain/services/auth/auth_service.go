package authservice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go_home/internal/domain/entity"
	apperrors "go_home/pkg/errors"
	"go_home/pkg/jwt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthRepository interface {
	CreateUser(ctx context.Context, user entity.User) (entity.User, error)
	GetUserByLogin(ctx context.Context, login string) (entity.User, error)
	CreateSession(ctx context.Context, session entity.Session) error
	GetSession(ctx context.Context, id uuid.UUID) (entity.Session, error)
	DeleteSession(ctx context.Context, id uuid.UUID) error
	RotateSession(ctx context.Context, req entity.RotateSessionRequest) error
}

type AuthService struct {
	authRepository AuthRepository
	jwtService     *jwt.JWTService
}

func NewAuthService(authRepository AuthRepository, jwtService *jwt.JWTService) *AuthService {
	return &AuthService{
		authRepository: authRepository,
		jwtService:     jwtService,
	}
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("%w: %w", apperrors.ErrHashPassword, err)
	}

	return string(bytes), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))

	return hex.EncodeToString(sum[:])
}

func (s *AuthService) issueTokens(ctx context.Context, userID uuid.UUID) (*jwt.TokenPair, error) {
	sessionID := uuid.New()

	pair, err := s.jwtService.GenerateTokenPair(userID.String(), sessionID.String())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", apperrors.ErrGenerateTokens, err)
	}

	if err := s.authRepository.CreateSession(ctx, entity.Session{
		ID:        sessionID,
		UserID:    userID,
		TokenHash: hashToken(pair.RefreshToken),
		ExpiresAt: pair.RefreshExpiresAt,
	}); err != nil {
		return nil, fmt.Errorf("%w: create session: %w", apperrors.ErrInternal, err)
	}

	return pair, nil
}
