package route

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-on/method"
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
	Options(url string, callback func(js.Object))
}

type Route struct {
	GETHandler     http.Handler
	POSTHandler    http.Handler
	PUTHandler     http.Handler
	PATCHHandler   http.Handler
	DELETEHandler  http.Handler
	OPTIONSHandler http.Handler
	MountedPath    string
	DefinitionPath string
	Router         MountPather
	Id             string
}

func NewRoute(path string) *Route {
	// method.
	rt := &Route{DefinitionPath: path, MountedPath: path}
	rt.Id = fmt.Sprintf("//%p", rt)
	// rt.Handlers = map[method.Method]http.Handler{}
	return rt
}

func (r *Route) Clone() *Route {
	rt := &Route{
		GETHandler:     r.GETHandler,
		POSTHandler:    r.POSTHandler,
		PUTHandler:     r.PUTHandler,
		PATCHHandler:   r.PATCHHandler,
		DELETEHandler:  r.DELETEHandler,
		OPTIONSHandler: r.OPTIONSHandler,
		MountedPath:    r.MountedPath,
		DefinitionPath: r.DefinitionPath,
		Router:         r.Router,
	}
	rt.Id = fmt.Sprintf("//%p", rt)
	return rt
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

func (r *Route) Options(data interface{}, callback func(js.Object), params ...string) {
	ajax.Put(MustURL(r, params...), data, callback)
}

func (r *Route) Handler(meth method.Method) http.Handler {
	switch meth {
	case method.GET:
		return r.GETHandler
	case method.POST:
		return r.POSTHandler
	case method.PUT:
		return r.PUTHandler
	case method.PATCH:
		return r.PATCHHandler
	case method.DELETE:
		return r.DELETEHandler
	case method.OPTIONS:
		return r.OPTIONSHandler
	}
	return nil
}

func (r *Route) EachHandler(fn func(http.Handler) error) error {
	if r.GETHandler != nil {
		err := fn(r.GETHandler)
		if err != nil {
			return err
		}
	}

	if r.POSTHandler != nil {
		err := fn(r.POSTHandler)
		if err != nil {
			return err
		}
	}

	if r.PUTHandler != nil {
		err := fn(r.PUTHandler)
		if err != nil {
			return err
		}
	}

	if r.PATCHHandler != nil {
		err := fn(r.PATCHHandler)
		if err != nil {
			return err
		}
	}

	if r.DELETEHandler != nil {
		err := fn(r.DELETEHandler)
		if err != nil {
			return err
		}
	}

	if r.OPTIONSHandler != nil {
		err := fn(r.OPTIONSHandler)
		if err != nil {
			return err
		}
	}
	return nil
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
