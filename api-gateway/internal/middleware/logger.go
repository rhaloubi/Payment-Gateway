package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhaloubi/api-gateway/internal/config"
)

func Logger(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		if cfg.Logging.Format == "json" {
			log.Printf(`{"time":"%s","method":"%s","path":"%s","query":"%s","ip":"%s","status":%d,"latency":"%s"}`,
				time.Now().Format(time.RFC3339),
				c.Request.Method,
				path,
				raw,
				c.ClientIP(),
				c.Writer.Status(),
				time.Since(start),
			)
		}
	}
}
