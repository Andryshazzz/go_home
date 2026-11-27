package avatarhandler

import (
	"context"
	"errors"
	"fmt"
	"go_home/internal/domain/dto"
	"go_home/internal/domain/entity"
	"go_home/internal/endpoint/controller/http/httpresponse"
	httpmiddleware "go_home/internal/middleware"
	apperrors "go_home/pkg/errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	formField     = "file"
	maxAvatarSize = 10 << 20
	maxBodySize   = maxAvatarSize + (1 << 20)
)

type AvatarService interface {
	Upload(ctx context.Context, req entity.UploadAvatarRequest) (entity.Avatar, error)
	Delete(ctx context.Context, userID, avatarID uuid.UUID) error
}

type AvatarHTTPHandler struct {
	avatarService AvatarService
}

func NewAvatarHTTPHandler(avatarService AvatarService) *AvatarHTTPHandler {
	return &AvatarHTTPHandler{avatarService: avatarService}
}

func (h *AvatarHTTPHandler) RegisterRoutes(router *gin.RouterGroup) {
	avatars := router.Group("/profile/my/avatars")
	{
		avatars.POST("", h.uploadAvatar)
		avatars.DELETE("/:id", h.deleteAvatar)
	}
}

// uploadAvatar godoc
// @Summary      Upload an avatar
// @Description  Uploads a JPEG or PNG image (up to 10 MB) as an avatar of the current user. A user can have up to 5 active avatars.
// @Tags         avatars
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file  formData  file  true  "Image, JPEG or PNG, up to 10 MB"
// @Success      201   {object}  dto.AvatarResponse
// @Failure      400   {object}  dto.ErrorResponse
// @Failure      401   {object}  dto.ErrorResponse
// @Failure      409   {object}  dto.ErrorResponse
// @Failure      413   {object}  dto.ErrorResponse
// @Failure      415   {object}  dto.ErrorResponse
// @Failure      500   {object}  dto.ErrorResponse
// @Router       /profile/my/avatars [post]
func (h *AvatarHTTPHandler) uploadAvatar(c *gin.Context) {
	userID, ok := httpmiddleware.GetUserUUID(c)
	if !ok {
		httpresponse.Error(c, apperrors.ErrInvalidToken)
		return
	}

	if c.Request.ContentLength > maxBodySize {
		httpresponse.Error(c, apperrors.ErrFileTooLarge)
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodySize)

	header, err := c.FormFile(formField)
	if err != nil {
		httpresponse.Error(c, formError(err))
		return
	}

	switch {
	case header.Size <= 0:
		httpresponse.Error(c, apperrors.ErrInvalidArgument.WithDetails(
			apperrors.FieldError{Field: formField, Reason: "empty"},
		))
		return
	case header.Size > maxAvatarSize:
		httpresponse.Error(c, apperrors.ErrFileTooLarge)
		return
	}

	file, err := header.Open()
	if err != nil {
		httpresponse.Error(c, fmt.Errorf("%w: open uploaded file: %w", apperrors.ErrInternal, err))
		return
	}
	defer func() { _ = file.Close() }()

	avatar, err := h.avatarService.Upload(c.Request.Context(), entity.UploadAvatarRequest{
		UserID: userID,
		File:   file,
		Size:   header.Size,
	})
	if err != nil {
		httpresponse.Error(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.AvatarResponse{
		ID:        avatar.ID.String(),
		URL:       avatar.URL,
		CreatedAt: avatar.CreatedAt,
	})
}

// deleteAvatar godoc
// @Summary      Delete an avatar
// @Description  Soft-deletes an avatar of the current user. The slot becomes free for a new upload.
// @Tags         avatars
// @Security     BearerAuth
// @Param        id   path  string  true  "Avatar ID"  format(uuid)
// @Success      204
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /profile/my/avatars/{id} [delete]
func (h *AvatarHTTPHandler) deleteAvatar(c *gin.Context) {
	userID, ok := httpmiddleware.GetUserUUID(c)
	if !ok {
		httpresponse.Error(c, apperrors.ErrInvalidToken)
		return
	}

	avatarID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpresponse.Error(c, apperrors.ErrInvalidArgument.WithDetails(
			apperrors.FieldError{Field: "id", Reason: "uuid"},
		))
		return
	}

	if err := h.avatarService.Delete(c.Request.Context(), userID, avatarID); err != nil {
		httpresponse.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func formError(err error) error {
	if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
		return apperrors.ErrFileTooLarge
	}

	if errors.Is(err, http.ErrMissingFile) {
		return apperrors.ErrInvalidArgument.WithDetails(
			apperrors.FieldError{Field: formField, Reason: "required"},
		)
	}

	return apperrors.ErrInvalidRequest
}
