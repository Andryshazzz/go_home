package authhandler

import (
	"context"
	"errors"
	"go_home/internal/domain/dto"
	entity "go_home/internal/domain/entity"
	apperrors "go_home/pkg/errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthService interface {
	Register(ctx context.Context, req entity.RegisterUserRequest) (*entity.RegisterUserResponse, error)
	Login(ctx context.Context, req entity.LoginUserRequest) (*entity.LoginUserResponse, error)
	Logout(ctx context.Context, req entity.LogoutUserRequest) error
	Refresh(ctx context.Context, req entity.RefreshTokenRequest) (*entity.RefreshTokenResponse, error)
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
		auth.POST("/login", h.login)
		auth.POST("/logout", h.logout)
		auth.POST("/refresh", h.refresh)
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
		entity.RegisterUserRequest{
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

// login godoc
// @Summary      Log in
// @Description  Authenticates a user by login and password and returns a new access/refresh token pair.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.LoginRequest   true  "Credentials"
// @Success      200      {object}  dto.LoginResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Failure      401      {object}  dto.ErrorResponse
// @Failure      500      {object}  dto.ErrorResponse
// @Router       /auth/login [post]
func (h *AuthHTTPHandler) login(c *gin.Context) {
	ctx := c.Request.Context()

	var request dto.LoginRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid request",
			Message: err.Error(),
		})
		return
	}

	response, err := h.authService.Login(
		ctx,
		entity.LoginUserRequest{
			Login:    request.Login,
			Password: request.Password,
		},
	)
	if err != nil {
		if errors.Is(err, apperrors.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "unauthorized",
				Message: "invalid login or password",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal error",
			Message: "something went wrong",
		})
		return
	}

	c.JSON(http.StatusOK, dto.LoginResponse{
		User: dto.UserResponse{
			ID:    response.User.ID.String(),
			Login: response.User.Login,
			Email: response.User.Email,
			Age:   response.User.Age,
			Bio:   response.User.Bio,
		},
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
	})
}

// logout godoc
// @Summary      Log out
// @Description  Invalidates the given refresh token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.LogoutRequest   true  "Refresh token"
// @Success      204
// @Failure      400      {object}  dto.ErrorResponse
// @Failure      401      {object}  dto.ErrorResponse
// @Failure      500      {object}  dto.ErrorResponse
// @Router       /auth/logout [post]
func (h *AuthHTTPHandler) logout(c *gin.Context) {
	ctx := c.Request.Context()

	var request dto.LogoutRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid request",
			Message: err.Error(),
		})
		return
	}

	err := h.authService.Logout(
		ctx,
		entity.LogoutUserRequest{
			RefreshToken: request.RefreshToken,
		},
	)
	if err != nil {
		if errors.Is(err, apperrors.ErrInvalidToken) || errors.Is(err, apperrors.ErrTokenMismatch) {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "unauthorized",
				Message: "invalid or expired token",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal error",
			Message: err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// refresh godoc
// @Summary      Refresh tokens
// @Description  Exchanges a valid refresh token for a new access/refresh pair. The old refresh token stops working immediately.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.RefreshRequest   true  "Refresh token"
// @Success      200      {object}  dto.RefreshResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Failure      401      {object}  dto.ErrorResponse
// @Failure      500      {object}  dto.ErrorResponse
// @Router       /auth/refresh [post]
func (h *AuthHTTPHandler) refresh(c *gin.Context) {
	var request dto.RefreshRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid request",
			Message: err.Error(),
		})
		return
	}

	response, err := h.authService.Refresh(
		c.Request.Context(),
		entity.RefreshTokenRequest{RefreshToken: request.RefreshToken},
	)
	if err != nil {
		if appErr, ok := errors.AsType[apperrors.AppError](err); ok {
			c.JSON(appErr.HTTPCode, dto.ErrorResponse{
				Error:   appErr.Code,
				Message: appErr.Message,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   apperrors.ErrInternal.Code,
			Message: apperrors.ErrInternal.Message,
		})
		return
	}

	c.JSON(http.StatusOK, dto.RefreshResponse{
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
	})
}
