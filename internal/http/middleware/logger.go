package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		lat := time.Since(start)
		status := c.Writer.Status()

		reqID, _ := c.Get("request_id")
		log.Info().
			Str("request_id", toStr(reqID)).
			Int("status", status).
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Dur("latency", lat).
			Msg("http")
	}
}

func toStr(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
