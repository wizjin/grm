package grm

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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
}

func TestInvalidDB(t *testing.T) {
	g := New("mongodb://127.0.0.2/test")
	defer g.Close()
	time.Sleep(time.Second)
}

func TestReconnectDB(t *testing.T) {
	g := New("mongodb://127.0.0.1/test")
	defer g.Close()
	for i := -1; i < int(dbReconnectInterval.Seconds()); i++ {
		s := g.s.Load().(*mgo.Session)
		if s != nil {
			s.SetSocketTimeout(time.Nanosecond) // Force ping failed
		}
		time.Sleep(time.Second)
	}
}

func TestMiddleware(t *testing.T) {
	g := New("mongodb://127.0.0.1/test")
	defer g.Close()
	r := gin.New()
	r.GET("/test", g.Middleware())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}
