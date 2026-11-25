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
	"go_home/pkg/password"
	"go_home/pkg/tokenhash"

	"github.com/google/uuid"
)

const testPassword = "supersecret123"

type fakeRepo struct {
	users    map[string]entity.User
	sessions map[uuid.UUID]entity.Session

	createUserErr error
	getUserErr    error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		users:    map[string]entity.User{},
		sessions: map[uuid.UUID]entity.Session{},
	}
}

func (r *fakeRepo) CreateUserWithSession(_ context.Context, u entity.User, s entity.Session) error {
	if r.createUserErr != nil {
		return r.createUserErr
	}
	if _, exists := r.users[u.Login]; exists {
		return apperrors.ErrUserAlreadyExists
	}

	r.users[u.Login] = u
	r.sessions[s.ID] = s

	return nil
}

func (r *fakeRepo) GetUserByLogin(_ context.Context, login string) (entity.User, error) {
	if r.getUserErr != nil {
		return entity.User{}, r.getUserErr
	}
	u, ok := r.users[login]
	if !ok {
		return entity.User{}, apperrors.ErrNotFoundUser
	}

	return u, nil
}

func (r *fakeRepo) CreateSession(_ context.Context, s entity.Session) error {
	r.sessions[s.ID] = s

	return nil
}

func (r *fakeRepo) GetSession(_ context.Context, id uuid.UUID) (entity.Session, error) {
	s, ok := r.sessions[id]
	if !ok {
		return entity.Session{}, apperrors.ErrInvalidToken
	}

	return s, nil
}

func (r *fakeRepo) DeleteSession(_ context.Context, id uuid.UUID) error {
	delete(r.sessions, id)

	return nil
}

func (r *fakeRepo) RotateSession(_ context.Context, req entity.RotateSessionRequest) error {
	s, ok := r.sessions[req.SessionID]
	if !ok || s.UserID != req.UserID || s.TokenHash != req.OldTokenHash || !s.ExpiresAt.After(time.Now()) {
		return apperrors.ErrInvalidToken
	}

	s.TokenHash = req.NewTokenHash
	s.ExpiresAt = req.ExpiresAt
	r.sessions[req.SessionID] = s

	return nil
}

func newService(t *testing.T) (*AuthService, *fakeRepo) {
	t.Helper()

	jwtService, err := jwt.NewJWTService(jwt.Config{
		AccessSecret:      strings.Repeat("a", 32),
		RefreshSecret:     strings.Repeat("r", 32),
		AccessExpiration:  time.Minute,
		RefreshExpiration: time.Hour,
	})
	if err != nil {
		t.Fatalf("new jwt service: %v", err)
	}

	repo := newFakeRepo()

	return NewAuthService(repo, jwtService), repo
}

func validRegister() entity.RegisterUserRequest {
	return entity.RegisterUserRequest{
		Login:    "bobby",
		Email:    "bob@example.com",
		Password: testPassword,
		Age:      30,
		Bio:      "hi",
	}
}

func mustRegister(t *testing.T, s *AuthService) *entity.TokenPairResponse {
	t.Helper()

	resp, err := s.Register(context.Background(), validRegister())
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	return resp
}

func TestRegister_Success(t *testing.T) {
	s, repo := newService(t)

	resp := mustRegister(t, s)

	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Fatal("both tokens must be returned")
	}

	user := repo.users["bobby"]
	if user.PasswordHash == testPassword {
		t.Fatal("password must not be stored in plain text")
	}
	if err := password.Compare(user.PasswordHash, testPassword); err != nil {
		t.Errorf("stored hash does not match the password: %v", err)
	}

	if len(repo.sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(repo.sessions))
	}
	for _, session := range repo.sessions {
		if session.UserID != user.ID {
			t.Error("session must belong to the created user (check that NewUser sets the ID)")
		}
		if session.TokenHash != tokenhash.Hash(resp.RefreshToken) {
			t.Error("session must hold the hash of the refresh token, not the token itself")
		}
	}
}

