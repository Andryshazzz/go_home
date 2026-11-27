package avatarservice

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"strings"
	"testing"
	"time"

	"go_home/internal/domain/entity"
	apperrors "go_home/pkg/errors"
	"go_home/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type fakeRepo struct {
	avatars map[uuid.UUID]entity.Avatar

	createErr error
	deleteErr error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{avatars: map[uuid.UUID]entity.Avatar{}}
}

func (r *fakeRepo) activeCount(userID uuid.UUID) int {
	n := 0
	for _, a := range r.avatars {
		if a.UserID == userID && a.DeletedAt == nil {
			n++
		}
	}

	return n
}

func (r *fakeRepo) CountActive(_ context.Context, userID uuid.UUID) (int, error) {
	return r.activeCount(userID), nil
}

func (r *fakeRepo) CreateAvatar(_ context.Context, a entity.Avatar, maxActive int) (entity.Avatar, error) {
	if r.createErr != nil {
		return entity.Avatar{}, r.createErr
	}
	if r.activeCount(a.UserID) >= maxActive {
		return entity.Avatar{}, apperrors.ErrAvatarLimitReached
	}

	a.CreatedAt = time.Now()
	r.avatars[a.ID] = a

	return a, nil
}

func (r *fakeRepo) SoftDeleteAvatar(_ context.Context, userID, avatarID uuid.UUID) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}

	a, ok := r.avatars[avatarID]
	if !ok || a.UserID != userID || a.DeletedAt != nil {
		return apperrors.ErrNotFoundAvatar
	}

	now := time.Now()
	a.DeletedAt = &now
	r.avatars[avatarID] = a

	return nil
}

func (r *fakeRepo) add(userID uuid.UUID, deleted bool) entity.Avatar {
	a := entity.Avatar{ID: uuid.New(), UserID: userID, ObjectKey: uuid.NewString() + ".jpg", CreatedAt: time.Now()}
	if deleted {
		now := time.Now()
		a.DeletedAt = &now
	}
	r.avatars[a.ID] = a

	return a
}

type fakeStorage struct {
	objects      map[string][]byte
	contentTypes map[string]string
	deletedKeys  []string

	putErr    error
	deleteErr error
	urlErr    error
}

func newFakeStorage() *fakeStorage {
	return &fakeStorage{objects: map[string][]byte{}, contentTypes: map[string]string{}}
}

func (s *fakeStorage) Put(_ context.Context, key string, r io.Reader, size int64, contentType string) error {
	if s.putErr != nil {
		return s.putErr
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	if int64(len(data)) != size {
		return errors.New("size does not match the stream length")
	}

	s.objects[key] = data
	s.contentTypes[key] = contentType

	return nil
}

func (s *fakeStorage) Delete(_ context.Context, key string) error {
	if s.deleteErr != nil {
		return s.deleteErr
	}

	delete(s.objects, key)
	s.deletedKeys = append(s.deletedKeys, key)

	return nil
}

func (s *fakeStorage) PresignedGetURL(_ context.Context, key string, _ time.Duration) (string, error) {
	if s.urlErr != nil {
		return "", s.urlErr
	}

	return "http://storage.test/" + key, nil
}

func newService(t *testing.T) (*AvatarService, *fakeRepo, *fakeStorage) {
	t.Helper()

	repo, storage := newFakeRepo(), newFakeStorage()
	log := &logger.Logger{Logger: zap.NewNop()}

	return NewAvatarService(repo, storage, log), repo, storage
}

func imageBytes(t *testing.T, format string) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 8, 8))

	var buf bytes.Buffer

	var err error
	switch format {
	case "png":
		err = png.Encode(&buf, img)
	case "jpeg":
		err = jpeg.Encode(&buf, img, nil)
	default:
		t.Fatalf("unknown format %q", format)
	}
	if err != nil {
		t.Fatalf("encode %s: %v", format, err)
	}

	buf.Write(bytes.Repeat([]byte{0xAB}, 2000))

	return buf.Bytes()
}

func uploadReq(userID uuid.UUID, data []byte) entity.UploadAvatarRequest {
	return entity.UploadAvatarRequest{UserID: userID, File: bytes.NewReader(data), Size: int64(len(data))}
}

