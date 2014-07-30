package router

import (
	"bytes"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/go-on/method"
	"github.com/go-on/router/route"
	"github.com/go-on/wrap"
	"github.com/go-on/wrap-contrib/wraps"
)

// Router is a mountable router routing paths to routes.
//
// Concurrently adding and serving routes is not supported.
// Routes must be defined none concurrently and before serving
type Router struct {
	node       *node
	wrapper    []wrap.Wrapper
	routes     map[string]*routeHandler
	parent     *Router
	mountPoint string
	path       string
	muxed      bool
	submounted bool
	// NotFound is called, if a http.Handler could not be found.
	// If it is set to nil, the status 405 is set
	NotFound http.Handler
}

// the given wrappers are near the inner call and called before the
// etag and IfMatch and IfNoneMatch wrappers. wrappers around them
// could be easily done by making a go-on/wrap.New() and use the Router
// as final http.Handler surrounded by other middleware
/*
func (r *Router) addWrappers(wrapper ...wrap.Wrapper) {
	r.wrapper = append(r.wrapper, wrapper...)
}
*/

// New creates a new router with optional wrappers
func New(wrapper ...wrap.Wrapper) (r *Router) {
	r = newRouter()
	r.wrapper = append(r.wrapper, wrapper...)
	// r.addWrappers(wrapper...)
	return
}

func newRouter() *Router {
	return &Router{
		routes: map[string]*routeHandler{},
		node:   newNode(),
	}
}

func (ø *Router) mustAddRoute(rt *routeHandler) {
	err := ø.addRoute(rt)
	if err != nil {
		panic(err)
	}
}

func (r *Router) IsMounted() bool {
	return r.mountPoint != ""
}

// Serve serves the request on the top level. It handles method override and path cleaning
// and then serves via the corresponding http.Handler of the route or passes it to a given wrapper
//
// Serve will selfmount the router under / if it is not already mounted
func (ø *Router) ServingHandler() http.Handler {
	if !ø.IsMounted() {
		ø.Mount("/", nil)
	}
	stack := []wrap.Wrapper{}
	if !ø.muxed {
		stack = append(stack, wraps.PrepareLikeMux())
	}
	// we can't handle the method override as part of the wraps, because it has to
	// be run before we look for the method (or we would have to run all wrappers before)
	// maybe we should not handle this case since it can be handled by given wrapper
	stack = append(stack, wraps.MethodOverride())
	/*
		if wrapper != nil {
			stack = append(stack, wrapper)
		}
	*/
	stack = append(stack, wrap.Handler(ø))
	return wrap.New(stack...)
}

func (ø *Router) addRoute(rt *routeHandler) error {
	if _, has := ø.routes[rt.DefinitionPath]; has {
		return ErrDoubleRegistration{rt.DefinitionPath}
	}
	ø.routes[rt.DefinitionPath] = rt
	rt.Router = ø
	rt.Id = fmt.Sprintf("//%p", rt)
	return nil
}

func (ø *Router) MountPath() string { return ø.path }

func (ø *Router) RequestRoute(rq *http.Request) (rt *route.Route) {
	_, rh, _ := ø.getHandler(rq)
	return rh.Route
}

