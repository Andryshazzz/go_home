package entity

type RegisterUserRequest struct {
	Login    string
	Email    string
	Password string
	Age      int
	Bio      string
}

type LoginUserRequest struct {
	Login    string
	Password string
}

type LogoutUserRequest struct {
	RefreshToken string
}