func TestUpload_Success(t *testing.T) {
	tests := []struct {
		name, format, wantType, wantExt string
	}{
		{"jpeg", "jpeg", "image/jpeg", ".jpg"},
		{"png", "png", "image/png", ".png"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, repo, storage := newService(t)
			userID := uuid.New()
			data := imageBytes(t, tt.format)

			avatar, err := s.Upload(context.Background(), uploadReq(userID, data))
			if err != nil {
				t.Fatalf("upload: %v", err)
			}

			if avatar.UserID != userID || avatar.CreatedAt.IsZero() || avatar.URL == "" {
				t.Errorf("incomplete avatar: %+v", avatar)
			}

			if !strings.HasSuffix(avatar.ObjectKey, tt.wantExt) {
				t.Errorf("key %q must end with %s", avatar.ObjectKey, tt.wantExt)
			}
			if _, err := uuid.Parse(strings.TrimSuffix(avatar.ObjectKey, tt.wantExt)); err != nil {
				t.Errorf("key %q must be a UUID plus extension: %v", avatar.ObjectKey, err)
			}

			if !bytes.Equal(storage.objects[avatar.ObjectKey], data) {
				t.Error("stored bytes differ from the uploaded file (the reader was not rewound after type detection)")
			}
			if storage.contentTypes[avatar.ObjectKey] != tt.wantType {
				t.Errorf("content type %q, want %q", storage.contentTypes[avatar.ObjectKey], tt.wantType)
			}

			if repo.activeCount(userID) != 1 {
				t.Error("exactly one active avatar must be created")
			}
		})
	}
}

func TestUpload_Rejected(t *testing.T) {
	gif := append([]byte("GIF89a"), make([]byte, 200)...)

	tests := []struct {
		name string
		req  entity.UploadAvatarRequest
		want error
	}{
		{
			"text file with an image name",
			entity.UploadAvatarRequest{File: strings.NewReader("not an image, just text"), Size: 23},
			apperrors.ErrUnsupportedFileType,
		},
		{
			"gif is not allowed",
			entity.UploadAvatarRequest{File: bytes.NewReader(gif), Size: int64(len(gif))},
			apperrors.ErrUnsupportedFileType,
		},
		{
			"too large",
			entity.UploadAvatarRequest{File: strings.NewReader("x"), Size: maxFileSize + 1},
			apperrors.ErrFileTooLarge,
		},
		{
			"empty file",
			entity.UploadAvatarRequest{File: strings.NewReader(""), Size: 0},
			apperrors.ErrInvalidArgument,
		},
		{
			"no file",
			entity.UploadAvatarRequest{Size: 10},
			apperrors.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, repo, storage := newService(t)
			tt.req.UserID = uuid.New()

			_, err := s.Upload(context.Background(), tt.req)
			if !errors.Is(err, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, err)
			}
			if len(storage.objects) != 0 || len(repo.avatars) != 0 {
				t.Error("a rejected file must reach neither the storage nor the database")
			}
		})
	}
}

func TestUpload_Limit(t *testing.T) {
	data := imageBytes(t, "jpeg")

	t.Run("sixth avatar is rejected before the upload", func(t *testing.T) {
		s, repo, storage := newService(t)
		userID := uuid.New()
		for range MaxActiveAvatars {
			repo.add(userID, false)
		}

		_, err := s.Upload(context.Background(), uploadReq(userID, data))
		if !errors.Is(err, apperrors.ErrAvatarLimitReached) {
			t.Fatalf("expected ErrAvatarLimitReached, got %v", err)
		}
		if len(storage.objects) != 0 {
			t.Error("the file must not be uploaded when the limit is already reached")
		}
	})

	t.Run("deleted avatars and other users do not count", func(t *testing.T) {
		s, repo, _ := newService(t)
		userID := uuid.New()
		for range MaxActiveAvatars - 1 {
			repo.add(userID, false)
		}
		repo.add(userID, true)
		repo.add(uuid.New(), false)

		if _, err := s.Upload(context.Background(), uploadReq(userID, data)); err != nil {
			t.Fatalf("upload must succeed with four active avatars: %v", err)
		}
	})

	t.Run("delete frees a slot", func(t *testing.T) {
		s, repo, _ := newService(t)
		userID := uuid.New()

		var first entity.Avatar
		for i := range MaxActiveAvatars {
			a := repo.add(userID, false)
			if i == 0 {
				first = a
			}
		}

		if _, err := s.Upload(context.Background(), uploadReq(userID, data)); !errors.Is(err, apperrors.ErrAvatarLimitReached) {
			t.Fatalf("expected ErrAvatarLimitReached, got %v", err)
		}
		if err := s.Delete(context.Background(), userID, first.ID); err != nil {
			t.Fatalf("delete: %v", err)
		}
		if _, err := s.Upload(context.Background(), uploadReq(userID, data)); err != nil {
			t.Fatalf("upload after delete must succeed: %v", err)
		}
	})
}

