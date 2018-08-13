package grm

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

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
			s := []string{}
			for _, i := range append(GetSort(ctx), strings.Split(sstr, "|")...) {
				if i == "id" {
					s = append(s, "id")
				} else {
					s = append(s, i)
				}
			}
			if len(s) > 0 {
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
