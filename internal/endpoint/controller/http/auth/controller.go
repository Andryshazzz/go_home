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

// register godoc
// @Summary      Register a new user
// @Description  Creates a user and returns an access/refresh token pair.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.RegisterRequest   true  "Registration data"
// @Success      201      {object}  dto.RegisterResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Failure      409      {object}  dto.ErrorResponse
// @Failure      500      {object}  dto.ErrorResponse
// @Router       /auth/register [post]
func (h *AuthHTTPHandler) register(c *gin.Context) {
	ctx := c.Request.Context()

	var request dto.RegisterRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid request",
			Message: err.Error(),
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
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "invalid request",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.RegisterResponse{
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
	})
}
