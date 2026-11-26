package httpmiddleware

import (
	"time"

	"go_home/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RequestLogger(base *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := GetRequestID(c)

		reqLog := base.With(
			zap.String("request_id", reqID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		)

		c.Set("logger", reqLog)
		ctx := logger.WithContext(c.Request.Context(), reqLog)
		ctx = logger.WithRequestID(ctx, reqID)
		c.Request = c.Request.WithContext(ctx)

		start := time.Now()
		c.Next()

		status := c.Writer.Status()
		fields := []zap.Field{
			zap.Int("status", status),
			zap.Duration("latency", time.Since(start)),
			zap.String("client_ip", c.ClientIP()),
			zap.Int("bytes_out", c.Writer.Size()),
		}
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("gin_error", c.Errors.String()))
		}

		switch {
		case status >= 500:
			reqLog.Error("request completed", fields...)
		case status >= 400:
			reqLog.Warn("request completed", fields...)
		default:
			reqLog.Info("request completed", fields...)
		}
	}
}

func LoggerFromGin(c *gin.Context) *logger.Logger {
	if v, ok := c.Get("logger"); ok {
		if l, ok := v.(*logger.Logger); ok {
			return l
		}
	}
	return logger.FromContext(c.Request.Context())
}
