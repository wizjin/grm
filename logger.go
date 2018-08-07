package grm

import (
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
)

const loggerTimeFormat = "2006/01/02 15:04:05"

type logger struct {
	name string
	w    io.Writer
}

func mlogger(name string) *logger {
	return &logger{
		name: name,
		w:    gin.DefaultWriter,
	}
}

func clogger(ctx *gin.Context, name string) *logger {
	uid := fmt.Sprintf("%08X", reflect.ValueOf(ctx).Pointer())
	return &logger{
		name: fmt.Sprintf("%s.%s", name, uid[4:]),
		w:    gin.DefaultWriter,
	}
}

func (l *logger) Error(format string, args ...interface{}) {
	l.Output('E', format, args...)
}

func (l *logger) Warn(format string, args ...interface{}) {
	l.Output('W', format, args...)
}

func (l *logger) Info(format string, args ...interface{}) {
	l.Output('I', format, args...)
}

func (l *logger) Debug(format string, args ...interface{}) {
	l.Output('D', format, args...)
}

func (l *logger) Output(lvl byte, format string, v ...interface{}) {
	str := fmt.Sprintf(format, v...)
	if len(str) > 0 && str[len(str)-1] == '\n' {
		str = str[0 : len(str)-1]
	}
	now := time.Now()
	fmt.Fprintf(l.w, "%s %c/%s: %s\n", now.Format(loggerTimeFormat), lvl, l.name, str) // nolint: gas
}
