package profileservice

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"go_home/internal/domain/entity"
	apperrors "go_home/pkg/errors"

	"github.com/google/uuid"
)

type fakeRepo struct {
	updateErr   error
	updated     *entity.UpdateProfileRequest
	activeUsers []entity.ActiveUser
	activeErr   error
	gotMin      int
	gotActive   entity.ListActiveUsersRequest
}

type fakeURLs struct{ err error }

func newProfileService(repo ProfileRepository) *ProfileService {
	return NewProfileService(repo, fakeURLs{})
}

func (r *fakeRepo) GetUserByID(context.Context, uuid.UUID) (entity.User, error) {
	return entity.User{}, nil
}

func (r *fakeRepo) ListUsers(context.Context, entity.ListUsersRequest) ([]entity.User, int, error) {
	return nil, 0, nil
}

func (r *fakeRepo) UpdateUser(_ context.Context, req entity.UpdateProfileRequest) (entity.User, error) {
	if r.updateErr != nil {
		return entity.User{}, r.updateErr
	}
	r.updated = &req

	return entity.User{ID: req.UserID}, nil
}

func (r *fakeRepo) DeleteUser(context.Context, uuid.UUID) error { return nil }

func (r *fakeRepo) ListActiveUsers(_ context.Context, req entity.ListActiveUsersRequest, minAvatars int) ([]entity.ActiveUser, error) {
	r.gotActive, r.gotMin = req, minAvatars

	return r.activeUsers, r.activeErr
}

func (f fakeURLs) PresignedGetURL(_ context.Context, key string, _ time.Duration) (string, error) {
	if f.err != nil {
		return "", f.err
	}

	return "http://storage.test/" + key, nil
}

func ptr[T any](v T) *T { return &v }

func TestUpdateMyProfile_Validation(t *testing.T) {
	tests := []struct {
		name string
		req  entity.UpdateProfileRequest
	}{
		{"nothing to update", entity.UpdateProfileRequest{}},
		{"invalid email", entity.UpdateProfileRequest{Email: ptr("invalid-email")}},
		{"login too short", entity.UpdateProfileRequest{Login: ptr("x")}},
		{"blank login", entity.UpdateProfileRequest{Login: ptr("   ")}},
		{"login too long", entity.UpdateProfileRequest{Login: ptr(strings.Repeat("a", 65))}},
		{"age zero", entity.UpdateProfileRequest{Age: ptr(0)}},
		{"age too big", entity.UpdateProfileRequest{Age: ptr(121)}},
		{"bio too long", entity.UpdateProfileRequest{Bio: ptr(strings.Repeat("a", 1001))}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepo{}
			s := newProfileService(repo)

			tt.req.UserID = uuid.New()

			_, err := s.UpdateMyProfile(context.Background(), tt.req)
			if !errors.Is(err, apperrors.ErrInvalidArgument) {
				t.Fatalf("expected ErrInvalidArgument, got %v", err)
			}
			if repo.updated != nil {
				t.Error("repository must not be called with invalid data")
			}
		})
	}
}

func TestUpdateMyProfile_PassesNormalizedFields(t *testing.T) {
	repo := &fakeRepo{}
	s := newProfileService(repo)
	id := uuid.New()

	_, err := s.UpdateMyProfile(context.Background(), entity.UpdateProfileRequest{
		UserID: id,
		Login:  ptr("  erin  "),
		Email:  ptr(" erin@example.com "),
		Bio:    ptr(""),
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	got := repo.updated
	if got == nil {
		t.Fatal("repository was not called")
	}
	if got.UserID != id {
		t.Error("user id must be passed through")
	}
	if got.Login == nil || *got.Login != "erin" {
		t.Errorf("login must be trimmed, got %v", got.Login)
	}
	if got.Email == nil || *got.Email != "erin@example.com" {
		t.Errorf("email must be trimmed, got %v", got.Email)
	}
	if got.Bio == nil || *got.Bio != "" {
		t.Errorf("empty bio must reach the repository as an empty string, got %v", got.Bio)
	}
	if got.Age != nil {
		t.Error("age was not sent and must stay nil")
	}
}

func TestUpdateMyProfile_RepositoryErrors(t *testing.T) {
	tests := []struct {
		name string
		repo error
		want error
	}{
		{"login or email taken", apperrors.ErrUserAlreadyExists, apperrors.ErrUserAlreadyExists},
		{"user not found", apperrors.ErrNotFoundUser, apperrors.ErrNotFoundUser},
		{"unexpected failure", errors.New("db down"), apperrors.ErrInternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newProfileService(&fakeRepo{updateErr: tt.repo})

			_, err := s.UpdateMyProfile(context.Background(), entity.UpdateProfileRequest{
				UserID: uuid.New(),
				Bio:    ptr("new bio"),
			})
			if !errors.Is(err, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, err)
			}
		})
	}
}