func TestRegister_Errors(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*fakeRepo)
		req   func(*entity.RegisterUserRequest)
		want  error
	}{
		{"invalid email", nil, func(r *entity.RegisterUserRequest) { r.Email = "invalid-email" }, apperrors.ErrInvalidArgument},
		{"login too short", nil, func(r *entity.RegisterUserRequest) { r.Login = "x" }, apperrors.ErrInvalidArgument},
		{"password too short", nil, func(r *entity.RegisterUserRequest) { r.Password = "1234567" }, apperrors.ErrInvalidArgument},
		{"password too long", nil, func(r *entity.RegisterUserRequest) { r.Password = strings.Repeat("a", 73) }, apperrors.ErrInvalidArgument},
		{"age out of range", nil, func(r *entity.RegisterUserRequest) { r.Age = 121 }, apperrors.ErrInvalidArgument},
		{"bio too long", nil, func(r *entity.RegisterUserRequest) { r.Bio = strings.Repeat("a", 1001) }, apperrors.ErrInvalidArgument},
		{
			"login already taken",
			func(r *fakeRepo) { r.users["bobby"] = entity.User{Login: "bobby"} },
			nil,
			apperrors.ErrUserAlreadyExists,
		},
		{
			"repository failure",
			func(r *fakeRepo) { r.createUserErr = errors.New("db down") },
			nil,
			apperrors.ErrCreateUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, repo := newService(t)
			if tt.setup != nil {
				tt.setup(repo)
			}
			usersBefore := len(repo.users)

			req := validRegister()
			if tt.req != nil {
				tt.req(&req)
			}

			_, err := s.Register(context.Background(), req)
			if !errors.Is(err, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, err)
			}

			if len(repo.sessions) != 0 {
				t.Error("no session must be created on error")
			}
			if len(repo.users) != usersBefore {
				t.Error("no user must be created on error")
			}
		})
	}
}

func TestLogin_Success(t *testing.T) {
	s, repo := newService(t)
	mustRegister(t, s)

	resp, err := s.Login(context.Background(), entity.LoginUserRequest{Login: "bobby", Password: testPassword})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Error("both tokens must be returned")
	}
	if resp.User.Login != "bobby" {
		t.Errorf("unexpected user %q", resp.User.Login)
	}
	if len(repo.sessions) != 2 {
		t.Errorf("each login must create its own session, got %d sessions in total", len(repo.sessions))
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	s, repo := newService(t)
	mustRegister(t, s)

	tests := []struct {
		name string
		req  entity.LoginUserRequest
	}{
		{"wrong password", entity.LoginUserRequest{Login: "bobby", Password: "wrongpassword"}},
		{"unknown login gives the same error", entity.LoginUserRequest{Login: "nobody", Password: testPassword}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := len(repo.sessions)

			_, err := s.Login(context.Background(), tt.req)
			if !errors.Is(err, apperrors.ErrInvalidCredentials) {
				t.Fatalf("expected ErrInvalidCredentials, got %v", err)
			}
			if len(repo.sessions) != before {
				t.Error("a failed login must not create a session")
			}
		})
	}

	t.Run("db failure is not reported as wrong credentials", func(t *testing.T) {
		repo.getUserErr = errors.New("db down")
		defer func() { repo.getUserErr = nil }()

		_, err := s.Login(context.Background(), entity.LoginUserRequest{Login: "bobby", Password: testPassword})
		if !errors.Is(err, apperrors.ErrInternal) || errors.Is(err, apperrors.ErrInvalidCredentials) {
			t.Fatalf("expected ErrInternal, got %v", err)
		}
	})
}

func TestRefresh_RotationAndReuse(t *testing.T) {
	s, repo := newService(t)
	old := mustRegister(t, s)

	fresh, err := s.Refresh(context.Background(), entity.RefreshTokenRequest{RefreshToken: old.RefreshToken})
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if fresh.RefreshToken == old.RefreshToken {
		t.Fatal("refresh token must be rotated")
	}
	if len(repo.sessions) != 1 {
		t.Fatalf("session must be replaced in place, got %d", len(repo.sessions))
	}
	for _, session := range repo.sessions {
		if session.TokenHash != tokenhash.Hash(fresh.RefreshToken) {
			t.Error("session must hold the hash of the new refresh token")
		}
	}

	_, err = s.Refresh(context.Background(), entity.RefreshTokenRequest{RefreshToken: old.RefreshToken})
	if !errors.Is(err, apperrors.ErrInvalidToken) {
		t.Fatalf("reused token: expected ErrInvalidToken, got %v", err)
	}

	if _, err := s.Refresh(context.Background(), entity.RefreshTokenRequest{RefreshToken: fresh.RefreshToken}); err != nil {
		t.Fatalf("new token must work: %v", err)
	}
}

