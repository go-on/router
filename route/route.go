package route

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-on/method"
)

type MountPather interface {
	MountPath() string
}

type Route struct {
	Handlers         map[method.Method]http.Handler
	RessourceOptions string
	MountedPath      string
	OriginalPath     string
	Router           MountPather
}

func NewRoute(path string) *Route {
	rt := &Route{OriginalPath: path, MountedPath: path}
	rt.Handlers = map[method.Method]http.Handler{}
	return rt
}

func (r *Route) AddHandler(handler http.Handler, v method.Method) error {
	h := r.Handlers[v]
	if h != nil {
		return fmt.Errorf("handler for method %s already defined", v)
	}
	r.Handlers[v] = handler
	return nil
}

func (r *Route) SetHandler(m method.Method, h http.Handler) {
	r.Handlers[m] = h
}

// just a stupid helper to make shared routes look nicer
func (r *Route) AddMethod(v method.Method) *Route {
	r.AddHandler(nil, v)
	return r
}

/*

func Get(path string, h http.Handler) *Route {
	rt := NewRoute(path)
	rt.AddHandler(h, method.GET)
	return rt
}

func Post(path string, h http.Handler) *Route {
	rt := NewRoute(path)
	rt.AddHandler(h, method.POST)
	return rt
}

func Patch(path string, h http.Handler) *Route {
	rt := NewRoute(path)
	rt.AddHandler(h, method.PATCH)
	return rt
}

func Delete(path string, h http.Handler) *Route {
	rt := NewRoute(path)
	rt.AddHandler(h, method.DELETE)
	return rt
}

func Put(path string, h http.Handler) *Route {
	rt := NewRoute(path)
	rt.AddHandler(h, method.PUT)
	return rt
}
*/

func (rt *Route) HasMethod(m method.Method) bool {
	_, has := rt.Handlers[m]
	return has
}

var colon string = ":"

func URL(r *Route, params ...string) (string, error) {
	if len(params)%2 != 0 {
		panic("number of params must be even (pairs of key, value)")
	}

	vars := map[string]string{}
	for i := 0; i < len(params); i += 2 {
		vars[params[i]] = params[i+1]
	}

	parts := strings.Split(r.MountedPath, "/")

	for i, part := range parts {
		if part[:1] == colon {
			//if parts[0] == ":" {
			param, has := vars[part[1:]]
			if !has {
				return "", fmt.Errorf("missing parameter: %s", part[1:])
			}
			parts[i] = param
		}
	}

	return r.Router.MountPath() + strings.Join(parts, "/"), nil
}

func MustURL(r *Route, params ...string) string {
	url, err := URL(r, params...)
	if err != nil {
		panic(err.Error())
	}
	return url
}
