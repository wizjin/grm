package grm

import (
	"log"
	"net/http"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
)

// GRM module instance
type GRM struct {
	s atomic.Value
}

var nullSession *mgo.Session

// New GRM module
func New(dburl string) *GRM {
	g := &GRM{}
	g.s.Store(nullSession)
	s, err := mgo.Dial(dburl)
	if err != nil {
		log.Println(err.Error())
		return nil
	}
	g.s.Store(s)
	return g
}

// Close all grm resource
func (g *GRM) Close() {
	s := g.s.Load().(*mgo.Session)
	if s != nil {
		g.s.Store(nullSession)
		s.Close()
	}
}

// Middleware return gin handler
func (g *GRM) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.AbortWithStatus(http.StatusOK)
	}
}
