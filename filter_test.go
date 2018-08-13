package grm

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
	"github.com/stretchr/testify/assert"
)

func TestSetAndGet(t *testing.T) {
	c := &gin.Context{}
	assert.Equal(t, bson.M{}, GetFilter(c))
	SetFilter(c, nil)
	assert.Equal(t, bson.M{}, GetFilter(c))
	f := bson.M{"test": "123"}
	SetFilter(c, f)
	assert.Equal(t, f, GetFilter(c))
}

func TestPaseFilter(t *testing.T) {
	assert.Equal(t, bson.M{
		"_id":   bson.M{"$ne": "xyz"},
		"name":  "abc",
		"test":  false,
		"block": bson.M{"$ne": true},
		"age":   bson.M{"$gte": 20, "$lt": 50},
		"score": bson.M{"$gt": 10.0, "$lte": 30.0},
	}, paseFilter(bson.M{}, "id ne 'xyz'|name eq 'abc'|block ne true|test eq false|age ge 20|age lt 50|score gt 10.0|score le 30.0"))
}
