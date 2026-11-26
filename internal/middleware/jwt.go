package httpmiddleware

import (
	"go_home/internal/endpoint/controller/http/httpresponse"
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
			httpresponse.AbortWithError(c, apperrors.ErrInvalidToken)

			return
		}

		parts := strings.Fields(authHeader)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			httpresponse.AbortWithError(c, apperrors.ErrInvalidToken)

			return
		}

		claims, err := m.jwtService.ParseAccess(parts[1])
		if err != nil {
			httpresponse.AbortWithError(c, apperrors.ErrInvalidToken)

			return
		}
		c.Set(userIDKey, claims.Subject)
		c.Next()
	}
}

func GetUserID(c *gin.Context) (string, bool) {
	value, exists := c.Get(userIDKey)
	if !exists {
		return "", false
	}

	userID, ok := value.(string)

	return userID, ok
}
