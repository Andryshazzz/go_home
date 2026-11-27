package httpmiddleware

import (
	"go_home/internal/endpoint/controller/http/httpresponse"
	apperrors "go_home/pkg/errors"
	"go_home/pkg/jwt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const userIDKey = "user_id"

type JWTAuth struct {
	jwtService *jwt.Service
}

func NewJWTAuth(jwtService *jwt.Service) *JWTAuth {
	return &JWTAuth{
		jwtService: jwtService,
	}
}

func (m *JWTAuth) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			LoggerFromGin(c).Warn("missing authorization header")
			httpresponse.AbortWithError(c, apperrors.ErrInvalidToken)

			return
		}

		parts := strings.Fields(authHeader)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			LoggerFromGin(c).Warn("malformed authorization header")
			httpresponse.AbortWithError(c, apperrors.ErrInvalidToken)

			return
		}

		claims, err := m.jwtService.ParseAccess(parts[1])
		if err != nil {
			LoggerFromGin(c).Warn("invalid access token", zap.Error(err))
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

func GetUserUUID(c *gin.Context) (uuid.UUID, bool) {
	raw, ok := GetUserID(c)
	if !ok {
		return uuid.Nil, false
	}

	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, false
	}

	return id, true
}
