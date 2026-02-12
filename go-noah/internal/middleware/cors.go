package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.GetHeader("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,PATCH,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin, Referer, User-Agent")
		c.Header("Access-Control-Expose-Headers", "Authorization, Content-Length")

		if method == "OPTIONS" {
			c.Header("Access-Control-Max-Age", "7200")
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
