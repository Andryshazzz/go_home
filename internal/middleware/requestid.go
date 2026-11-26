package httpmiddleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	requestIDHeader = "X-Request-ID"
	requestIDKey    = "request_id"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(requestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}

		c.Set(requestIDKey, id)
		c.Header(requestIDHeader, id)

		c.Next()
	}
}

func GetRequestID(c *gin.Context) string {
	id, _ := c.Get(requestIDKey)
	s, _ := id.(string)
	return s
}
