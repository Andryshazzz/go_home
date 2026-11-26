package httpresponse

import (
	"errors"
	"strings"
	"unicode"

	"go_home/internal/domain/dto"
	apperrors "go_home/pkg/errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func Error(c *gin.Context, err error) {
	status, body := build(c, err)
	c.JSON(status, body)
}

func AbortWithError(c *gin.Context, err error) {
	status, body := build(c, err)
	c.AbortWithStatusJSON(status, body)
}

func BindError(c *gin.Context, err error) {
	ve, ok := errors.AsType[validator.ValidationErrors](err)
	if !ok {
		Error(c, apperrors.ErrInvalidRequest)

		return
	}

	details := make([]apperrors.FieldError, 0, len(ve))
	for _, fe := range ve {
		details = append(details, apperrors.FieldError{
			Field:  snakeCase(fe.Field()),
			Reason: fe.Tag(),
		})
	}

	Error(c, apperrors.ErrInvalidArgument.WithDetails(details...))
}

func build(c *gin.Context, err error) (int, dto.ErrorResponse) {
	appErr, ok := errors.AsType[apperrors.AppError](err)
	if !ok {
		appErr = apperrors.ErrInternal
	}

	if appErr.HTTPCode >= 500 {
		_ = c.Error(err)
	}

	return appErr.HTTPCode, dto.ErrorResponse{
		Error:   appErr.Code,
		Message: appErr.Message,
		Details: appErr.Details,
	}
}

func snakeCase(s string) string {
	var b strings.Builder

	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			r = unicode.ToLower(r)
		}
		b.WriteRune(r)
	}

	return b.String()
}
