package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-on/wrap"
)

type Context struct {
	http.ResponseWriter
	Time time.Time
}

func (c *Context) Context(ctx interface{}) {
	*(ctx.(*time.Time)) = c.Time
}

func (c *Context) SetContext(ctx interface{}) {
	c.Time = *(ctx.(*time.Time))
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
