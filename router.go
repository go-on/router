package router

import (
	"fmt"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-on/lib/internal/menu"

	"github.com/go-on/method"
	"github.com/go-on/router/route"
	"github.com/go-on/wrap"
	"github.com/go-on/wrap-contrib/wraps"
)

// concurrently adding and serving routes is not supported.
// routes must be defined none concurrently and before serving
type Router struct {
	pathNode   *pathNode
	wrapper    []wrap.Wrapper
	routes     map[string]*route.Route
	parent     *Router
	mountPoint string
	path       string
	muxed      bool
}

// New creates a bare router, without method override middleware
func New() (ø *Router) {
	ø = &Router{
		wrapper:  []wrap.Wrapper{},
		routes:   map[string]*route.Route{},
		pathNode: newPathNode(),
	}
	return
}

func ETagged() (ø *Router) {
	ø = New()
	ø.AddWrappers(
		wraps.IfNoneMatch,
		wraps.IfMatch(ø),
		wraps.ETag,
	)
	return
}

func (r *Router) Route(path string) *route.Route {
	return r.routes[path]
}

// the given wrappers are near the inner call and called before the
// etag and IfMatch and IfNoneMatch wrappers. wrappers around them
// could be easily done by making a go-on/wrap.New() and use the Router
// as final http.Handler surrounded by other middleware
func (r *Router) AddWrappers(wrapper ...wrap.Wrapper) {
	r.wrapper = append(r.wrapper, wrapper...)
}

func (r *Router) Parent() *Router {
	return r.parent
}

func (r *Router) MountPoint() string {
	return r.mountPoint
}

func (ø *Router) getFinalHandler(path string, meth method.Method, rq *http.Request) (h http.Handler, route *route.Route) {
	if len(path) == 0 || !filepath.HasPrefix(path, ø.Path()) {
		return
	}

	start, end := ø.trimmedUrl(path)
	wc := &wildcards{path: path[start:end]}
	// return
	ø.pathNode.FindPlaceholders(wc)
	// _ = wc
	// return nil, wc.route

	if wc.route == nil {
		return
	}

	switch meth {
	case method.GET:
		h = wc.route.GETHandler
	case method.POST:
		h = wc.route.POSTHandler
	case method.PUT:
		h = wc.route.PUTHandler
	case method.PATCH:
		h = wc.route.PATCHHandler
	case method.DELETE:
		h = wc.route.DELETEHandler
	case method.OPTIONS:
		h = wc.route.OPTIONSHandler
	}

	if h == nil {
		return
	}

	route = wc.route

	rt, isRouter := h.(*Router)
	if isRouter {
		return rt.getFinalHandler(path, meth, rq)
	}

	if len(wc.found) > 0 {
		rq.URL.Fragment = wc.route.OriginalPath + "//" + wc.ParamStr()
	}

	return
}

func (ø *Router) wrapit(h http.Handler) http.Handler {
	for i := len(ø.wrapper) - 1; i >= 0; i-- {
		h = ø.wrapper[i].Wrap(h)
	}
	if ø.parent != nil {
		return ø.parent.wrapit(h)
	}
	return h
}

func (ø *Router) MountPath() string {
	return ø.Path()
}

/*
func SetRouteParam(req *http.Request, key, value string) {
	req.Header.Set(fmt.Sprintf("X-Route-Param-%s", key), value)
}
*/

// since req.URL.Path has / unescaped so that originally escaped / are
// indistinguishable from escaped ones, we are save here, i.e. / is
// already handled as path splitted and no key or value has / in it
// also it is save to use req.URL.Fragment since that will never be transmitted
// by the request
func GetRouteParam(req *http.Request, key string) string {
	i := strings.Index(req.URL.Fragment, "//"+key+"/")
	if i == -1 {
		return ""
	}
	startId := i + 3 + len(key)
	end := strings.Index(req.URL.Fragment[startId:], "//")
	return req.URL.Fragment[startId : startId+end]
}

// also it is save to use req.URL.Fragment since that will never be transmitted
// by the request
func GetRouteRelPath(req *http.Request) string {
	i := strings.Index(req.URL.Fragment, "//")
	if i == -1 {
		return ""
	}
	return req.URL.Fragment[:i]
}

func (ø *Router) RequestRoute(rq *http.Request) (rt *route.Route) {
	_, rt, _ = ø.getHandler(rq)
	return
}

