package route

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gopherjs/gopherjs/js"
)

type MountPather interface {
	MountPath() string
}

var ajax AjaxHandler

func RegisterAjaxHandler(aj AjaxHandler) {
	if ajax != nil {
		panic("already registered")
	}
	ajax = aj
}

type AjaxHandler interface {
	Get(url string, callback func(js.Object))
	Post(url string, data interface{}, callback func(js.Object))
	Put(url string, data interface{}, callback func(js.Object))
	Patch(url string, data interface{}, callback func(js.Object))
	Delete(url string, callback func(js.Object))
}

type Route struct {
	// Handlers         map[method.Method]http.Handler
	GETHandler     http.Handler
	POSTHandler    http.Handler
	PUTHandler     http.Handler
	PATCHHandler   http.Handler
	DELETEHandler  http.Handler
	OPTIONSHandler http.Handler
	// RessourceOptions string
	MountedPath  string
	OriginalPath string
	Router       MountPather
	Id           string
}

func NewRoute(path string) *Route {
	// method.
	rt := &Route{OriginalPath: path, MountedPath: path}
	rt.Id = fmt.Sprintf("%p", rt)
	// rt.Handlers = map[method.Method]http.Handler{}
	return rt
}

func (r *Route) Clone() *Route {
	return &Route{
		GETHandler:     r.GETHandler,
		POSTHandler:    r.POSTHandler,
		PUTHandler:     r.PUTHandler,
		PATCHHandler:   r.PATCHHandler,
		DELETEHandler:  r.DELETEHandler,
		OPTIONSHandler: r.OPTIONSHandler,
		MountedPath:    r.MountedPath,
		OriginalPath:   r.OriginalPath,
		Router:         r.Router,
	}
}

func (r *Route) Get(callback func(js.Object), params ...string) {
	ajax.Get(MustURL(r, params...), callback)
}

func (r *Route) Delete(callback func(js.Object), params ...string) {
	ajax.Delete(MustURL(r, params...), callback)
}

func (r *Route) Post(data interface{}, callback func(js.Object), params ...string) {
	ajax.Post(MustURL(r, params...), data, callback)
}

func (r *Route) Patch(data interface{}, callback func(js.Object), params ...string) {
	ajax.Patch(MustURL(r, params...), data, callback)
}

func (r *Route) Put(data interface{}, callback func(js.Object), params ...string) {
	ajax.Put(MustURL(r, params...), data, callback)
}

/*
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
*/

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

/*
func (rt *Route) HasMethod(m method.Method) bool {
	_, has := rt.Handlers[m]
	return has
}
*/

var colon string = ":"

func URL(r *Route, params ...string) (string, error) {
	if len(params)%2 != 0 {
		panic("number of params must be even (pairs of key, value)")
	}

	vars := map[string]string{}
	for i := 0; i < len(params); i += 2 {
		vars[params[i]] = params[i+1]
	}

	// panic(fmt.Sprintf("mounted path: %#v\n", r.MountedPath))
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
