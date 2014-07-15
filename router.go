package router

import (
	"net/http"
	"path"
	"strings"

	"github.com/go-on/method"
	"github.com/go-on/router/route"
	"github.com/go-on/wrap"
)

// Router is a mountable router routing paths to routes.
//
// Concurrently adding and serving routes is not supported.
// Routes must be defined none concurrently and before serving
type Router struct {
	node       *node
	wrapper    []wrap.Wrapper
	routes     map[string]*route.Route
	parent     *Router
	mountPoint string
	path       string
	muxed      bool
	// NotFound is called, if a http.Handler could not be found.
	// If it is set to nil, the status 405 is set
	NotFound http.Handler
}

// New creates a new bare router
func New() *Router {
	return &Router{
		routes: map[string]*route.Route{},
		node:   newNode(),
	}
}

func (ø *Router) MustAdd(rt *route.Route) {
	err := ø.Add(rt)
	if err != nil {
		panic(err.Error())
	}
}

// the given wrappers are near the inner call and called before the
// etag and IfMatch and IfNoneMatch wrappers. wrappers around them
// could be easily done by making a go-on/wrap.New() and use the Router
// as final http.Handler surrounded by other middleware
func (r *Router) AddWrappers(wrapper ...wrap.Wrapper) {
	r.wrapper = append(r.wrapper, wrapper...)
}

func (ø *Router) Add(rt *route.Route) error {
	if _, has := ø.routes[rt.DefinitionPath]; has {
		return ErrDoubleRegistration{rt.DefinitionPath}
	}
	ø.routes[rt.DefinitionPath] = rt
	rt.Router = ø
	return nil
}

func (ø *Router) MountPath() string { return ø.path }

func (ø *Router) RequestRoute(rq *http.Request) (rt *route.Route) {
	_, rt, _ = ø.getHandler(rq)
	return
}

// SetOPTIONSHandlers sets the OPTIONSHandler for all routes of the router.
// If routes of subrouter could be set via SetAllOPTIONSHandlers
func (r *Router) SetOPTIONSHandlers() {
	for _, rt := range r.routes {
		SetOPTIONSHandler(rt)
	}
}

func setAllOPTIONSHandlers(h http.Handler) error {
	if rtr, ok := h.(*Router); ok {
		rtr.SetAllOPTIONSHandlers()
	}
	return nil
}

// SetAllOPTIONSHandlers is like SetOPTIONSHandlers but also does so recursively in subrouters
func (r *Router) SetAllOPTIONSHandlers() {
	for _, rt := range r.routes {
		SetOPTIONSHandler(rt)
		rt.EachHandler(setAllOPTIONSHandlers)
	}
}

// ServeOPTIONS serves OPTIONS request by setting the Allow header for all
// defined methods of the request handling route.
// Performance could be improved, if the OPTIONSHandler is set for the route.
// This could be done via SetOPTIONSHandlers and SetAllOPTIONSHandlers
func (r *Router) ServeOPTIONS(w http.ResponseWriter, rq *http.Request) {
	h, rt := r.findHandler(0, len(rq.URL.Path), rq, method.OPTIONS)
	if h != nil {
		h.ServeHTTP(w, rq)
		return
	}
	w.Header().Set("Allow", optionsString(rt))
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
	if r.mountPoint == "" {
		panic(ErrNotMounted{})
	}
	h := r.Dispatch(rq)
	if h == nil {
		if h = r.NotFound; h == nil {
			w.WriteHeader(405)
			return
		}
	}
	h.ServeHTTP(w, rq)
}

// Dispatch returns the corresponding http.Handler for the request
func (ø *Router) Dispatch(rq *http.Request) http.Handler {
	h, rt, meth := ø.getHandler(rq)

	if h == nil {
		return nil
	}

	if meth != method.OPTIONS {
		return rt.Router.(*Router).wrapHandler(h)
	}
	return h
}

// Mount mounts the router under the given path, i.e. all routing paths will be
// relative to this path. If a ServeMux is given, its Handle method is used to mount
// the router. Otherwise the router is self mounted and will be the main handler.
func (ø *Router) Mount(path string, m *http.ServeMux) error {
	if strings.Index(path, ":") > -1 {
		return ErrInvalidMountPath{path, "path with ':' not allowed"}
	}

	if ø.mountPoint != "" {
		return ErrDoubleMounted{ø.path}
	}

	// fmt.Printf("setting mountpoint of %p to %#v\n", ø, path)
	ø.mountPoint = path
	ø.setPaths()
	ø.prepareRoutes()

	if m != nil {
		ø.muxed = true
		if path == "/" {
			m.Handle("/", ø)
			return nil
		}

		m.Handle(ø.path+"/", ø)
	}
	return nil
}

func (r *Router) MustMount(path string, m *http.ServeMux) {
	err := r.Mount(path, m)
	if err != nil {
		panic(err)
	}
}

// private methods

func (ø *Router) findHandler(start, end int, req *http.Request, meth method.Method) (h http.Handler, rt *route.Route) {
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

func (ø *Router) wrapHandler(h http.Handler) http.Handler {
	for i := len(ø.wrapper) - 1; i >= 0; i-- {
		h = ø.wrapper[i].Wrap(h)
	}
	if ø.parent != nil {
		return ø.parent.wrapHandler(h)
	}
	return h
}

func (ø *Router) getHandler(rq *http.Request) (h http.Handler, rt *route.Route, meth method.Method) {
	meth = method.StringToMethod[rq.Method]
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
					panic(err.Error())
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
		r.node.add(p, rt)
	}
}

func (r *Router) submount(path string, parent *Router) error {
	if r.parent == parent {
		return nil
	}
	if strings.Index(path, ":") > -1 {
		return ErrInvalidMountPath{path, "mount path must not contain ':'"}
	}
	if r.mountPoint != "" {
		return ErrDoubleMounted{path}
	}
	r.mountPoint = path
	r.parent = parent
	r.prepareRoutes()
	return nil
}

func (ø *Router) setupHandlersX(rt *route.Route) func(h http.Handler) error {
	return func(h http.Handler) error {
		if r, has := h.(*Router); has {
			if err := r.submount(rt.DefinitionPath, ø); err != nil {
				return err
			}
		}
		if fs, has := h.(*FileServer); has {
			if fs.Handler == nil {
				fs.SetHandler()
			}
		}
		return nil
	}
}

// Handle registers a GET route
func (r *Router) Handle(path string, handler http.Handler) {
	rt := route.NewRoute(path)
	rt.GETHandler = handler
	r.MustAdd(rt)
}

func (r *Router) newRoute(path string) *route.Route {
	rt := r.routes[path]
	if rt == nil {
		rt = route.NewRoute(path)
		r.MustAdd(rt)
	}
	return rt
}

func (r *Router) MustHandle(path string, m method.Method, handler http.Handler) *route.Route {
	rt := r.newRoute(path)
	rt.SetHandler(m, handler)
	return rt
}
