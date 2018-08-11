package grm

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
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

// SetQuery to gin context
func SetQuery(ctx *gin.Context, q *mgo.Query) {
	ctx.Set("grm.rest.query", q)
}

// GetQuery from gin context
func GetQuery(ctx *gin.Context) *mgo.Query {
	if f, ok := ctx.Get("grm.rest.query"); ok {
		return f.(*mgo.Query)
	}
	return nil
}

// SetItems to gin context
func SetItems(ctx *gin.Context, v interface{}) {
	ctx.Set("grm.rest.items", v)
}

// GetItems from gin context
func GetItems(ctx *gin.Context) interface{} {
	if f, ok := ctx.Get("grm.rest.items"); ok {
		return f
	}
	return nil
}

// SetSort to gin context
func SetSort(ctx *gin.Context, s []string) {
	ctx.Set("grm.rest.sort", s)
}

// GetSort from gin context
func GetSort(ctx *gin.Context) []string {
	if f, ok := ctx.Get("grm.rest.sort"); ok {
		return f.([]string)
	}
	return []string{}
}

// List document from database
func (g *GRM) List(name string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		s := g.s.Load().(*mgo.Session)
		if s == nil {
			clogger(ctx, "GRM").Debug("link collection failed: %s", name)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c := s.DB("").C(name)
		filter := GetFilter(ctx)
		if str, ok := ctx.GetQuery("filter"); ok {
			filter = paseFilter(filter, str)
		}
		n, err := c.Find(filter).Count()
		if err != nil {
			clogger(ctx, "GRM").Error("count data failed: %v", err)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		ctx.Set("grm.rest.count", n)
		q := c.Find(filter)
		// Sort
		if sstr, ok := ctx.GetQuery("sort"); ok {
			s := strToParams(sstr)
			if len(s) > 0 {
				ss := GetSort(ctx)
				if len(ss) > 0 {
					s = append(ss, s...)
				}
				q = q.Sort(s...)
			}
		}
		// Pagination
		if s, ok := ctx.GetQuery("start"); ok {
			if i, err := strconv.Atoi(s); err == nil && i > 0 {
				q = q.Skip(i)
			}
		}
		if s, ok := ctx.GetQuery("size"); ok {
			if i, err := strconv.Atoi(s); err == nil && i > 0 {
				q = q.Limit(i)
			}
		}
		SetQuery(ctx, q)
		ctx.Next()
		if !ctx.Writer.Written() {
			var res struct {
				Count int         `json:"count"`
				Items interface{} `json:"items"`
			}
			res.Count = n
			res.Items = GetItems(ctx)
			if res.Items == nil {
				lst := []bson.M{}
				i := q.Iter()
				for !i.Done() {
					item := bson.M{}
					if !i.Next(&item) {
						break
					}
					if id, ok := item["_id"]; ok {
						item["id"] = id
						delete(item, "_id")
					}
					lst = append(lst, item)
				}
			}
			if res.Items == nil {
				res.Items = []string{}
			}
			ctx.JSON(http.StatusOK, res)
		}
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

// Helper function
func strToParams(str string) []string {
	return strings.Split(strings.Replace(str, "id", "_id", -1), "|")
}