// ServeHTTP serves the request via the http.Handler that is defined in the route
// to which the url points. If no route is found or no handler for the requested method is
// found, the NotFound handler serves the request. If there is no NotFound handler, the
// status code 405 (Method not allowed) is sent.
//
// ServeHTTP should be used as part of a composition. There are things that should only be
// done once per request, such as protocol checking and path normalization.
// These should be done by the toplevel Handler, see the Serve() http.HandlerFunc
func (r *Router) ServeHTTP(w http.ResponseWriter, rq *http.Request) {
	h := r.Dispatch(rq)
	if h == nil {
		// fmt.Println("found no handler")
		nf := r.NotFound
		if nf == nil {
			//w.WriteHeader(405)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		nf.ServeHTTP(w, rq)
		return
	}
	// fmt.Println("found a handler")
	// h.ServeHTTP(helper.NewPanicOnWrongWriteOrder(w), rq)
	h.ServeHTTP(w, rq)
}

// Dispatch returns the corresponding http.Handler for the request
// If no handler could be found, nil is returned
func (r *Router) Dispatch(rq *http.Request) http.Handler {
	// fmt.Printf("dispatching: %#v (%p)\n", rq.URL.Path, rq)
	if r.mountPoint == "" {
		panic(ErrNotMounted{})
	}
	h, rt, meth := r.getHandler(rq)

	if h == nil {
		// fmt.Printf("[dispatch] no handler for %s\n", rq.URL.Path)
		return nil
	}

	if meth != method.OPTIONS {
		return rt.Router.(*Router).wrap(h)
	}
	return h
}

// Wrap wraps the given inner handler with all wrappers
// of the router and its parents
// Wrap is part of the route/MountedRouter interface and should
// only be used internally, even if being exported
func (ø *Router) wrap(h http.Handler) http.Handler {
	for i := len(ø.wrapper) - 1; i >= 0; i-- {
		h = ø.wrapper[i].Wrap(h)
	}
	if ø.parent != nil {
		return ø.parent.wrap(h)
	}
	return h
}

type Muxer interface {
	Handle(path string, handler http.Handler)
}

func (r *Router) Mount(path string, m Muxer) {
	err := r.mayMount(path, m)
	if err != nil {
		panic(err)
	}
}

func mustNotBeRouter(handler http.Handler) {
	if _, is := handler.(*Router); is {
		panic(ErrRouterNotAllowed{})
	}
}

func (r *Router) HandleMethod(path string, handler http.Handler, m method.Method) *route.Route {
	mustNotBeRouter(handler)
	rt := r.newRoute(path, m)
	rt.SetHandlerForMethod(handler, m)
	return rt.Route
}

func (r *Router) HandleRoute(rt *route.Route, handler http.Handler) {
	mustNotBeRouter(handler)
	if _, has := r.routes[rt.DefinitionPath]; has {
		panic(ErrDoubleRegistration{rt.DefinitionPath})
	}
	rh := newRouteHandler(rt)
	r.mustAddRoute(rh)
	methods := []method.Method{}

	for m := range rt.Methods {
		methods = append(methods, m)
	}

	if len(methods) == 1 {
		rh.SetHandlerForMethods(handler, methods[0])
		return
	}
	rh.SetHandlerForMethods(handler, methods[0], methods[1:]...)
}

func (r *Router) HandleRouteFunc(rt *route.Route, handler http.HandlerFunc) {
	r.HandleRoute(rt, handler)
}

func (r *Router) HandleRouteMethods(rt *route.Route, handler http.Handler, method1 method.Method, furtherMethods ...method.Method) {
	mustNotBeRouter(handler)
	methods := append(furtherMethods, method1)

	for _, m := range methods {
		if !rt.HasMethod(m) {
			panic(&ErrMethodNotDefinedForRoute{m, rt})
		}
	}

	if _, has := r.routes[rt.DefinitionPath]; has {
		panic(ErrDoubleRegistration{rt.DefinitionPath})
	}

	rh := newRouteHandler(rt)
	r.mustAddRoute(rh)
	rh.SetHandlerForMethods(handler, method1, furtherMethods...)
}

func (r *Router) HandleRouteMethodsFunc(rt *route.Route, handler http.HandlerFunc, method1 method.Method, furtherMethods ...method.Method) {
	r.HandleRouteMethods(rt, handler, method1, furtherMethods...)
}

func (r *Router) handleMethods(path string, handler http.Handler, method1 method.Method, furtherMethods ...method.Method) *routeHandler {
	rt := r.newRoute(path, method1, furtherMethods...)
	rt.SetHandlerForMethods(handler, method1, furtherMethods...)
	return rt
}

func (r *Router) HandleMethods(path string, handler http.Handler, method1 method.Method, furtherMethods ...method.Method) *route.Route {
	mustNotBeRouter(handler)
	return r.handleMethods(path, handler, method1, furtherMethods...).Route
}

// Handle registers a handler for all routes. Use it to mount sub routers
func (r *Router) Handle(path string, handler http.Handler) {
	if rtr, is := handler.(*Router); is {
		if rtr.parent != nil {
			panic(ErrDoubleMounted{Path: rtr.path})
		}
		if rtr.path != "" {
			panic(ErrDoubleRegistration{DefinitionPath: rtr.path})
		}
		rtr.path = path
		rtr.parent = r
	}
	r.handleMethods(
		path, handler,
		method.GET,
		method.POST,
		method.PATCH,
		method.PUT,
		method.DELETE,
		method.OPTIONS,
	)
}

func (r *Router) EachRoute(fn func(mountPoint string, route *route.Route)) {
	for mP, rt := range r.routes {
		fn(mP, rt.Route)
	}
}

// private methods

func (ø *Router) findHandler(start, end int, req *http.Request, meth method.Method) (h http.Handler, rt *routeHandler) {
	if start == end {
		return
	}

	oldStart, oldEnd := start, end
	ln := len(ø.path)

	// trimming down the path
	if ln != 1 {
		if !strings.HasPrefix(req.URL.Path[start:end], ø.path) {
			return
		}

		if end-start == ln {
			end = start + 1
		} else {
			start += ln
		}
	}

	var parms *[]byte
	parms, rt = ø.node.FindPlaceholders(start, end, req)

	if rt == nil {
		return
	}

	h = rt.Handler(meth)

	if h == nil {
		return
	}

	if rtr, isRouter := h.(*Router); isRouter {
		return rtr.findHandler(oldStart, oldEnd, req, meth)
	}

	if parms == nil {
		req.URL.Fragment = rt.Id
		return
	}
	req.URL.Fragment = string(*parms) + rt.Id
	return
}

func (ø *Router) getHandler(rq *http.Request) (h http.Handler, rt *routeHandler, meth method.Method) {
	meth = method.Method(rq.Method)
	if meth == method.HEAD {
		meth = method.GET
	}

	h, rt = ø.findHandler(0, len(rq.URL.Path), rq, meth)
	return
}

// route not found boils down to method not allowed.
// I think this allows a better seperation  between a missing route (405, Method not allowed) and
// a missing entity (such as a missing page served by a cms or a missing entity requested via REST
// API call). Method not allowed errors (missing routes) should be tracked, because:
//
// - they might be programming errors (call of wrong path)
// - they might be attacking attempts (we might want to block calls on certain patterns and
//  block further requests from them)
//
// on the other hand, there is no need to track 404 response, simply return the answer to the client
// in the appropriate format

func (r *Router) setPath() {
	if r.parent == nil {
		r.path = path.Join("/", r.mountPoint)
		return
	}
	r.path = path.Join(r.parent.path, r.mountPoint)
}

func (r *Router) setPaths() {
	r.setPath()
	for _, rt := range r.routes {
		rt.EachHandler(func(h http.Handler) error {
			if rtr, has := h.(*Router); has {
				if err := rtr.submount(rt.DefinitionPath, r); err != nil {
					panic(err)
				}
				rtr.setPaths()
			}
			if fs, has := h.(*FileServer); has {
				if fs.Handler == nil {
					fs.SetHandler()
				}
			}
			return nil
		})
	}
}

func (r *Router) prepareRoutes() {
	for p, rt := range r.routes {
		missing := rt.MissingHandler()
		if len(missing) > 0 {
			panic(&ErrMissingHandler{missing, rt.Route})
		}
		r.node.add(p, rt)
	}
}

func (r *Router) submount(path string, parent *Router) error {
	if r.submounted && r.parent == parent {
		return nil
	}

	if bytes.IndexByte([]byte(path), route.PARAM_PREFIX) > -1 {
		return ErrInvalidMountPath{path, fmt.Sprintf("mount path must not contain wildcardseparator")}
	}
	if r.mountPoint != "" {
		return ErrDoubleMounted{path}
	}
	r.mountPoint = path
	r.parent = parent
	r.submounted = true
	r.prepareRoutes()
	return nil
}

func (r *Router) newRoute(path string, method1 method.Method, furtherMethods ...method.Method) *routeHandler {
	// methods := append(furtherMethods, method1)
	rt := r.routes[path]
	if rt == nil {
		rt = newRouteHandler(route.New(path, method1, furtherMethods...))
		// rt = route.New(path)
		r.mustAddRoute(rt)
	} else {
		methods := append(furtherMethods, method1)

		for _, m := range methods {
			if !rt.Route.HasMethod(m) {
				rt.Route.Methods[m] = struct{}{}
			}
		}
	}
	return rt
}

// MayMount mounts the router under the given path, i.e. all routing paths will be
// relative to this path.
// If parent is a *Router, the current router will be sub router of parent.
// If parent is a http.ServeMux its Handle method is used to mount the router.
// If parent is nil the router is self mounted and will be the main handler.
func (ø *Router) mayMount(path string, parent Muxer) error {
	if bytes.IndexByte([]byte(path), route.PARAM_PREFIX) > -1 {
		return ErrInvalidMountPath{path, fmt.Sprintf("path with wildcardseparator not allowed")}
	}

	if parentRtr, ok := parent.(*Router); ok {
		parentRtr.Handle(path, ø)
		return nil
	}

	if ø.mountPoint != "" {
		return ErrDoubleMounted{ø.path}
	}

	ø.mountPoint = path
	ø.setPaths()
	ø.prepareRoutes()

	if parent != nil {
		ø.muxed = true
		if path == "/" {
			parent.Handle("/", ø)
			return nil
		}

		parent.Handle(ø.path+"/", ø)
	}
	return nil
}
