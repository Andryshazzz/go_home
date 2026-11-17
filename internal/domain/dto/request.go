package dto

type RegisterRequest struct {
	Login    string `json:"login"    binding:"required,min=3,max=64" example:"ivan"`
	Email    string `json:"email"    binding:"required,email"        example:"ivan@example.com"`
	Password string `json:"password" binding:"required,min=8"        example:"Str0ngPass!"`
	Age      int    `json:"age"      binding:"required,gt=0,lt=150"  example:"25"`
	Bio      string `json:"bio"      binding:"max=1000"              example:"Hello, I am Ivan"`
}
