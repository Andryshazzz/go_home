package dto

type RegisterResponse struct {
	AccessToken  string `json:"access_token"  example:"eyJhbGciOiJIUzI1NiIs..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIs..."`
}

type ErrorResponse struct {
	Error   string `json:"error" example:"invalid request"`
	Message string `json:"message,omitempty" example:"Key: 'RegisterRequest.Email' Error:Field validation for 'Email' failed on the 'email' tag"`
}
