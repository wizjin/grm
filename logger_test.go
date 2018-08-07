package grm

import (
	"bytes"
	"strings"
	"testing"
)

func TestLogger(t *testing.T) {
	b := bytes.Buffer{}
	l := mlogger("test")
	l.w = &b
	l.Error("123\n")
	if !strings.HasSuffix(b.String(), "123\n") {
		t.Error("invalid log new line suffix")
	}
}
