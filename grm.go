package grm

import (
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
)

// GRM module instance
type GRM struct {
	lck  sync.Mutex
	s    atomic.Value
	run  int32
	done chan bool
}

var (
	nullSession         *mgo.Session
	dbReconnectInterval = 5 * time.Second
)

// New GRM module
func New(dburl string) *GRM {
	g := &GRM{}
	g.run = 1
	g.done = make(chan bool)
	g.s.Store(nullSession)
	go g.guarddb(dburl)
	return g
}

// Close all grm resource
func (g *GRM) Close() {
	g.lck.Lock()
	defer g.lck.Unlock()
	if g.run > 0 {
		atomic.StoreInt32(&g.run, 0)
		<-g.done
		s := g.s.Load().(*mgo.Session)
		if s != nil {
			g.s.Store(nullSession)
			s.Close()
		}
	}
}

// Middleware return gin handler
func (g *GRM) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.AbortWithStatus(http.StatusOK)
	}
}

func (g *GRM) guarddb(dburl string) {
	l := mlogger("GRM")
	defer func() {
		if r := recover(); r != nil {
			l.Error("recovered from %v", r)
		}
		close(g.done)
	}()
	info, err := mgo.ParseURL(dburl)
	if err != nil {
		panic(err.Error())
	}
	info.Timeout = time.Second * 2
	wait := time.Now()
	for atomic.LoadInt32(&g.run) > 0 {
		if time.Until(wait) > 0 {
			time.Sleep(time.Millisecond * 100)
			continue
		}
		s := g.s.Load().(*mgo.Session)
		if s != nil {
			if err := s.Ping(); err != nil {
				g.s.Store(nullSession)
				s.Close()
				l.Warn("test database connect failed: %v", err)
				wait = time.Now()
				continue
			}
			wait = time.Now().Add(time.Second)
			continue
		}
		s, err := mgo.DialWithInfo(info)
		if err != nil {
			l.Error("connect mongodb failed: %v", err)
			wait = time.Now().Add(dbReconnectInterval)
			continue
		}
		l.Info("connect mongodb success")
		g.s.Store(s)
	}
}