func TestRefresh_InvalidTokens(t *testing.T) {
	s, _ := newService(t)
	resp := mustRegister(t, s)

	tests := []struct {
		name  string
		token string
	}{
		{"garbage", "abc.def.ghi"},
		{"empty", ""},
		{"access token instead of refresh", resp.AccessToken},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := s.Refresh(context.Background(), entity.RefreshTokenRequest{RefreshToken: tt.token})
			if !errors.Is(err, apperrors.ErrInvalidToken) {
				t.Fatalf("expected ErrInvalidToken, got %v", err)
			}
		})
	}
}

func TestTokens_SecretsAreSeparate(t *testing.T) {
	s, _ := newService(t)
	resp := mustRegister(t, s)

	if _, err := s.jwtService.ParseAccess(resp.RefreshToken); err == nil {
		t.Error("refresh token must not pass as an access token")
	}
	if _, err := s.jwtService.ParseRefresh(resp.AccessToken); err == nil {
		t.Error("access token must not pass as a refresh token")
	}
}

func TestLogout(t *testing.T) {
	t.Run("deletes the session and cannot be repeated", func(t *testing.T) {
		s, repo := newService(t)
		resp := mustRegister(t, s)

		if err := s.Logout(context.Background(), entity.LogoutUserRequest{RefreshToken: resp.RefreshToken}); err != nil {
			t.Fatalf("logout: %v", err)
		}
		if len(repo.sessions) != 0 {
			t.Fatalf("session must be deleted, got %d", len(repo.sessions))
		}

		err := s.Logout(context.Background(), entity.LogoutUserRequest{RefreshToken: resp.RefreshToken})
		if !errors.Is(err, apperrors.ErrInvalidToken) {
			t.Fatalf("second logout: expected ErrInvalidToken, got %v", err)
		}

		_, err = s.Refresh(context.Background(), entity.RefreshTokenRequest{RefreshToken: resp.RefreshToken})
		if !errors.Is(err, apperrors.ErrInvalidToken) {
			t.Fatalf("refresh after logout: expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("stale token after rotation is a mismatch and keeps the session", func(t *testing.T) {
		s, repo := newService(t)
		old := mustRegister(t, s)

		fresh, err := s.Refresh(context.Background(), entity.RefreshTokenRequest{RefreshToken: old.RefreshToken})
		if err != nil {
			t.Fatalf("refresh: %v", err)
		}

		err = s.Logout(context.Background(), entity.LogoutUserRequest{RefreshToken: old.RefreshToken})
		if !errors.Is(err, apperrors.ErrTokenMismatch) {
			t.Fatalf("expected ErrTokenMismatch, got %v", err)
		}
		if len(repo.sessions) != 1 {
			t.Fatal("session must survive a logout with a stale token")
		}

		if err := s.Logout(context.Background(), entity.LogoutUserRequest{RefreshToken: fresh.RefreshToken}); err != nil {
			t.Fatalf("logout with the current token: %v", err)
		}
	})

	t.Run("deletes only the session of the given token", func(t *testing.T) {
		s, repo := newService(t)
		first := mustRegister(t, s)

		second, err := s.Login(context.Background(), entity.LoginUserRequest{Login: "bobby", Password: testPassword})
		if err != nil {
			t.Fatalf("login: %v", err)
		}

		if err := s.Logout(context.Background(), entity.LogoutUserRequest{RefreshToken: first.RefreshToken}); err != nil {
			t.Fatalf("logout: %v", err)
		}
		if len(repo.sessions) != 1 {
			t.Fatalf("the other session must survive, got %d", len(repo.sessions))
		}

		if _, err := s.Refresh(context.Background(), entity.RefreshTokenRequest{RefreshToken: second.RefreshToken}); err != nil {
			t.Fatalf("the other session must keep working: %v", err)
		}
	})

	t.Run("garbage token", func(t *testing.T) {
		s, _ := newService(t)

		err := s.Logout(context.Background(), entity.LogoutUserRequest{RefreshToken: "abc.def.ghi"})
		if !errors.Is(err, apperrors.ErrInvalidToken) {
			t.Fatalf("expected ErrInvalidToken, got %v", err)
		}
	})
}
