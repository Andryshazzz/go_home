package profilehandler

import (
	"context"
	"errors"
	"go_home/internal/domain/dto"
	"go_home/internal/domain/entity"
	apperrors "go_home/pkg/errors"
	"net/http"
	"strings"

	httpmiddleware "go_home/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProfileService interface {
	GetMyProfile(ctx context.Context, userID uuid.UUID) (entity.User, error)
	ListUsers(ctx context.Context, req entity.ListUsersRequest) (*entity.ListUsersResponse, error)
	UpdateMyProfile(ctx context.Context, req entity.UpdateProfileRequest) (entity.User, error)
	DeleteMyProfile(ctx context.Context, userID uuid.UUID) error
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
		profile.GET("", h.list)
		profile.GET("/my", h.getMy)
		profile.PATCH("/my", h.updateMy)
		profile.DELETE("/my", h.deleteMy)
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
	userID, ok := currentUserID(c)
	if !ok {
		respondError(c, apperrors.ErrInvalidToken)
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

// list godoc
// @Summary      List users
// @Description  Returns a paginated list of users. Optional case-insensitive search by login.
// @Tags         profile
// @Produce      json
// @Security     BearerAuth
// @Param        login   query     string  false  "Login search (substring)"
// @Param        limit   query     int     false  "Page size"  minimum(1)  maximum(100)  default(20)
// @Param        offset  query     int     false  "Offset"     minimum(0)  default(0)
// @Success      200     {object}  dto.ListUsersResponse
// @Failure      400     {object}  dto.ErrorResponse
// @Failure      401     {object}  dto.ErrorResponse
// @Failure      500     {object}  dto.ErrorResponse
// @Router       /profile [get]
func (h *ProfileHTTPHandler) list(c *gin.Context) {
	var request dto.ListUsersRequest

	if err := c.ShouldBindQuery(&request); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid request",
			Message: err.Error(),
		})
		return
	}

	response, err := h.profileService.ListUsers(
		c.Request.Context(),
		entity.ListUsersRequest{
			Login:  strings.TrimSpace(request.Login),
			Limit:  request.Limit,
			Offset: request.Offset,
		},
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

	items := make([]dto.UserResponse, 0, len(response.Users))
	for _, u := range response.Users {
		items = append(items, dto.UserResponse{
			ID:    u.ID.String(),
			Login: u.Login,
			Email: u.Email,
			Age:   u.Age,
			Bio:   u.Bio,
		})
	}

	c.JSON(http.StatusOK, dto.ListUsersResponse{
		Items:  items,
		Total:  response.Total,
		Limit:  request.Limit,
		Offset: request.Offset,
	})
}

// updateMy godoc
// @Summary      Update my profile
// @Description  Partially updates the profile of the current user. Only the provided fields are changed.
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      dto.UpdateProfileRequest  true  "Fields to update"
// @Success      200      {object}  dto.UserResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Failure      401      {object}  dto.ErrorResponse
// @Failure      404      {object}  dto.ErrorResponse
// @Failure      409      {object}  dto.ErrorResponse
// @Failure      422      {object}  dto.ErrorResponse
// @Failure      500      {object}  dto.ErrorResponse
// @Router       /profile/my [patch]
func (h *ProfileHTTPHandler) updateMy(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		respondError(c, apperrors.ErrInvalidToken)
		return
	}

	var request dto.UpdateProfileRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid request",
			Message: err.Error(),
		})
		return
	}

	user, err := h.profileService.UpdateMyProfile(
		c.Request.Context(),
		entity.UpdateProfileRequest{
			UserID: userID,
			Login:  request.Login,
			Email:  request.Email,
			Age:    request.Age,
			Bio:    request.Bio,
		},
	)
	if err != nil {
		respondError(c, err)
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

// deleteMy godoc
// @Summary      Delete my profile
// @Description  Soft-deletes the current user. The account can no longer log in or be listed.
// @Tags         profile
// @Security     BearerAuth
// @Success      204
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /profile/my [delete]
func (h *ProfileHTTPHandler) deleteMy(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		respondError(c, apperrors.ErrInvalidToken)
		return
	}

	if err := h.profileService.DeleteMyProfile(c.Request.Context(), userID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func currentUserID(c *gin.Context) (uuid.UUID, bool) {
	rawID, ok := httpmiddleware.GetUserID(c)
	if !ok {
		return uuid.Nil, false
	}

	userID, err := uuid.Parse(rawID)
	if err != nil {
		return uuid.Nil, false
	}

	return userID, true
}

func respondError(c *gin.Context, err error) {
	appErr, ok := errors.AsType[apperrors.AppError](err)
	if !ok {
		appErr = apperrors.ErrInternal
	}

	if appErr.HTTPCode >= http.StatusInternalServerError {
		c.Error(err)
	}

	c.JSON(appErr.HTTPCode, dto.ErrorResponse{
		Error:   appErr.Code,
		Message: appErr.Message,
	})
}
