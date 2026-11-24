package authservice

import (
	"context"
	"errors"
	"fmt"

	"go_home/internal/domain/entity"
	apperrors "go_home/pkg/errors"

	"github.com/google/uuid"
)

func (s *AuthService) Refresh(
	ctx context.Context,
	req entity.RefreshTokenRequest,
) (*entity.TokenPairResponse, error) {
	claims, err := s.jwtService.ParseRefresh(req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", apperrors.ErrInvalidToken, err)
	}

	sessionID, err := uuid.Parse(claims.SessionID)
	if err != nil {
		return nil, apperrors.ErrInvalidToken
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, apperrors.ErrInvalidToken
	}

	pair, err := s.jwtService.GenerateTokenPair(claims.Subject, claims.SessionID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", apperrors.ErrGenerateTokens, err)
	}

	if err := s.authRepository.RotateSession(ctx, entity.RotateSessionRequest{
		SessionID:    sessionID,
		UserID:       userID,
		OldTokenHash: hashToken(req.RefreshToken),
		NewTokenHash: hashToken(pair.RefreshToken),
		ExpiresAt:    pair.RefreshExpiresAt,
	}); err != nil {
		if errors.Is(err, apperrors.ErrInvalidToken) {
			return nil, err
		}

		return nil, fmt.Errorf("%w: rotate session: %w", apperrors.ErrInternal, err)
	}

	return &entity.TokenPairResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	}, nil
}
