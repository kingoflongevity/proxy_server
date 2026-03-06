package middleware

import (
	"time"

	"proxy_server/pkg/logger"

	"github.com/gin-gonic/gin"
)

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		
		// 处理请求
		c.Next()
		
		// 记录日志
		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		
		logger.Info("[%s] %s %s %d %v %s",
			method,
			path,
			clientIP,
			status,
			latency,
			c.Errors.String(),
		)
	}
}
