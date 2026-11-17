package authservice

import (
	"context"
	"go_home/internal/domain/entity"
	"go_home/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

type AuthRepository interface {
	CreateUser(ctx context.Context, user entity.User) (entity.User, error)
	GetUserByLogin(ctx context.Context, login string) (entity.User, error)
}

type AuthService struct {
	authRepository AuthRepository
	jwtService     *jwt.JWTService
}

func NewAuthService(authRepository AuthRepository, jwtService *jwt.JWTService) *AuthService {
	return &AuthService{
		authRepository: authRepository,
		jwtService:     jwtService,
	}
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	return string(bytes), err
}