func (ø *Router) getHandler(rq *http.Request) (h http.Handler, rt *route.Route, meth method.Method) {
	meth = method.StringToMethod[rq.Method]
	if meth == method.HEAD {
		meth = method.GET
	}

	h, rt = ø.getFinalHandler(rq.URL.Path, meth, rq)
	return
}

func (ø *Router) Dispatch(rq *http.Request) http.Handler {
	h, route, meth := ø.getHandler(rq)

	if h == nil {
		return nil
	}

	if meth != method.OPTIONS {
		h = route.Router.(*Router).wrapit(h)
	}

	return h
}

// this handler should be used for the top level router
// it handles method override and cleanpath
func (ø *Router) Serve(w http.ResponseWriter, rq *http.Request) {
	if !ø.muxed {
		if rq.RequestURI == "*" {
			if rq.ProtoAtLeast(1, 1) {
				w.Header().Set("Connection", "close")
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// taken from net/http ServeMux and modified
		if rq.Method != "CONNECT" {
			if p := cleanPath(rq.URL.Path); p != rq.URL.Path {
				url := *rq.URL
				url.Path = p
				http.RedirectHandler(url.String(), http.StatusMovedPermanently).ServeHTTP(w, rq)
				return
			}
		}
	}
	// we can't handle the method override as part of the wraps, because it has to
	// be run before we look for the method (or we would have to run all wrappers before)
	// maybe we should not handle this case since it can be handled be the outside
	wraps.MethodOverride().ServeHTTP(w, rq)
	ø.ServeHTTP(w, rq)
}

// if server is nil, the default server is used
func (ø *Router) ListenAndServe(addr string, server *http.Server) error {
	if server == nil {
		return http.ListenAndServe(addr, http.HandlerFunc(ø.Serve))
	}
	server.Addr = addr
	server.Handler = http.HandlerFunc(ø.Serve)
	return server.ListenAndServe()
	// http.ListenAndServeTLS(addr, certFile, keyFile, handler)
	// server.ListenAndServeTLS(certFile, keyFile)
}

func (ø *Router) ListenAndServeTLS(addr string, certFile string, keyFile string, server *http.Server) error {
	if server == nil {
		return http.ListenAndServeTLS(addr, certFile, keyFile, http.HandlerFunc(ø.Serve))
	}
	server.Addr = addr
	server.Handler = http.HandlerFunc(ø.Serve)
	return server.ListenAndServeTLS(certFile, keyFile)
}

// if you want to use the top level router, use Serve. that does some
// stuff that only should be done once per request
func (ø *Router) ServeHTTP(w http.ResponseWriter, rq *http.Request) {
	if ø.mountPoint == "" {
		panic("router not mounted")
	}
	h := ø.Dispatch(rq)
	if h == nil {
		ø.serveNotFound(w, rq)
		return
	}
	h.ServeHTTP(w, rq)
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
func (r *Router) serveNotFound(w http.ResponseWriter, rq *http.Request) {
	// w.Header().Set("Allow", "")
	w.WriteHeader(405)
	// w.WriteHeader(405)
}

func (r *Router) Path() string {
	return r.path
}

func (r *Router) setPath() {
	if r.parent == nil {
		r.path = path.Join("/", r.mountPoint)
	} else {
		r.path = path.Join(r.parent.Path(), r.mountPoint)
	}
	// fmt.Printf("setting path of %p to %#v\n", r, r.path)
}

func (r *Router) trimmedUrl(url string) (startPos, endPos int) {
	if len(r.Path()) == 1 {
		return 0, len(url)
	}
	if !strings.HasPrefix(url, r.Path()) {
		panic(fmt.Sprintf("url %#v not relative to %#v", url, r.Path()))
	}
	if len(url) == len(r.Path()) {
		return 0, 1
	}
	return len(r.Path()), len(url)
}

func (r *Router) setPaths() {
	r.setPath()
	//for p, rt := range r.routes {
	for _, rt := range r.routes {
		if rtr, ok := rt.GETHandler.(*Router); ok {
			/*
				err := rtr.submount(p, r)
				if err != nil {
					panic(err.Error())
				}
			*/
			rtr.setPaths()
		}
		if rtr, ok := rt.DELETEHandler.(*Router); ok {
			/*
				err := rtr.submount(p, r)
				if err != nil {
					panic(err.Error())
				}
			*/
			rtr.setPaths()
		}
		if rtr, ok := rt.POSTHandler.(*Router); ok {
			/*
				err := rtr.submount(p, r)
				if err != nil {
					panic(err.Error())
				}
			*/
			// rtr.submount(p, r)
			rtr.setPaths()
		}
		if rtr, ok := rt.PATCHHandler.(*Router); ok {
			/*
				err := rtr.submount(p, r)
				if err != nil {
					panic(err.Error())
				}
			*/
			rtr.setPaths()
		}
		if rtr, ok := rt.PUTHandler.(*Router); ok {
			/*
				err := rtr.submount(p, r)
				if err != nil {
					panic(err.Error())
				}
			*/
			rtr.setPaths()
		}
		if rtr, ok := rt.OPTIONSHandler.(*Router); ok {
			/*
				err := rtr.submount(p, r)
				if err != nil {
					panic(err.Error())
				}
			*/
			rtr.setPaths()
		}
	}
}

func (ø *Router) Mount(path string, m *http.ServeMux) error {
	if strings.Index(path, ":") > -1 {
		return fmt.Errorf("mount on path with vars not allowed")
	}
	if ø.mountPoint != "" {
		return fmt.Errorf("already mounted on %s", ø.Path())
	}

	// fmt.Printf("setting mountpoint of %p to %#v\n", ø, path)
	ø.mountPoint = path
	ø.setPaths()
	err := ø.registerRoutes()

	if err != nil {
		return err
	}

	if m != nil {
		ø.muxed = true
		if path == "/" {
			m.Handle("/", ø)
			return nil
		}

		m.Handle(ø.Path()+"/", ø)
	}
	return nil
}

func (r *Router) MustMount(path string, m *http.ServeMux) {
	err := r.Mount(path, m)
	if err != nil {
		panic(err.Error())
	}
}

/*
func (r *Router) subMountRoute(p string, rt *route.Route) error {
	fmt.Println("subMountRoute")
	fmt.Printf("get handler: %T\n", rt.GETHandler)
	if rtr, is := rt.GETHandler.(*Router); is {
		err := rtr.submount(p, r)
		if err != nil {
			return err
		}
	}
	return nil
}
*/

func (r *Router) registerRoutes() error {
	// fmt.Println("registerRoutes")
	for p, rt := range r.routes {
		/*
			err := r.subMountRoute(p, rt)
			if err != nil {
				return err
			}
		*/
		if rt.OPTIONSHandler == nil {
			setOptionsHandler(rt)
		}
		r.pathNode.add(p, rt)
	}
	return nil
}

func (r *Router) submount(path string, parent *Router) error {
	// fmt.Printf("submounting %p on path: %#v\n", r, path)
	if strings.Index(path, ":") > -1 {
		return fmt.Errorf("submount on path with vars not allowed")
	}
	if r.mountPoint != "" {
		return fmt.Errorf("already mounted on %s", r.Path())
	}
	r.mountPoint = path
	r.parent = parent
	return r.registerRoutes()
}

// wrapper around a http.Handler generator to make it a http.Handler
type RouterFunc func() http.Handler

func (hc RouterFunc) ServeHTTP(rw http.ResponseWriter, req *http.Request) { hc().ServeHTTP(rw, req) }

func (ø *Router) MustRegisterRoute(rt *route.Route, v method.Method, handler http.Handler) {
	err := ø.RegisterRoute(rt, v, handler)

	if err != nil {
		panic(err)
	}

}

func (ø *Router) RegisterRoute(rt *route.Route, v method.Method, handler http.Handler) error {
	rt.Router = ø
	rt2, has := ø.routes[rt.OriginalPath]
	if has && rt != rt2 {
		return fmt.Errorf("path %#v already has another route", rt.OriginalPath)
	}
	if !has {
		ø.routes[rt.OriginalPath] = rt
	}
	//= rt
	return ø.assocHandler(rt, v, handler)
}

/*
func (ø *Router) submountMe(rt *route.Route) {
	}
*/

func (ø *Router) MustRegisterRoute2(rt *route.Route) {
	_, has := ø.routes[rt.OriginalPath]
	if has {
		panic(fmt.Sprintf("path %#v already has another route", rt.OriginalPath))
	}
	ø.routes[rt.OriginalPath] = rt
	rt.Router = ø

	if r, has := rt.GETHandler.(*Router); has {
		err := r.submount(rt.OriginalPath, ø)
		if err != nil {
			panic(err.Error())
		}
	}

	if r, has := rt.POSTHandler.(*Router); has {
		err := r.submount(rt.OriginalPath, ø)
		if err != nil {
			panic(err.Error())
		}
	}

	if r, has := rt.PUTHandler.(*Router); has {
		err := r.submount(rt.OriginalPath, ø)
		if err != nil {
			panic(err.Error())
		}
	}

	if r, has := rt.DELETEHandler.(*Router); has {
		err := r.submount(rt.OriginalPath, ø)
		if err != nil {
			panic(err.Error())
		}
	}

	if r, has := rt.PATCHHandler.(*Router); has {
		err := r.submount(rt.OriginalPath, ø)
		if err != nil {
			panic(err.Error())
		}
	}

	//return ø.assocHandler(rt, v, handler)
}

func (ø *Router) assocHandler(rt *route.Route, m method.Method, handler http.Handler) error {
	rt.Router = ø

	rtr, is := handler.(*Router)
	if is {
		err := rtr.submount(rt.OriginalPath, ø)
		if err != nil {
			return err
		}
	}

	switch m {
	case method.GET:
		rt.GETHandler = handler
	case method.PUT:
		rt.PUTHandler = handler
	case method.POST:
		rt.POSTHandler = handler
	case method.DELETE:
		rt.DELETEHandler = handler
	case method.PATCH:
		rt.PATCHHandler = handler
	case method.OPTIONS:
		rt.OPTIONSHandler = handler
	default:
		if m&method.GET != 0 {
			rt.GETHandler = handler
		}
		if m&method.PUT != 0 {
			rt.PUTHandler = handler
		}
		if m&method.POST != 0 {
			rt.POSTHandler = handler
		}
		if m&method.DELETE != 0 {
			rt.DELETEHandler = handler
		}
		if m&method.PATCH != 0 {
			rt.PATCHHandler = handler
		}
		if m&method.OPTIONS != 0 {
			rt.OPTIONSHandler = handler
		}
	}

	/*
		err := rt.AddHandler(handler, v)
		if err != nil {
			return err
		}
	*/
	// ø.routes[rt.OriginalPath] = rt
	return nil
}

func (ø *Router) Handle(path string, handler http.Handler) {
	ø.GET(path, handler)
}

// Handle(mountpoint string, h http.Handler)

func (ø *Router) HandleMethod(path string, v method.Method, handler http.Handler) (*route.Route, error) {
	if ø.mountPoint != "" {
		return nil, fmt.Errorf("can't register handlers: already mounted on %s", ø.Path())
	}

	rt, exists := ø.routes[path]
	/*
		if exists && rt.getHandler(v.String()) != nil {
			panic(fmt.Sprintf("handler for %s (%s) already exists", path, v))
		}
	*/
	if !exists {
		rt = route.NewRoute(path)
		rt.Router = ø
		ø.routes[path] = rt
	}

	err := ø.assocHandler(rt, v, handler)
	return rt, err
}

func (r *Router) MustHandleMethod(path string, v method.Method, handler http.Handler) *route.Route {
	rt, err := r.HandleMethod(path, v, handler)
	if err != nil {
		panic(err.Error())
	}
	return rt
}

func (r *Router) GET(path string, handler http.Handler) *route.Route {
	return r.MustHandleMethod(path, method.GET, handler)
}

func (r *Router) GETFunc(path string, handler http.HandlerFunc) *route.Route {
	return r.GET(path, handler)
}

func (r *Router) POST(path string, handler http.Handler) *route.Route {
	return r.MustHandleMethod(path, method.POST, handler)
}

func (r *Router) POSTFunc(path string, handler http.HandlerFunc) *route.Route {
	return r.POST(path, handler)
}

func (r *Router) PUT(path string, handler http.Handler) *route.Route {
	return r.MustHandleMethod(path, method.PUT, handler)
}

func (r *Router) PUTFunc(path string, handler http.HandlerFunc) *route.Route {
	return r.PUT(path, handler)
}

func (r *Router) DELETE(path string, handler http.Handler) *route.Route {
	return r.MustHandleMethod(path, method.DELETE, handler)
}

func (r *Router) DELETEFunc(path string, handler http.HandlerFunc) *route.Route {
	return r.DELETE(path, handler)
}

func (r *Router) PATCH(path string, handler http.Handler) *route.Route {
	return r.MustHandleMethod(path, method.PATCH, handler)
}

func (r *Router) PATCHFunc(path string, handler http.HandlerFunc) *route.Route {
	return r.PATCH(path, handler)
}

func (r *Router) OPTIONS(path string, handler http.Handler) *route.Route {
	return r.MustHandleMethod(path, method.OPTIONS, handler)
}

func (r *Router) OPTIONSFunc(path string, handler http.HandlerFunc) *route.Route {
	return r.OPTIONS(path, handler)
}

func (r *Router) HEAD(path string, handler http.Handler) *route.Route {
	return r.MustHandleMethod(path, method.HEAD, handler)
}

func (r *Router) HEADFunc(path string, handler http.HandlerFunc) *route.Route {
	return r.HEAD(path, handler)
}

func (r *Router) TRACE(path string, handler http.Handler) *route.Route {
	return r.MustHandleMethod(path, method.TRACE, handler)
}

func (r *Router) TRACEFunc(path string, handler http.HandlerFunc) *route.Route {
	return r.TRACE(path, handler)
}

func (r *Router) EachRoute(fn func(mountPoint string, route *route.Route)) {
	for mP, rt := range r.routes {
		fn(mP, rt)
	}
}

func (r *Router) EachGETRoute(fn func(mountPoint string, route *route.Route)) {
	for mP, rt := range r.routes {
		if rt.GETHandler != nil {
			fn(mP, rt)
		}
		/*
			if getHandler(rt, "GET") != nil {
				fn(mP, rt)
			}
		*/
	}
}

type RouteParameterFunc func(*route.Route) []map[string]string

func (rpf RouteParameterFunc) Params(rt *route.Route) []map[string]string {
	return rpf(rt)
}

type RouteParameter interface {
	Params(*route.Route) []map[string]string
}

type MenuParameter interface {
	Params(*route.Route) []map[string]string

	// Text returns the menu text for the given route with the given
	// parameters
	Text(rt *route.Route, params map[string]string) string
}

type MenuAdder interface {
	// Add adds the given item somewhere. Where might be decided
	// by looking at the given route
	Add(item menu.Leaf, rt *route.Route, params map[string]string)
}

// Menu creates a menu item for each route via solver
// and adds it via appender
func (r *Router) Menu(adder MenuAdder, solver MenuParameter) {
	fn := func(mountPoint string, rt *route.Route) {
		if HasParams(rt) {
			paramsArr := solver.Params(rt)
			for _, params := range paramsArr {
				adder.Add(
					menu.Item(solver.Text(rt, params), MustURLMap(rt, params)),
					rt,
					params,
				)
			}

		} else {
			adder.Add(
				menu.Item(solver.Text(rt, nil), MustURL(rt)),
				rt,
				nil,
			)
		}
	}
	r.EachGETRoute(fn)
}

// the paths of all get routes
func (r *Router) AllGETPaths(paramSolver RouteParameter) (paths []string) {
	paths = []string{}
	fn := func(mountPoint string, rt *route.Route) {

		if HasParams(rt) {
			paramsArr := paramSolver.Params(rt)

			for _, params := range paramsArr {
				paths = append(paths, MustURLMap(rt, params))
			}

		} else {
			paths = append(paths, MustURL(rt))
		}
	}

	r.EachGETRoute(fn)
	return paths
}

// saves the results of all get routes
func (r *Router) SavePages(paramSolver RouteParameter, mainHandler http.Handler, targetDir string) map[string]error {
	return DumpPaths(mainHandler, r.AllGETPaths(paramSolver), targetDir)
}

func (r *Router) MustSavePages(paramSolver RouteParameter, mainHandler http.Handler, targetDir string) {
	errs := r.SavePages(paramSolver, mainHandler, targetDir)
	for _, err := range errs {
		panic(err.Error())
	}
}

// map[string][]interface{} is tag => []struct
func (r *Router) GETPathsByStruct(parameters map[*route.Route]map[string][]interface{}) (paths []string) {
	paths = []string{}

	fn := func(mountPoint string, route *route.Route) {
		paramPairs := parameters[route]

		// if route has : it has parameters
		if HasParams(route) {
			for tag, structs := range paramPairs {
				for _, stru := range structs {
					paths = append(paths, MustURLStruct(route, stru, tag))
				}
			}
		} else {
			paths = append(paths, MustURL(route))
		}
	}

	r.EachGETRoute(fn)
	return
}

func (r *Router) DynamicRoutes() (routes []*route.Route) {
	routes = []*route.Route{}
	for _, rt := range r.routes {
		if HasParams(rt) {
			routes = append(routes, rt)
		}
	}
	return routes
}

func (r *Router) StaticRoutePaths() (paths []string) {
	paths = []string{}
	for _, rt := range r.routes {
		if !HasParams(rt) {
			paths = append(paths, MustURL(rt))
		}
	}
	return paths
}

func Mount(path string, r *Router) error { return r.Mount(path, http.DefaultServeMux) }

func MustMount(path string, r *Router) {
	err := Mount(path, r)
	if err != nil {
		panic(err.Error())
	}
}
