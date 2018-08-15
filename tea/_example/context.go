package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-on/wrap"
)

type Context struct {
	http.ResponseWriter
	Time *time.Time
}

var _ = wrap.ValidateContextInjecter(&Context{})

func (c *Context) Context(ctx interface{}) (found bool) {
	found = true
	switch ty := ctx.(type) {
	case *http.ResponseWriter:
		*ty = c.ResponseWriter
	case *time.Time:
		if c.Time == nil {
			return false
		}
		*ty = *c.Time
	default:
		panic(&wrap.ErrUnsupportedContextGetter{ctx})

	}
	return
}

func (c *Context) SetContext(ctx interface{}) {
	switch ty := ctx.(type) {
	case *time.Time:
		c.Time = ty
	default:
		panic(&wrap.ErrUnsupportedContextSetter{ctx})

	}
}

func (c Context) Wrap(next http.Handler) http.Handler {
	var hf http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		c.ResponseWriter = w
		next.ServeHTTP(&c, r)
	}
	return hf
}

func start(next http.Handler, w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	w.(wrap.Contexter).SetContext(&now)
	next.ServeHTTP(w, r)
}

func stop(w http.ResponseWriter, r *http.Request) {
	var t time.Time
	w.(wrap.Contexter).Context(&t)
	fmt.Fprintf(w, "Time elapsed: %0.5f secs", time.Since(t).Seconds())
}
