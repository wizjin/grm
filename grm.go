package grm

import (
	"net/http"
	"strings"
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
	dbConnectTimeout    = 5 * time.Second
	dbReconnectInterval = 5 * time.Second
)

// New GRM module
func New(dburl string) *GRM {
	g := &GRM{}
	g.run = 1
	g.done = make(chan bool)
	g.s.Store(nullSession)
	go g.watchDB(dburl)
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

// C get collection from context
func C(ctx *gin.Context, name string) *mgo.Collection {
	if c, ok := ctx.Get("grm.db." + name); ok {
		return c.(*mgo.Collection)
	}
	return nil
}

// C set collection by names
func (g *GRM) C(names ...string) gin.HandlerFunc {
	m := map[string]string{}
	for _, n := range names {
		ps := strings.Split(n, ":")
		if len(ps) > 1 {
			m[ps[1]] = ps[0]
		} else {
			m[ps[0]] = ps[0]
		}
	}
	return func(ctx *gin.Context) {
		s := g.s.Load().(*mgo.Session)
		if s == nil {
			clogger(ctx, "GRM").Debug("link collections failed: %s", strings.Join(names, ","))
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		db := s.DB("")
		for k, v := range m {
			ctx.Set("grm.db."+k, db.C(v))
		}
		ctx.Next()
	}
}

func (g *GRM) watchDB(dburl string) {
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
	info.Timeout = dbConnectTimeout
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
