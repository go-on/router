package router

import (
	"net/http"
	"strings"

	"github.com/go-on/wrap"

	"github.com/go-on/wrap-contrib/wraps"
)

/*
func MayMount(path string, r *Router) error { return r.MayMount(path, http.DefaultServeMux) }

func Mount(path string, r *Router) {
	err := MayMount(path, r)
	if err != nil {
		panic(err.Error())
	}
}
*/

/*
// wrapper around a http.Handler generator to make it a http.Handler
type RouterFunc func() http.Handler

func (hc RouterFunc) ServeHTTP(rw http.ResponseWriter, req *http.Request) { hc().ServeHTTP(rw, req) }
*/

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
/*
type Etagged struct{}

func (et *Etagged) Wrap(next http.Handler) http.Handler {
	return wrap.New(
		wraps.IfNoneMatch,
		wraps.IfMatch(next),
		wraps.ETag,
		wrap.Handler(next),
	)
}
*/

func NewETagged(wrappers ...wrap.Wrapper) (ø *Router) {
	ø = newRouter()
	wrappers = append(
		wrappers,
		wraps.IfNoneMatch,
		wraps.IfMatch(ø),
		wraps.ETag,
	)
	ø.addWrappers(wrappers...)
	return
}

/*
func ETagged() (r *Router) {
	r = New()
	r.AddWrappers(wraps.IfNoneMatch, wraps.IfMatch(r), wraps.ETag)
	return
}
*/
