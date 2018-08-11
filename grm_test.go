//+build old

package grm

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard
	dbReconnectInterval = time.Second
	os.Exit(m.Run())
}

func TestInvalidConfig(t *testing.T) {
	g := New("fake://111?test=0")
	defer g.Close()
	assert.Nil(t, g.s.Load().(*mgo.Session))
}

func TestInvalidDB(t *testing.T) {
	g := New("mongodb://127.0.0.2/test")
	defer g.Close()
	time.Sleep(time.Second)
	assert.Nil(t, g.s.Load().(*mgo.Session))
	r := gin.New()
	r.GET("/test", g.C())
	assert.HTTPError(t, r.ServeHTTP, "GET", "/test", nil)
}

func TestReconnectDB(t *testing.T) {
	g := New("mongodb://127.0.0.1/test")
	defer g.Close()
	for i := 0; i < int(dbReconnectInterval.Seconds())*2; i++ {
		s := g.s.Load().(*mgo.Session)
		if s != nil {
			ss := s.Copy()
			ss.SetSocketTimeout(time.Nanosecond) // Force ping failed
			g.s.Store(ss)
			break
		}
		time.Sleep(time.Second)
	}
	time.Sleep(time.Second)
	assert.NotNil(t, g.s.Load().(*mgo.Session))
}

func TestMiddleware(t *testing.T) {
	g := New("mongodb://127.0.0.1/test")
	defer g.Close()
	r := gin.New()
	time.Sleep(time.Second)
	r.GET("/test", g.C("t", "b:B"), func(c *gin.Context) {
		assert.Nil(t, C(c, "a"))
		assert.NotNil(t, C(c, "t"))
		assert.Nil(t, C(c, "b"))
		assert.NotNil(t, C(c, "B"))
	})
	assert.HTTPSuccess(t, r.ServeHTTP, "GET", "/test", nil)
}
