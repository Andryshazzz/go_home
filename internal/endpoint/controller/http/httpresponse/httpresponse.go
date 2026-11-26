package httpresponse

import (
	"errors"
	"strings"
	"unicode"

	"go_home/internal/domain/dto"
	apperrors "go_home/pkg/errors"
	"go_home/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

func Error(c *gin.Context, err error) {
	status, body := build(c, err)
	logHTTPError(c, status, err)
	c.JSON(status, body)
}

func AbortWithError(c *gin.Context, err error) {
	status, body := build(c, err)
	logHTTPError(c, status, err)
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

func logHTTPError(c *gin.Context, status int, err error) {
	log := logger.FromContext(c.Request.Context())

	fields := []zap.Field{
		zap.Int("status", status),
		zap.Error(err),
	}
	if appErr, ok := errors.AsType[apperrors.AppError](err); ok {
		fields = append(fields,
			zap.String("error_code", appErr.Code),
			zap.String("error_message", appErr.Message),
		)
	}

	if status >= 500 {
		log.Error("http request failed", fields...)
	} else {
		log.Warn("http request rejected", fields...)
	}
}
