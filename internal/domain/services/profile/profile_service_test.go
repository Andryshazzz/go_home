package profileservice

import (
	"context"
	"errors"
	"strings"
	"testing"

	"go_home/internal/domain/entity"
	apperrors "go_home/pkg/errors"

	"github.com/google/uuid"
)

type fakeRepo struct {
	updateErr error
	updated   *entity.UpdateProfileRequest
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
			s := NewProfileService(repo)

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
	s := NewProfileService(repo)
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
			s := NewProfileService(&fakeRepo{updateErr: tt.repo})

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
