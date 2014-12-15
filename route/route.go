/*
 Package route provides slim representation of routes that is used by go-on/router.Router
 and may be used by client side libraries such as gopherjs.

 A route intentionally has no handler. That is left to be defined/assigned by the server side
 library - that is go-on/router. Therefor some properties of Route that must not be changed
 by hand are exported to allow go-on/router to do its work.

*/
package route

import (
	"strings"

	"gopkg.in/go-on/method.v1"
)

// Route is a slim representation of a Route mainly to get the corresponding URLs.
// However a route can and must be "mounted" beneath a path either via Mount() (for clientside
// libraries) or via the go-on/router.Router.
//
// Routes could be shared between client and server by defining them in a seperate package and
// mounting them to a shared mountpoint. This package is then imported from the client and the
// server. The server then remounts the routes to a router with the same mountpath and assigns
// the handler for the various methods.

// Don't directly change the properties of Route.
// Instead use only the methods.
// These properties only are exported to be used by Router
type Route struct {

	// Id will be set by the go-on/router.Router, don't touch it!
	// An internal attribute but exported because it is cross-package
	Id string

	// Methods will be read by the go-on/router.Router, don't touch it!
	// An internal attribute but exported because it is cross-package
	Methods map[method.Method]struct{}

	// DefinitionPath will be read by the go-on/router.Router, don't touch it!
	// An internal attribute but exported because it is cross-package
	DefinitionPath string

	// Router will be set by either Mount or go-on/router.Router, don't touch it!
	// An internal attribute but exported because it is cross-package
	Router interface {
		// MountPath returns the path where the router is mounted
		MountPath() string
	}
}

// New creates a new route for the given path and methods.
// The methods for the route must not be changed after the call of New
func New(path string, method1 method.Method, furtherMethods ...method.Method) *Route {
	methods := append(furtherMethods, method1)
	rt := &Route{
		DefinitionPath: path,
		Methods:        map[method.Method]struct{}{},
	}
	for _, m := range methods {
		if !m.IsKnown() {
			panic(ErrUnknownMethod{m})
		}
		rt.Methods[m] = struct{}{}
	}
	return rt
}

// URL returns the url of the mounted route for the given params (key/value pairs)
func (r *Route) URL(params ...string) (string, error) {
	if len(params) == 0 {
		return r.URLMap(nil)
	}
	if len(params)%2 != 0 {
		panic(ErrPairParams{})
	}
	vars := map[string]string{}
	for i := 0; i < len(params); i += 2 {
		vars[params[i]] = params[i+1]
	}
	return r.URLMap(vars)
}

var PARAM_PREFIX = []byte(":")[0]

// URL returns the url of the mounted route for the given params (map)
func (r *Route) URLMap(params map[string]string) (string, error) {
	if r == nil {
		panic(ErrRouteIsNil{})
	}
	if params == nil && !r.HasParams() {
		return r.MountedPath(), nil
	}
	mountedPath := r.MountedPath()
	parts := strings.Split(mountedPath[1:], "/")
	for i, part := range parts {
		if part[0] == PARAM_PREFIX {
			param, has := params[part[1:]]
			if !has {
				return "", ErrMissingParam{part[1:], r.MountedPath()}
			}
			parts[i] = param
		}
	}
	return "/" + strings.Join(parts, "/"), nil
}

// MustURL is like URL but panics on errors
func (r *Route) MustURL(params ...string) string {
	url, err := r.URL(params...)
	if err != nil {
		panic(err)
	}
	return url
}

// MustURLMap is like URLMap but panics on errors
func (r *Route) MustURLMap(params map[string]string) string {
	url, err := r.URLMap(params)
	if err != nil {
		panic(err)
	}
	return url
}

// HasMethod checks if the route has the given method.
func (r *Route) HasMethod(m method.Method) (has bool) {
	if m == method.HEAD {
		return r.HasMethod(method.GET)
	}
	_, has = r.Methods[m]
	return
}

// HasParams checks if the path of the route has placeholders for params
func (r *Route) HasParams() bool {
	return strings.ContainsRune(r.DefinitionPath, ':')
}

// MountedPath returns the path of the mounted route
func (r *Route) MountedPath() string {
	if r.Router.MountPath() == "/" {
		return r.DefinitionPath
	}
	return r.Router.MountPath() + r.DefinitionPath
}

// Mount mounts the given routes beneath the given mountPoint
// It should be used in libraries that are used by a client or shared
// between client and server
func Mount(mountPoint string, routes ...*Route) {
	for _, rt := range routes {
		rt.mount(mountPoint)
	}
}

// mount is a helper for client side mounting, making use of pseudoRouter
func (r *Route) mount(mountPoint string) {
	if r.Router != nil {
		panic(&ErrDoubleMounted{r.Router.MountPath(), r})
	}
	r.Router = pseudoRouter(mountPoint)
}

// pseudoRouter is a helper to mount based on a string
type pseudoRouter string

func (mp pseudoRouter) MountPath() string {
	return string(mp)
}
