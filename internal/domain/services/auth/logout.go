package authservice

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go_home/internal/domain/entity"
	apperrors "go_home/pkg/errors"
)

func (s *AuthService) Logout(ctx context.Context, req entity.LogoutUserRequest) error {
	claims, err := s.jwtService.ParseRefresh(req.RefreshToken)
	if err != nil {
		return fmt.Errorf("%w: %w", apperrors.ErrInvalidToken, err)
	}

	sessionID, err := uuid.Parse(claims.SessionID)
	if err != nil {
		return apperrors.ErrInvalidToken
	}

	session, err := s.authRepository.GetSession(ctx, sessionID)
	if err != nil {
		if errors.Is(err, apperrors.ErrInvalidToken) {
			return err
		}

		return fmt.Errorf("%w: get session: %w", apperrors.ErrInternal, err)
	}

	if subtle.ConstantTimeCompare([]byte(session.TokenHash), []byte(hashToken(req.RefreshToken))) != 1 {
		return apperrors.ErrTokenMismatch
	}

	if err := s.authRepository.DeleteSession(ctx, sessionID); err != nil {
		return fmt.Errorf("%w: delete session: %w", apperrors.ErrInternal, err)
	}

	return nil
}
