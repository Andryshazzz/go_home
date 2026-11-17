package entity

type RegisterUserResponse struct {
	AccessToken  string
	RefreshToken string
}

type LoginUserResponse struct {
	User         User
	AccessToken  string
	RefreshToken string
}
