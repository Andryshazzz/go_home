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
	Secret            string
	AccessExpiration  time.Duration
	RefreshExpiration time.Duration
}

type Claims struct {
	TokenType TokenType `json:"typ"`
	jwtlib.RegisteredClaims
}

type TokenPair struct {
	AccessToken     string `json:"access_token"`
	RefreshToken    string `json:"refresh_token"`
	AccessExpiresIn int64  `json:"access_expires_in"`
}

type JWTService struct {
	secret            []byte
	accessExpiration  time.Duration
	refreshExpiration time.Duration
}

func NewJWTService(cfg Config) (*JWTService, error) {
	if len(cfg.Secret) < minSecretLen {
		return nil, fmt.Errorf("%w: secret must be at least %d bytes", ErrInvalidConfig, minSecretLen)
	}

	if cfg.AccessExpiration <= 0 || cfg.RefreshExpiration <= 0 {
		return nil, fmt.Errorf("%w: expiration must be greater than zero", ErrInvalidConfig)
	}

	return &JWTService{
		secret:            []byte(cfg.Secret),
		accessExpiration:  cfg.AccessExpiration,
		refreshExpiration: cfg.RefreshExpiration,
	}, nil
}

func (s *JWTService) GenerateTokenPair(subject string) (*TokenPair, error) {
	access, err := s.generate(subject, TypeAccess, s.accessExpiration)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refresh, err := s.generate(subject, TypeRefresh, s.refreshExpiration)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:     access,
		RefreshToken:    refresh,
		AccessExpiresIn: int64(s.accessExpiration.Seconds()),
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

func (s *JWTService) generate(subject string, typ TokenType, ttl time.Duration) (string, error) {
	id, err := newID()
	if err != nil {
		return "", err
	}

	now := time.Now()
	claims := Claims{
		TokenType: typ,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ID:        id,
			Subject:   subject,
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(ttl)),
		},
	}

	return jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims).SignedString(s.secret)
}

func (s *JWTService) parse(tokenString string, want TokenType) (*Claims, error) {
	claims := &Claims{}

	_, err := jwtlib.ParseWithClaims(
		tokenString,
		claims,
		func(*jwtlib.Token) (any, error) { return s.secret, nil },
		jwtlib.WithValidMethods([]string{jwtlib.SigningMethodHS256.Alg()}),
		jwtlib.WithExpirationRequired(),
	)
	switch {
	case errors.Is(err, jwtlib.ErrTokenExpired):
		return nil, ErrTokenExpired
	case err != nil:
		return nil, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	if claims.TokenType != want {
		return nil, fmt.Errorf("%w: unexpected token type %q", ErrInvalidToken, claims.TokenType)
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
