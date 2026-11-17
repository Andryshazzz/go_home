package authrepository

import (
	"context"
	"fmt"
	"go_home/internal/domain/entity"
	"go_home/pkg/database"
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
