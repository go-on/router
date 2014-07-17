package route

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-on/method"
	"github.com/gopherjs/gopherjs/js"
)

// MountedRouter is a minimalistic routing interface for a mountable router
type MountedRouter interface {
	// Wrap wraps the inner (final) http.Handler
	Wrap(inner http.Handler) http.Handler

	// MountPath returns the path where the router is mounted
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
	DefinitionPath string
	Router         MountedRouter
	Id             string
}

func NewRoute(path string) *Route {
	rt := &Route{DefinitionPath: path}
	rt.Id = fmt.Sprintf("//%p", rt)
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
		DefinitionPath: r.DefinitionPath,
		Router:         r.Router,
	}
	rt.Id = fmt.Sprintf("//%p", rt)
	return rt
}

func (r *Route) MountedPath() string {
	if r.Router.MountPath() == "/" {
		return r.DefinitionPath
	}
	return r.Router.MountPath() + r.DefinitionPath
}

func (r *Route) Get(callback func(js.Object), params ...string) {
	ajax.Get(r.MustURL(params...), callback)
}

func (r *Route) Delete(callback func(js.Object), params ...string) {
	ajax.Delete(r.MustURL(params...), callback)
}

func (r *Route) Post(data interface{}, callback func(js.Object), params ...string) {
	ajax.Post(r.MustURL(params...), data, callback)
}

func (r *Route) Patch(data interface{}, callback func(js.Object), params ...string) {
	ajax.Patch(r.MustURL(params...), data, callback)
}

func (r *Route) Put(data interface{}, callback func(js.Object), params ...string) {
	ajax.Put(r.MustURL(params...), data, callback)
}

func (r *Route) Options(data interface{}, callback func(js.Object), params ...string) {
	ajax.Put(r.MustURL(params...), data, callback)
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

func (rt *Route) SetHandlerForMethod(handler http.Handler, m method.Method) {
	switch m {
	case method.GET:
		if rt.GETHandler != nil {
			panic("handler for GET already defined")
		}
		rt.GETHandler = handler
	case method.PUT:
		if rt.PUTHandler != nil {
			panic("handler for PUT already defined")
		}
		rt.PUTHandler = handler
	case method.POST:
		if rt.POSTHandler != nil {
			panic("handler for POST already defined")
		}
		rt.POSTHandler = handler
	case method.DELETE:
		if rt.DELETEHandler != nil {
			panic("handler for DELETE already defined")
		}
		rt.DELETEHandler = handler
	case method.PATCH:
		if rt.PATCHHandler != nil {
			panic("handler for PATCH already defined")
		}
		rt.PATCHHandler = handler
	case method.OPTIONS:
		if rt.OPTIONSHandler != nil {
			panic("handler for OPTIONS already defined")
		}
		rt.OPTIONSHandler = handler
	default:
		panic("unsupported method " + m)
	}
}

func (rt *Route) SetHandlerForMethods(handler http.Handler, methods ...method.Method) {
	for _, m := range methods {
		rt.SetHandlerForMethod(handler, m)
	}
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

// params are key/value pairs
func (r *Route) URL(params ...string) (string, error) {
	if len(params)%2 != 0 {
		panic("number of params must be even (pairs of key, value)")
	}
	vars := map[string]string{}
	for i := 0; i < len(params); i += 2 {
		vars[params[i]] = params[i+1]
	}
	return r.URLMap(vars)
}

var WILDCARD_SEPARATOR = []byte(":")[0]

func (r *Route) URLMap(params map[string]string) (string, error) {
	// mounted path mustalways begin with a /
	parts := strings.Split(r.MountedPath()[1:], "/")
	for i, part := range parts {
		if part[0] == WILDCARD_SEPARATOR {
			param, has := params[part[1:]]
			if !has {
				return "", fmt.Errorf("missing parameter: %s", part[1:])
			}
			parts[i] = param
		}
	}
	if r.Router.MountPath() == "/" {
		return "/" + strings.Join(parts, "/"), nil
	}
	return r.Router.MountPath() + "/" + strings.Join(parts, "/"), nil
}

func (r *Route) MustURL(params ...string) string {
	url, err := r.URL(params...)
	if err != nil {
		panic(err.Error())
	}
	return url
}

func (r *Route) MustURLMap(params map[string]string) string {
	url, err := r.URLMap(params)
	if err != nil {
		panic(err.Error())
	}
	return url
}

func (r *Route) HasParams() bool {
	return strings.ContainsRune(r.DefinitionPath, ':')
}

func Options(r *Route) []string {
	allow := []string{method.OPTIONS.String()}

	if r.GETHandler != nil {
		allow = append(allow, method.GET.String())
		allow = append(allow, method.HEAD.String())
	}

	if r.POSTHandler != nil {
		allow = append(allow, method.POST.String())
	}

	if r.DELETEHandler != nil {
		allow = append(allow, method.DELETE.String())
	}

	if r.PATCHHandler != nil {
		allow = append(allow, method.PATCH.String())
	}

	if r.PUTHandler != nil {
		allow = append(allow, method.PUT.String())
	}

	return allow
}

func (r *Route) SetOPTIONSHandler() {
	optionsString := strings.Join(Options(r), ",")
	r.OPTIONSHandler = http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Allow", optionsString)
	})
}
