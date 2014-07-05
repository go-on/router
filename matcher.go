package router

import (
	"github.com/go-on/router/route"
	"github.com/go-on/wrap-contrib/wraps"
	"net/http"
)

type matchRoute struct {
	*route.Route
	router *Router
}

func (mr *matchRoute) Match(r *http.Request) bool {
	return mr.router.RequestRoute(r) == mr.Route
}

// MatchRoute returns a  wraps.Matcher that allows forking within middleware based on
// route matching
func (rter *Router) MatchRoute(rt *route.Route) wraps.Matcher {
	return &matchRoute{rt, rter}
}
