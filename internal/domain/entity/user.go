package entity

import "github.com/google/uuid"

type User struct {
	ID           uuid.UUID
	Login        string
	Email        string
	Age          int
	PasswordHash string
	Bio          string
}

func NewUser(
	login string,
	email string,
	age int,
	passwordHash string,
	bio string,
) User {
	return User{
		ID:           uuid.New(),
		Login:        login,
		Email:        email,
		Age:          age,
		PasswordHash: passwordHash,
		Bio:          bio,
	}
}
