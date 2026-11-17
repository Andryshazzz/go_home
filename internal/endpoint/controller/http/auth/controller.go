package authhandler

import (
	"context"
	"go_home/internal/domain/dto"
	authservice "go_home/internal/domain/services/auth"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthService interface {
	Register(ctx context.Context, req authservice.RegisterUserRequest) (*authservice.RegisterUserResponse, error)
}

type AuthHTTPHandler struct {
	authService AuthService
}

func NewAuthHTTPHandler(
	authService AuthService,
) *AuthHTTPHandler {
	return &AuthHTTPHandler{
		authService: authService,
	}
}

func (h *AuthHTTPHandler) RegisterRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", h.register)
	}
}

func (h *AuthHTTPHandler) register(c *gin.Context) {
	ctx := c.Request.Context()

	var request dto.RegisterRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request",
			"message": err.Error(),
		})
		return
	}

	response, err := h.authService.Register(
		ctx,
		authservice.RegisterUserRequest{
			Login:    request.Login,
			Email:    request.Email,
			Password: request.Password,
			Age:      request.Age,
			Bio:      request.Bio,
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "registration failed",
		})
		return
	}

	c.JSON(http.StatusCreated, dto.RegisterResponse{
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
	})
}
