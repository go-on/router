package routematcher

import (
	"net/http"

	"github.com/go-on/router"
	"github.com/go-on/router/route"
	"gopkg.in/go-on/wrap-contrib.v2/wraps"
)

type matchRoute struct {
	router *router.Router
	*route.Route
}

func (mr *matchRoute) Match(r *http.Request) bool {
	return mr.router.RequestRoute(r) == mr.Route
}

// MatchRoute returns a  wraps.Matcher that allows forking within middleware based on
// route matching
func MatchRoute(rtr *router.Router, rt *route.Route) wraps.Matcher {
	return &matchRoute{rtr, rt}
}
