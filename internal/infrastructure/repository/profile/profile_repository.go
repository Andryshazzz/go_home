package profilerepository

import (
	"context"
	"errors"
	"fmt"
	"go_home/internal/domain/entity"
	"go_home/pkg/database"
	apperrors "go_home/pkg/errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const uniqueViolation = "23505"

type ProfileRepository struct {
	pool *database.ConnectionPool
}

func NewProfileRepository(pool *database.ConnectionPool) *ProfileRepository {
	return &ProfileRepository{pool: pool}
}

func (r *ProfileRepository) GetUserByID(
	ctx context.Context,
	id uuid.UUID,
) (entity.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT id, login, email, age, password_hash, bio
		FROM users
		WHERE id = $1 AND deleted_at IS NULL;
	`

	var user entity.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Login,
		&user.Email,
		&user.Age,
		&user.PasswordHash,
		&user.Bio,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, apperrors.ErrNotFoundUser
		}

		return entity.User{}, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

func (r *ProfileRepository) ListUsers(
	ctx context.Context,
	filter entity.ListUsersRequest,
) ([]entity.User, int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	pattern := "%" + escapeLike(filter.Login) + "%"

	var total int
	if err := r.pool.QueryRow(
		ctx,
		`SELECT count(*) FROM users WHERE login ILIKE $1 ESCAPE '\' AND deleted_at IS NULL;`,
		pattern,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	rows, err := r.pool.Query(
		ctx,
		`
		SELECT id, login, email, age, password_hash, bio
		FROM users
		WHERE login ILIKE $1 ESCAPE '\' AND deleted_at IS NULL
		ORDER BY login, id
		LIMIT $2 OFFSET $3;
		`,
		pattern,
		filter.Limit,
		filter.Offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]entity.User, 0, filter.Limit)
	for rows.Next() {
		var u entity.User
		if err := rows.Scan(
			&u.ID,
			&u.Login,
			&u.Email,
			&u.Age,
			&u.PasswordHash,
			&u.Bio,
		); err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate users: %w", err)
	}

	return users, total, nil
}

func (r *ProfileRepository) UpdateUser(
	ctx context.Context,
	req entity.UpdateProfileRequest,
) (entity.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		UPDATE users
		SET login = COALESCE($2, login),
		    email = COALESCE($3, email),
		    age   = COALESCE($4, age),
		    bio   = COALESCE($5, bio)
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, login, email, age, password_hash, bio;
	`

	var user entity.User
	err := r.pool.QueryRow(
		ctx,
		query,
		req.UserID,
		req.Login,
		req.Email,
		req.Age,
		req.Bio,
	).Scan(
		&user.ID,
		&user.Login,
		&user.Email,
		&user.Age,
		&user.PasswordHash,
		&user.Bio,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, apperrors.ErrNotFoundUser
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return entity.User{}, apperrors.ErrUserAlreadyExists
		}

		return entity.User{}, fmt.Errorf("update user: %w", err)
	}

	return user, nil
}

func (r *ProfileRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(
		ctx,
		`UPDATE users SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL;`,
		id,
	)
	if err != nil {
		return fmt.Errorf("soft delete user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFoundUser
	}

	if _, err := tx.Exec(ctx, `DELETE FROM refresh_tokens WHERE user_id = $1;`, id); err != nil {
		return fmt.Errorf("delete user sessions: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func escapeLike(s string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

	return replacer.Replace(s)
}
