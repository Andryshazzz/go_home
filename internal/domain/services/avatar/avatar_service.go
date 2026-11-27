package avatarservice

import (
	"context"
	"go_home/internal/domain/entity"
	"go_home/pkg/logger"
	"io"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	MaxActiveAvatars = 5
	maxFileSize      = 10 << 20
	sniffLen         = 512
	urlTTL           = time.Hour
)

type AvatarRepository interface {
	CountActive(ctx context.Context, userID uuid.UUID) (int, error)
	CreateAvatar(ctx context.Context, avatar entity.Avatar, maxActive int) (entity.Avatar, error)
	SoftDeleteAvatar(ctx context.Context, userID, avatarID uuid.UUID) error
}

type FileStorage interface {
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	Delete(ctx context.Context, key string) error
	PresignedGetURL(ctx context.Context, key string, ttl time.Duration) (string, error)
}

type AvatarService struct {
	avatarRepository AvatarRepository
	storage          FileStorage
	log              *logger.Logger
}

func NewAvatarService(avatarRepository AvatarRepository, storage FileStorage, log *logger.Logger) *AvatarService {
	return &AvatarService{avatarRepository: avatarRepository, storage: storage, log: log}
}

func (s *AvatarService) logFor(ctx context.Context) *logger.Logger {
	if id := logger.RequestIDFromContext(ctx); id != "" {
		return s.log.With(zap.String("request_id", id))
	}

	return s.log
}
