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
)

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
		WHERE id = $1;
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
		`SELECT count(*) FROM users WHERE login ILIKE $1 ESCAPE '\';`,
		pattern,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	rows, err := r.pool.Query(
		ctx,
		`
		SELECT id, login, email, age, password_hash, bio
		FROM users
		WHERE login ILIKE $1 ESCAPE '\'
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

func escapeLike(s string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

	return replacer.Replace(s)
}