func TestListActiveUsers(t *testing.T) {
	valid := entity.ListActiveUsersRequest{MinAge: 18, MaxAge: 60, Limit: 10}

	t.Run("passes the rule threshold and builds links", func(t *testing.T) {
		repo := &fakeRepo{activeUsers: []entity.ActiveUser{
			{User: entity.User{ID: uuid.New()}, ActiveAvatars: 4, LastAvatar: entity.Avatar{ID: uuid.New(), ObjectKey: "a.jpg"}},
		}}

		users, err := newProfileService(repo).ListActiveUsers(context.Background(), valid)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if repo.gotMin != 3 {
			t.Errorf("the threshold must be 3 active avatars, got %d", repo.gotMin)
		}
		if users[0].LastAvatar.URL != "http://storage.test/a.jpg" {
			t.Errorf("url must be built from the object key, got %q", users[0].LastAvatar.URL)
		}
	})

	t.Run("default limit", func(t *testing.T) {
		repo := &fakeRepo{}

		if _, err := newProfileService(repo).ListActiveUsers(context.Background(),
			entity.ListActiveUsersRequest{MinAge: 18, MaxAge: 60}); err != nil {
			t.Fatalf("list: %v", err)
		}
		if repo.gotActive.Limit != defaultLimit {
			t.Errorf("expected limit %d, got %d", defaultLimit, repo.gotActive.Limit)
		}
	})

	t.Run("link failure keeps the user", func(t *testing.T) {
		repo := &fakeRepo{activeUsers: []entity.ActiveUser{
			{LastAvatar: entity.Avatar{ObjectKey: "a.jpg"}},
		}}
		s := NewProfileService(repo, fakeURLs{err: errors.New("presign failed")})

		users, err := s.ListActiveUsers(context.Background(), valid)
		if err != nil || len(users) != 1 || users[0].LastAvatar.URL != "" {
			t.Fatalf("expected one user with an empty url, got %+v, %v", users, err)
		}
	})

	invalid := map[string]entity.ListActiveUsersRequest{
		"min above max":    {MinAge: 40, MaxAge: 20, Limit: 10},
		"age out of range": {MinAge: 0, MaxAge: 200, Limit: 10},
		"limit above max":  {MinAge: 18, MaxAge: 60, Limit: maxLimit + 1},
		"negative offset":  {MinAge: 18, MaxAge: 60, Limit: 10, Offset: -1},
	}
	for name, req := range invalid {
		t.Run(name, func(t *testing.T) {
			repo := &fakeRepo{}

			_, err := newProfileService(repo).ListActiveUsers(context.Background(), req)
			if !errors.Is(err, apperrors.ErrInvalidArgument) {
				t.Fatalf("expected ErrInvalidArgument, got %v", err)
			}
			if repo.gotMin != 0 {
				t.Error("the repository must not be called with invalid input")
			}
		})
	}

	t.Run("repository failure", func(t *testing.T) {
		repo := &fakeRepo{activeErr: errors.New("db down")}

		if _, err := newProfileService(repo).ListActiveUsers(context.Background(), valid); !errors.Is(err, apperrors.ErrInternal) {
			t.Fatalf("expected ErrInternal, got %v", err)
		}
	})
}
