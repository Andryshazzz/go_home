package httpmiddleware

import (
	"errors"
	apperrors "go_home/pkg/errors"
	"go_home/pkg/jwt"
	"strings"

	"github.com/gin-gonic/gin"
)

const userIDKey = "user_id"

type JWTAuth struct {
	jwtService *jwt.JWTService
}

func NewJWTAuth(jwtService *jwt.JWTService) *JWTAuth {
	return &JWTAuth{
		jwtService: jwtService,
	}
}

func (m *JWTAuth) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			abortWithError(c, apperrors.ErrInvalidToken)

			return
		}
		
		parts := strings.Fields(authHeader)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			abortWithError(c, apperrors.ErrInvalidToken)

			return
		}

		claims, err := m.jwtService.ParseAccess(parts[1])
		if err != nil {
			switch {
			case errors.Is(err, jwt.ErrTokenExpired):
				abortWithError(c, apperrors.ErrInvalidToken)
			case errors.Is(err, jwt.ErrInvalidToken):
				abortWithError(c, apperrors.ErrInvalidToken)
			default:
				abortWithError(c, apperrors.ErrInvalidToken)
			}

			return
		}
		c.Set(userIDKey, claims.Subject)
		c.Next()
	}
}

func abortWithError(c *gin.Context, appErr apperrors.AppError) {
	c.AbortWithStatusJSON(appErr.HTTPCode, gin.H{
		"code":    appErr.Code,
		"message": appErr.Message,
		"details": appErr.Details,
	})
}

func GetUserID(c *gin.Context) (string, bool) {
	value, exists := c.Get(userIDKey)
	if !exists {
		return "", false
	}

	userID, ok := value.(string)
	
	return userID, ok
}
