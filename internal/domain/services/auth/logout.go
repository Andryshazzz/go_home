package authservice

import (
	"context"
	"errors"
	"fmt"
	"go_home/internal/domain/entity"
	apperrors "go_home/pkg/errors"
	"go_home/pkg/tokenhash"

	"github.com/google/uuid"
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

	if !tokenhash.Equal(session.TokenHash, req.RefreshToken) {
		return apperrors.ErrTokenMismatch
	}

	if err := s.authRepository.DeleteSession(ctx, sessionID); err != nil {
		return fmt.Errorf("%w: delete session: %w", apperrors.ErrInternal, err)
	}

	return nil
}
