package avatarservice

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"go_home/internal/domain/entity"
	apperrors "go_home/pkg/errors"
	"go_home/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var allowedTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
}

func (s *AvatarService) Upload(ctx context.Context, req entity.UploadAvatarRequest) (entity.Avatar, error) {
	log := s.logFor(ctx).With(zap.String("user_id", req.UserID.String()))

	if req.File == nil || req.Size <= 0 {
		return entity.Avatar{}, apperrors.ErrInvalidArgument.WithDetails(
			apperrors.FieldError{Field: "file", Reason: "empty"},
		)
	}
	if req.Size > maxFileSize {
		log.Warn("avatar rejected", zap.String("reason", "too_large"), zap.Int64("size", req.Size))
		return entity.Avatar{}, apperrors.ErrFileTooLarge
	}

	contentType, err := detectContentType(req.File)
	if err != nil {
		return entity.Avatar{}, fmt.Errorf("%w: read uploaded file: %w", apperrors.ErrInternal, err)
	}
	ext, ok := allowedTypes[contentType]
	if !ok {
		log.Warn("avatar rejected", zap.String("reason", "unsupported_type"), zap.String("detected_type", contentType))
		return entity.Avatar{}, apperrors.ErrUnsupportedFileType
	}

	active, err := s.avatarRepository.CountActive(ctx, req.UserID)
	if err != nil {
		return entity.Avatar{}, fmt.Errorf("%w: count avatars: %w", apperrors.ErrInternal, err)
	}
	if active >= MaxActiveAvatars {
		log.Warn("avatar rejected", zap.String("reason", "limit_reached"), zap.Int("active", active))
		return entity.Avatar{}, apperrors.ErrAvatarLimitReached
	}

	key := uuid.NewString() + ext

	if err := s.storage.Put(ctx, key, req.File, req.Size, contentType); err != nil {
		return entity.Avatar{}, fmt.Errorf("%w: store avatar: %w", apperrors.ErrInternal, err)
	}

	avatar, err := s.avatarRepository.CreateAvatar(ctx, entity.Avatar{
		ID:        uuid.New(),
		UserID:    req.UserID,
		ObjectKey: key,
	}, MaxActiveAvatars)
	if err != nil {
		s.discardObject(ctx, log, key)

		if errors.Is(err, apperrors.ErrAvatarLimitReached) {
			log.Warn("avatar rejected", zap.String("reason", "limit_reached_concurrent"))
			return entity.Avatar{}, err
		}

		return entity.Avatar{}, fmt.Errorf("%w: create avatar: %w", apperrors.ErrInternal, err)
	}

	if url, err := s.storage.PresignedGetURL(ctx, key, urlTTL); err != nil {
		log.Warn("avatar url not generated", zap.String("avatar_id", avatar.ID.String()), zap.Error(err))
	} else {
		avatar.URL = url
	}

	log.Info("avatar uploaded",
		zap.String("avatar_id", avatar.ID.String()),
		zap.String("object_key", key),
		zap.String("content_type", contentType),
		zap.Int64("size", req.Size),
	)

	return avatar, nil
}

func detectContentType(f io.ReadSeeker) (string, error) {
	buf := make([]byte, sniffLen)

	n, err := io.ReadFull(f, buf)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return "", err
	}

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	return http.DetectContentType(buf[:n]), nil
}

func (s *AvatarService) discardObject(ctx context.Context, log *logger.Logger, key string) {
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()

	if err := s.storage.Delete(cleanupCtx, key); err != nil {
		log.Error("orphaned avatar object", zap.String("object_key", key), zap.Error(err))
	}
}