func TestUpload_DatabaseFailureRemovesTheObject(t *testing.T) {
	tests := []struct {
		name      string
		createErr error
		want      error
	}{
		{"limit reached in the transaction", apperrors.ErrAvatarLimitReached, apperrors.ErrAvatarLimitReached},
		{"database failure", errors.New("db down"), apperrors.ErrInternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, repo, storage := newService(t)
			repo.createErr = tt.createErr

			_, err := s.Upload(context.Background(), uploadReq(uuid.New(), imageBytes(t, "png")))
			if !errors.Is(err, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, err)
			}
			if len(storage.objects) != 0 || len(storage.deletedKeys) != 1 {
				t.Errorf("the uploaded object must be removed: objects=%d, deleted=%v", len(storage.objects), storage.deletedKeys)
			}
		})
	}

	t.Run("failed cleanup does not hide the original error", func(t *testing.T) {
		s, repo, storage := newService(t)
		repo.createErr = errors.New("db down")
		storage.deleteErr = errors.New("storage down")

		_, err := s.Upload(context.Background(), uploadReq(uuid.New(), imageBytes(t, "png")))
		if !errors.Is(err, apperrors.ErrInternal) {
			t.Fatalf("expected ErrInternal, got %v", err)
		}
	})
}

func TestUpload_StorageFailure(t *testing.T) {
	s, repo, storage := newService(t)
	storage.putErr = errors.New("storage down")

	_, err := s.Upload(context.Background(), uploadReq(uuid.New(), imageBytes(t, "jpeg")))
	if !errors.Is(err, apperrors.ErrInternal) {
		t.Fatalf("expected ErrInternal, got %v", err)
	}
	if len(repo.avatars) != 0 {
		t.Error("no database record must be created when the upload failed")
	}
}

func TestUpload_URLFailureDoesNotFailTheUpload(t *testing.T) {
	s, repo, storage := newService(t)
	storage.urlErr = errors.New("presign failed")
	userID := uuid.New()

	avatar, err := s.Upload(context.Background(), uploadReq(userID, imageBytes(t, "jpeg")))
	if err != nil {
		t.Fatalf("a link failure must not fail an upload that already succeeded: %v", err)
	}
	if avatar.URL != "" {
		t.Errorf("url must be empty, got %q", avatar.URL)
	}
	if repo.activeCount(userID) != 1 {
		t.Error("the avatar must be saved")
	}
}

func TestDelete(t *testing.T) {
	t.Run("soft deletes the avatar and keeps the record and the file", func(t *testing.T) {
		s, repo, storage := newService(t)
		userID := uuid.New()
		a := repo.add(userID, false)

		if err := s.Delete(context.Background(), userID, a.ID); err != nil {
			t.Fatalf("delete: %v", err)
		}

		stored, ok := repo.avatars[a.ID]
		if !ok || stored.DeletedAt == nil {
			t.Error("the record must stay with deleted_at set (soft delete)")
		}
		if len(storage.deletedKeys) != 0 {
			t.Error("the file is removed later by the cleanup, not on delete")
		}
	})

	t.Run("repeated delete", func(t *testing.T) {
		s, repo, _ := newService(t)
		userID := uuid.New()
		a := repo.add(userID, false)

		_ = s.Delete(context.Background(), userID, a.ID)

		if err := s.Delete(context.Background(), userID, a.ID); !errors.Is(err, apperrors.ErrNotFoundAvatar) {
			t.Fatalf("expected ErrNotFoundAvatar, got %v", err)
		}
	})

	t.Run("someone else's avatar looks like a missing one and stays intact", func(t *testing.T) {
		s, repo, _ := newService(t)
		owner := uuid.New()
		a := repo.add(owner, false)

		err := s.Delete(context.Background(), uuid.New(), a.ID)
		if !errors.Is(err, apperrors.ErrNotFoundAvatar) {
			t.Fatalf("expected ErrNotFoundAvatar, got %v", err)
		}
		if repo.avatars[a.ID].DeletedAt != nil {
			t.Error("another user must not be able to delete the avatar")
		}
	})

	t.Run("unknown id", func(t *testing.T) {
		s, _, _ := newService(t)

		if err := s.Delete(context.Background(), uuid.New(), uuid.New()); !errors.Is(err, apperrors.ErrNotFoundAvatar) {
			t.Fatalf("expected ErrNotFoundAvatar, got %v", err)
		}
	})

	t.Run("repository failure", func(t *testing.T) {
		s, repo, _ := newService(t)
		repo.deleteErr = errors.New("db down")

		if err := s.Delete(context.Background(), uuid.New(), uuid.New()); !errors.Is(err, apperrors.ErrInternal) {
			t.Fatalf("expected ErrInternal, got %v", err)
		}
	})
}
