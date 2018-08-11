package grm

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

// SetFilter to gin context
func SetFilter(ctx *gin.Context, f bson.M) {
	ctx.Set("grm.rest.filter", f)
}

// GetFilter from gin context
func GetFilter(ctx *gin.Context) bson.M {
	if f, ok := ctx.Get("grm.rest.filter"); ok {
		b := f.(bson.M)
		if b != nil {
			return b
		}
	}
	return bson.M{}
}

func paseFilter(filter bson.M, in string) bson.M {
	for _, f := range strToParams(in) {
		parts := strings.Split(f, " ")
		if len(parts) > 2 {
			key := parts[0]
			if parts[1] == "eq" {
				filter[key] = toValue(parts[2])
				continue
			}
			var m bson.M
			if mm, ok := filter[key]; ok {
				if bm, ok := mm.(bson.M); ok {
					m = bm
				}
			}
			if m == nil {
				m = bson.M{}
			}
			switch parts[1] {
			case "ne":
				m["$ne"] = toValue(parts[2])
			case "gt":
				m["$gt"] = toValue(parts[2])
			case "lt":
				m["$lt"] = toValue(parts[2])
			case "ge":
				m["$gte"] = toValue(parts[2])
			case "le":
				m["$lte"] = toValue(parts[2])
			}
			filter[key] = m
		}
	}
	return filter
}

func toValue(val string) interface{} {
	if val[0] == '\'' && val[len(val)-1] == '\'' {
		return val[1 : len(val)-1]
	}
	switch val {
	case "true":
		return true
	case "false":
		return false
	}
	if strings.ContainsAny(val, ".") {
		v, _ := strconv.ParseFloat(val, 64)
		return v
	}
	v, _ := strconv.ParseInt(val, 10, 64)
	return int(v)
}
