package profilehandler

import (
	"context"
	"errors"
	"go_home/internal/domain/dto"
	"go_home/internal/domain/entity"
	apperrors "go_home/pkg/errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	httpmiddleware "go_home/internal/middleware"
)

type ProfileService interface {
	GetMyProfile(ctx context.Context, userID uuid.UUID) (entity.User, error)
}

type ProfileHTTPHandler struct {
	profileService ProfileService
}

func NewProfileHTTPHandler(profileService ProfileService) *ProfileHTTPHandler {
	return &ProfileHTTPHandler{profileService: profileService}
}

func (h *ProfileHTTPHandler) RegisterRoutes(router *gin.RouterGroup) {
	profile := router.Group("/profile")
	{
		profile.GET("/my", h.getMy)
	}
}

// getMy godoc
// @Summary      Get my profile
// @Description  Returns the profile of the currently authenticated user.
// @Tags         profile
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.UserResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /profile/my [get]
func (h *ProfileHTTPHandler) getMy(c *gin.Context) {
	rawID, ok := httpmiddleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "invalid or expired token",
		})
		return
	}

	userID, err := uuid.Parse(rawID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "invalid or expired token",
		})
		return
	}

	user, err := h.profileService.GetMyProfile(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFoundUser) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "not found",
				Message: "user not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.UserResponse{
		ID:    user.ID.String(),
		Login: user.Login,
		Email: user.Email,
		Age:   user.Age,
		Bio:   user.Bio,
	})
}
