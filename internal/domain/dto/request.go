package dto

type RegisterRequest struct {
	Login    string `json:"login"    binding:"required,min=3,max=64" example:"ivan"`
	Email    string `json:"email"    binding:"required,email"        example:"ivan@example.com"`
	Password string `json:"password" binding:"required,min=8"        example:"Str0ngPass!"`
	Age      int    `json:"age"      binding:"required,gt=0,lt=150"  example:"25"`
	Bio      string `json:"bio"      binding:"max=1000"              example:"Hello, I am Ivan"`
}

type LoginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ListUsersRequest struct {
	Login  string `form:"login"`
	Limit  int    `form:"limit,default=20" binding:"min=1,max=100"`
	Offset int    `form:"offset" binding:"min=0"`
}

type UpdateProfileRequest struct {
	Login *string `json:"login"`
	Email *string `json:"email"`
	Age   *int    `json:"age"`
	Bio   *string `json:"bio"`
}
