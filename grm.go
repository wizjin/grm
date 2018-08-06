package grm

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GRM module instance
type GRM struct {
}

// New GRM module
func New(dburl string) *GRM {
	return &GRM{}
}

// Close all grm resource
func (g *GRM) Close() {
}

// Middleware return gin handler
func (g *GRM) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.AbortWithStatus(http.StatusOK)
	}
}
