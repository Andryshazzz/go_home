package authservice

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"go_home/internal/domain/entity"
	apperrors "go_home/pkg/errors"
	"go_home/pkg/jwt"
	"go_home/pkg/tokenhash"

	"github.com/google/uuid"
)

type registrationRepository struct {
	AuthRepository
	user    entity.User
	session entity.Session
	err     error
}

func (r *registrationRepository) CreateUserWithSession(_ context.Context, user entity.User, session entity.Session) error {
	r.user = user
	r.session = session
	return r.err
}

func (r *registrationRepository) CreateSession(_ context.Context, session entity.Session) error {
	r.session = session
	return r.err
}

func TestRegisterCreatesUserAndMatchingSession(t *testing.T) {
	jwtService, err := jwt.NewJWTService(jwt.Config{
		AccessSecret: strings.Repeat("a", 32), RefreshSecret: strings.Repeat("b", 32),
		AccessExpiration: time.Minute, RefreshExpiration: time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}
	repo := &registrationRepository{}
	service := NewAuthService(repo, jwtService)
	req := entity.RegisterUserRequest{Login: "user", Email: "user@example.com", Password: "password", Age: 20}

	pair, err := service.Register(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := jwtService.ParseRefresh(pair.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if repo.user.ID != repo.session.UserID || claims.Subject != repo.user.ID.String() ||
		claims.SessionID != repo.session.ID.String() || repo.session.TokenHash != tokenhash.Hash(pair.RefreshToken) {
		t.Fatal("registration returned tokens for a different user or session")
	}

	repo.err = apperrors.ErrUserAlreadyExists
	pair, err = service.Register(context.Background(), req)
	if pair != nil || !errors.Is(err, apperrors.ErrUserAlreadyExists) {
		t.Fatalf("expected conflict without tokens, got pair=%v err=%v", pair, err)
	}

	repo.err = errors.New("session insert failed")
	pair, err = service.Register(context.Background(), req)
	if pair != nil || !errors.Is(err, repo.err) {
		t.Fatalf("expected repository failure without tokens, got pair=%v err=%v", pair, err)
	}
}

func TestIssueTokensStillCreatesSession(t *testing.T) {
	jwtService, err := jwt.NewJWTService(jwt.Config{
		AccessSecret: strings.Repeat("a", 32), RefreshSecret: strings.Repeat("b", 32),
		AccessExpiration: time.Minute, RefreshExpiration: time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}
	repo := &registrationRepository{}
	userID := uuid.New()
	pair, err := NewAuthService(repo, jwtService).issueTokens(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := jwtService.ParseRefresh(pair.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if repo.session.UserID != userID || repo.session.ID.String() != claims.SessionID ||
		repo.session.TokenHash != tokenhash.Hash(pair.RefreshToken) {
		t.Fatal("issued tokens do not match the saved session")
	}
}
