package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

type TokenType string

const (
	TypeAccess  TokenType = "access"
	TypeRefresh TokenType = "refresh"
)

const minSecretLen = 32

var (
	ErrTokenExpired  = errors.New("token expired")
	ErrInvalidToken  = errors.New("invalid token")
	ErrInvalidConfig = errors.New("invalid jwt config")
)

type Config struct {
	AccessSecret      string
	RefreshSecret     string
	AccessExpiration  time.Duration
	RefreshExpiration time.Duration
}

type Claims struct {
	TokenType TokenType `json:"typ"`
	SessionID string    `json:"sid"`
	jwtlib.RegisteredClaims
}

type TokenPair struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	AccessExpiresIn  int64     `json:"access_expires_in"`
	SessionID        string    `json:"session_id"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

type JWTService struct {
	accessSecret      []byte
	refreshSecret     []byte
	accessExpiration  time.Duration
	refreshExpiration time.Duration
}

func NewJWTService(cfg Config) (*JWTService, error) {
	if len(cfg.AccessSecret) < minSecretLen {
		return nil, fmt.Errorf(
			"%w: access secret must be at least %d bytes",
			ErrInvalidConfig,
			minSecretLen,
		)
	}

	if len(cfg.RefreshSecret) < minSecretLen {
		return nil, fmt.Errorf(
			"%w: refresh secret must be at least %d bytes",
			ErrInvalidConfig,
			minSecretLen,
		)
	}

	if cfg.AccessExpiration <= 0 || cfg.RefreshExpiration <= 0 {
		return nil, fmt.Errorf(
			"%w: expiration must be greater than zero",
			ErrInvalidConfig,
		)
	}

	return &JWTService{
		accessSecret:      []byte(cfg.AccessSecret),
		refreshSecret:     []byte(cfg.RefreshSecret),
		accessExpiration:  cfg.AccessExpiration,
		refreshExpiration: cfg.RefreshExpiration,
	}, nil
}

func (s *JWTService) GenerateTokenPair(subject, sessionID string) (*TokenPair, error) {
	access, err := s.generate(subject, sessionID, TypeAccess, s.accessExpiration)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refresh, err := s.generate(subject, sessionID, TypeRefresh, s.refreshExpiration)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:      access,
		RefreshToken:     refresh,
		AccessExpiresIn:  int64(s.accessExpiration.Seconds()),
		SessionID:        sessionID,
		RefreshExpiresAt: time.Now().Add(s.refreshExpiration),
	}, nil
}

func (s *JWTService) ParseAccess(token string) (*Claims, error) {
	return s.parse(token, TypeAccess)
}

func (s *JWTService) ParseRefresh(token string) (*Claims, error) {
	return s.parse(token, TypeRefresh)
}

func (s *JWTService) AccessExpiration() time.Duration  { return s.accessExpiration }
func (s *JWTService) RefreshExpiration() time.Duration { return s.refreshExpiration }

func (s *JWTService) generate(
	subject, sessionID string,
	typ TokenType,
	ttl time.Duration,
) (string, error) {
	id, err := newID()
	if err != nil {
		return "", err
	}

	now := time.Now()

	claims := Claims{
		TokenType: typ,
		SessionID: sessionID,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ID:        id,
			Subject:   subject,
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(ttl)),
		},
	}

	secret := s.accessSecret
	if typ == TypeRefresh {
		secret = s.refreshSecret
	}

	return jwtlib.NewWithClaims(
		jwtlib.SigningMethodHS256,
		claims,
	).SignedString(secret)
}

func (s *JWTService) parse(
	tokenString string,
	want TokenType,
) (*Claims, error) {
	claims := &Claims{}

	secret := s.accessSecret
	if want == TypeRefresh {
		secret = s.refreshSecret
	}

	_, err := jwtlib.ParseWithClaims(
		tokenString,
		claims,
		func(*jwtlib.Token) (any, error) {
			return secret, nil
		},
		jwtlib.WithValidMethods(
			[]string{jwtlib.SigningMethodHS256.Alg()},
		),
		jwtlib.WithExpirationRequired(),
	)

	switch {
	case errors.Is(err, jwtlib.ErrTokenExpired):
		return nil, ErrTokenExpired
	case err != nil:
		return nil, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	if claims.TokenType != want {
		return nil, fmt.Errorf(
			"%w: unexpected token type %q",
			ErrInvalidToken,
			claims.TokenType,
		)
	}

	if claims.SessionID == "" {
		return nil, fmt.Errorf("%w: missing session id", ErrInvalidToken)
	}

	return claims, nil
}
func newID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token id: %w", err)
	}

	return hex.EncodeToString(b), nil
}
