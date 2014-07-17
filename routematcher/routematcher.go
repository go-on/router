package routematcher

import (
	"net/http"

	"github.com/go-on/router"
	"github.com/go-on/router/route"
	"github.com/go-on/wrap-contrib/wraps"
)

type matchRoute struct {
	*route.Route
	router *router.Router
}

func (mr *matchRoute) Match(r *http.Request) bool {
	return mr.router.RequestRoute(r) == mr.Route
}

// MatchRoute returns a  wraps.Matcher that allows forking within middleware based on
// route matching
func MatchRoute(rtr *Router, rt *route.Route) wraps.Matcher {
	return &matchRoute{rtr, rt}
}
