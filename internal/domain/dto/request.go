package dto

type RegisterRequest struct {
	Login    string `json:"login" binding:"required,min=3,max=32"`
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=72"`
	Age      int    `json:"age" binding:"required,min=18,max=120"`
	Bio      string `json:"bio" binding:"max=1000"`
}
