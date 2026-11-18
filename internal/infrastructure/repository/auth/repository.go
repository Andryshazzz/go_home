package authrepository

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

type AuthRepository struct {
	pool *database.ConnectionPool
}

func NewAuthRepository(
	pool *database.ConnectionPool,
) *AuthRepository {
	return &AuthRepository{
		pool: pool,
	}
}

func (r *AuthRepository) CreateUser(
	ctx context.Context,
	user entity.User,
) (entity.User, error) {
	ctx, cancel := context.WithTimeout(
		ctx,
		r.pool.OpTimeout(),
	)
	defer cancel()

	query := `
		INSERT INTO users (
			id,
			login,
			email,
			age,
			password_hash,
			bio
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING
			id,
			login,
			email,
			age,
			password_hash,
			bio;
	`

	var createdUser entity.User
	err := r.pool.QueryRow(
		ctx,
		query,
		user.ID,
		user.Login,
		user.Email,
		user.Age,
		user.PasswordHash,
		user.Bio,
	).Scan(
		&createdUser.ID,
		&createdUser.Login,
		&createdUser.Email,
		&createdUser.Age,
		&createdUser.PasswordHash,
		&createdUser.Bio,
	)

	if err != nil {
		return entity.User{}, fmt.Errorf("scan error: %w", err)
	}

	return createdUser, nil
}

func (r *AuthRepository) GetUserByLogin(
	ctx context.Context,
	login string,
) (entity.User, error) {
	ctx, cancel := context.WithTimeout(
		ctx,
		r.pool.OpTimeout(),
	)
	defer cancel()

	query := `
		SELECT
			id,
			login,
			email,
			age,
			password_hash,
			bio
		FROM users
		WHERE login = $1;
	`

	var user entity.User
	err := r.pool.QueryRow(ctx, query, login).Scan(
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

		return entity.User{}, fmt.Errorf("get user by login: %w", err)
	}

	return user, nil
}

func (r *AuthRepository) GetUserByID(
	ctx context.Context,
	id uuid.UUID,
) (entity.User, error) {
	ctx, cancel := context.WithTimeout(
		ctx,
		r.pool.OpTimeout(),
	)
	defer cancel()

	query := `
		SELECT
			id,
			login,
			email,
			age,
			password_hash,
			bio
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
