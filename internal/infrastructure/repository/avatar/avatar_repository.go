package avatarrepository

import (
	"context"
	"errors"
	"fmt"

	"go_home/internal/domain/entity"
	"go_home/pkg/database"
	apperrors "go_home/pkg/errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type AvatarRepository struct {
	pool *database.ConnectionPool
}

func NewAvatarRepository(pool *database.ConnectionPool) *AvatarRepository {
	return &AvatarRepository{pool: pool}
}

func (r *AvatarRepository) CountActive(ctx context.Context, userID uuid.UUID) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var count int
	if err := r.pool.QueryRow(ctx,
		`SELECT count(*) FROM avatars WHERE user_id = $1 AND deleted_at IS NULL;`,
		userID,
	).Scan(&count); err != nil {
		return 0, fmt.Errorf("count active avatars: %w", err)
	}

	return count, nil
}

func (r *AvatarRepository) CreateAvatar(
	ctx context.Context,
	avatar entity.Avatar,
	maxActive int,
) (entity.Avatar, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return entity.Avatar{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var locked int
	if err := tx.QueryRow(ctx,
		`SELECT 1 FROM users WHERE id = $1 AND deleted_at IS NULL FOR UPDATE;`,
		avatar.UserID,
	).Scan(&locked); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Avatar{}, apperrors.ErrNotFoundUser
		}

		return entity.Avatar{}, fmt.Errorf("lock user: %w", err)
	}

	var active int
	if err := tx.QueryRow(ctx,
		`SELECT count(*) FROM avatars WHERE user_id = $1 AND deleted_at IS NULL;`,
		avatar.UserID,
	).Scan(&active); err != nil {
		return entity.Avatar{}, fmt.Errorf("count active avatars: %w", err)
	}
	if active >= maxActive {
		return entity.Avatar{}, apperrors.ErrAvatarLimitReached
	}

	created := avatar
	if err := tx.QueryRow(ctx,
		`INSERT INTO avatars (id, user_id, object_key) VALUES ($1, $2, $3) RETURNING created_at;`,
		avatar.ID, avatar.UserID, avatar.ObjectKey,
	).Scan(&created.CreatedAt); err != nil {
		return entity.Avatar{}, fmt.Errorf("insert avatar: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return entity.Avatar{}, fmt.Errorf("commit tx: %w", err)
	}

	return created, nil
}

func (r *AvatarRepository) SoftDeleteAvatar(ctx context.Context, userID, avatarID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	tag, err := r.pool.Exec(ctx,
		`UPDATE avatars SET deleted_at = now() WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;`,
		avatarID, userID,
	)
	if err != nil {
		return fmt.Errorf("soft delete avatar: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFoundAvatar
	}

	return nil
}
