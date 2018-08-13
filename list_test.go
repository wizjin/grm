package grm

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRESTSort(t *testing.T) {
	g := New("mongodb://127.0.0.1/test")
	defer g.Close()
	r := gin.New()
	r.GET("/err", g.List("user1"))
	assert.HTTPError(t, r.ServeHTTP, "GET", "http://127.0.0.1/err", nil)

	time.Sleep(time.Second)
	r.GET("/none", func(ctx *gin.Context) { assert.Nil(t, GetQuery(ctx)) })
	r.GET("/user", g.List("user"), func(ctx *gin.Context) {
		q := GetQuery(ctx)
		assert.NotNil(t, q)
	})
	assert.HTTPSuccess(t, r.ServeHTTP, "GET", "http://127.0.0.1/none", nil)
	assert.HTTPSuccess(t, r.ServeHTTP, "GET", "http://127.0.0.1/user?filter=age eq 18&start=1&size=10&sort=id|name", nil)
}
