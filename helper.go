package router

import (
	"net/http"
	"strings"

	"github.com/go-on/wrap"

	"github.com/go-on/router/route"
	"github.com/go-on/wrap-contrib/wraps"
)

func Mount(path string, r *Router) error { return r.Mount(path, http.DefaultServeMux) }

func MustMount(path string, r *Router) {
	err := Mount(path, r)
	if err != nil {
		panic(err.Error())
	}
}

// wrapper around a http.Handler generator to make it a http.Handler
type RouterFunc func() http.Handler

func (hc RouterFunc) ServeHTTP(rw http.ResponseWriter, req *http.Request) { hc().ServeHTTP(rw, req) }

type RouteParameterFunc func(*route.Route) []map[string]string

func (rpf RouteParameterFunc) Params(rt *route.Route) []map[string]string {
	return rpf(rt)
}

type RouteParameter interface {
	Params(*route.Route) []map[string]string
}

func IsCurrentRoute(req *http.Request, rt *route.Route) bool {
	return GetRouteId(req) == rt.Id
}

func GetRouteId(req *http.Request) (id string) {
	if i := strings.Index(req.URL.Fragment, "//0x"); i != -1 {
		return req.URL.Fragment[i:]
	}
	return
}

var slashB = []byte("/")[0]

// since req.URL.Path has / unescaped so that originally escaped / are
// indistinguishable from escaped ones, we are save here, i.e. / is
// already handled as path splitted and no key or value has / in it
// also it is save to use req.URL.Fragment since that will never be transmitted
// by the request
func GetRouteParam(req *http.Request, key string) (res string) {
	start, end := func() (start, end int) {
		var keyStart = 0
		var valStart = -1
		var inSlash bool
		for i := 0; i < len(req.URL.Fragment); i++ {
			if req.URL.Fragment[i] == slashB {
				if inSlash {
					break
				}
				inSlash = true
				if keyStart > -1 {
					if req.URL.Fragment[keyStart:i] == key {
						valStart = i + 1
					}
					keyStart = -1
					continue
				}

				keyStart = i + 1

				if valStart > -1 {
					return valStart, i
				}
				continue
			}
			inSlash = false
		}
		return -1, -1
	}()

	if start == -1 {
		return
	}
	return req.URL.Fragment[start:end]
}

// func (r *Router) Route(path string) *route.Route { return r.routes[path] }
type Etagged struct{}

func (et *Etagged) Wrap(next http.Handler) http.Handler {
	return wrap.New(
		wraps.IfNoneMatch,
		wraps.IfMatch(next),
		wraps.ETag,
		wrap.Handler(next),
	)
}

/*
func ETagged() (r *Router) {
	r = New()
	r.AddWrappers(wraps.IfNoneMatch, wraps.IfMatch(r), wraps.ETag)
	return
}
*/
