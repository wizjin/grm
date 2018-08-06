package grm

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Middleware return gin handler
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.AbortWithStatus(http.StatusOK)
	}
}
