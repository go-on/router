package router

import (
	"fmt"
	"github.com/go-on/menu"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-on/method"
	"github.com/go-on/wrap"
	// "github.com/go-on/wrap-contrib-testing/wrapstesting"
	"github.com/go-on/wrap-contrib/wraps"
)

type Router struct {
	pathNode   *pathNode
	wrapper    []wrap.Wrapper
	routes     map[string]*Route
	parent     *Router
	mountPoint string
}

//func New(wrapper ...wrap.Wrapper) (ø *Router) {
func New() (ø *Router) {
	ø = &Router{
		wrapper:  []wrap.Wrapper{},
		routes:   map[string]*Route{},
		pathNode: newPathNode(),
	}
	return
}

func NewETagged() (ø *Router) {
	ø = New()
	ø.AddWrappers(
		wraps.IfNoneMatch,
		wraps.IfMatch(ø),
		wraps.ETag,
	)
	return
}

func (r *Router) Route(path string) *Route {
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

func (ø *Router) getFinalHandler(path, method string) (h http.Handler, route *Route, wc map[string]string) {
	var leaf *pathLeaf
	leaf, wc = ø.findLeaf(path)
	if leaf == nil {
		return
	}

	h = leaf.getHandler(method)

	if h == nil {
		return
	}

	route = leaf.Route

	rt, isRouter := h.(*Router)
	if isRouter {
		return rt.getFinalHandler(path, method)
	}

	// fmt.Println("method", method, "h", h)
	if h == nil && method == "OPTIONS" {
		h = route
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

func (ø *Router) serveHTTP(w http.ResponseWriter, rq *http.Request) {
	method := rq.Method
	h, route, wc := ø.getFinalHandler(rq.URL.Path, method)

	if h == nil && method == "HEAD" {
		method = "GET"
		h, route, wc = ø.getFinalHandler(rq.URL.Path, method)
	}

	if h == nil {
		ø.serveNotFound(w, rq)
		return
	}

	if method != "OPTIONS" {
		h = route.router.wrapit(h)
	}

	q := rq.URL.Query()
	for k, v := range wc {
		q.Set(":"+k, v)
	}
	rq.URL.RawQuery = q.Encode()
	rq.URL.Fragment = route.originalPath

	//h.ServeHTTP(&Vars{w, wc}, rq)
	h.ServeHTTP(w, rq)
	return
}

func (ø *Router) ServeHTTP(w http.ResponseWriter, rq *http.Request) {
	if ø.mountPoint == "" {
		panic("router not mounted")
	}
	wraps.MethodOverride().ServeHandle(http.HandlerFunc(ø.serveHTTP), w, rq)
}

func (ø *Router) findLeaf(url string) (leaf *pathLeaf, wc map[string]string) {
	if url == "" || !filepath.HasPrefix(url, ø.Path()) {
		return
	}

	leaf, wc = ø.pathNode.Match(ø.trimmedUrl(url))
	if leaf == nil {
		i := strings.LastIndex(url, "/")
		if i >= 0 {
			return ø.findLeaf(url[:i])
		}
	}
	return
}

// route not found boils down to method not allowed.
// I find this allows a better seperation  between a missing route (405, Method not allowed) and
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
	w.Header().Set("Allow", "")
	w.WriteHeader(405)
}

func (r *Router) Path() string {
	if r.parent == nil {
		return path.Join("/", r.mountPoint)
	} else {
		return path.Join(r.parent.Path(), r.mountPoint)
	}
}

func (r *Router) trimmedUrl(url string) (trimmed string) {
	tr, err := filepath.Rel(r.Path(), url)
	if err != nil {
		panic(err.Error())
	}

	return filepath.Clean("/" + tr)
}

func (ø *Router) Mount(path string, m *http.ServeMux) error {
	if ø.mountPoint != "" {
		return fmt.Errorf("already mounted on %s", ø.Path())
	}

	ø.mountPoint = path
	err := ø.registerRoutes()

	if err != nil {
		return err
	}

	if path == "/" {
		m.Handle("/", ø)
		return nil
	}

	// fmt.Printf("mount %s\n", ø.Path()+"/")
	m.Handle(ø.Path()+"/", ø)
	return nil
}

func (r *Router) MustMount(path string, m *http.ServeMux) {
	err := r.Mount(path, m)
	if err != nil {
		panic(err.Error())
	}
}

func (r *Router) registerRoutes() error {
	for p, rt := range r.routes {
		for v, h := range rt.handler {
			err := r.pathNode.add(p, v, h, r)
			if err != nil {
				return fmt.Errorf("can't register %s %s", v.String(), p)
			}
		}
	}
	return nil
}

func (r *Router) submount(path string, parent *Router) error {
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

func (ø *Router) Handle(path string, v method.Method, handler http.Handler) (*Route, error) {
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
		rt = newRoute(ø, path)
	}
	rtr, is := handler.(*Router)
	if is {
		err := rtr.submount(path, ø)
		if err != nil {
			return nil, err
		}
	}

	err := rt.addHandler(handler, v)
	if err != nil {
		return nil, err
	}

	ø.routes[path] = rt
	return rt, err
}

func (r *Router) MustHandle(path string, v method.Method, handler http.Handler) *Route {
	rt, err := r.Handle(path, v, handler)
	if err != nil {
		panic(err.Error())
	}
	return rt
}

func (r *Router) GET(path string, handler http.Handler) *Route {
	return r.MustHandle(path, method.GET, handler)
}
func (r *Router) POST(path string, handler http.Handler) *Route {
	return r.MustHandle(path, method.POST, handler)
}
func (r *Router) PUT(path string, handler http.Handler) *Route {
	return r.MustHandle(path, method.PUT, handler)
}
func (r *Router) DELETE(path string, handler http.Handler) *Route {
	return r.MustHandle(path, method.DELETE, handler)
}
func (r *Router) PATCH(path string, handler http.Handler) *Route {
	return r.MustHandle(path, method.PATCH, handler)
}
func (r *Router) OPTIONS(path string, handler http.Handler) *Route {
	return r.MustHandle(path, method.OPTIONS, handler)
}
func (r *Router) HEAD(path string, handler http.Handler) *Route {
	return r.MustHandle(path, method.HEAD, handler)
}
func (r *Router) TRACE(path string, handler http.Handler) *Route {
	return r.MustHandle(path, method.TRACE, handler)
}

func (r *Router) EachRoute(fn func(mountPoint string, route *Route)) {
	for mP, rt := range r.routes {
		fn(mP, rt)
	}
}

func (r *Router) EachGETRoute(fn func(mountPoint string, route *Route)) {
	for mP, rt := range r.routes {
		if rt.getHandler("GET") != nil {
			fn(mP, rt)
		}
	}
}

type RouteParameterFunc func(*Route) []map[string]string

func (rpf RouteParameterFunc) Params(rt *Route) []map[string]string {
	return rpf(rt)
}

type RouteParameter interface {
	Params(*Route) []map[string]string
}

type MenuParameter interface {
	Params(*Route) []map[string]string

	// Text returns the menu text for the given route with the given
	// parameters
	Text(rt *Route, params map[string]string) string
}

type MenuAdder interface {
	// Add adds the given item somewhere. Where might be decided
	// by looking at the given route
	Add(item menu.Leaf, rt *Route, params map[string]string)
}

// Menu creates a menu item for each route via solver
// and adds it via appender
func (r *Router) Menu(adder MenuAdder, solver MenuParameter) {
	fn := func(mountPoint string, rt *Route) {
		if rt.HasParams() {
			paramsArr := solver.Params(rt)
			for _, params := range paramsArr {
				adder.Add(
					menu.Item(solver.Text(rt, params), rt.MustURLMap(params)),
					rt,
					params,
				)
			}

		} else {
			adder.Add(
				menu.Item(solver.Text(rt, nil), rt.MustURL()),
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
	fn := func(mountPoint string, rt *Route) {

		if rt.HasParams() {
			paramsArr := paramSolver.Params(rt)

			for _, params := range paramsArr {
				paths = append(paths, rt.MustURLMap(params))
			}

		} else {
			paths = append(paths, rt.MustURL())
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
func (r *Router) GETPathsByStruct(parameters map[*Route]map[string][]interface{}) (paths []string) {
	paths = []string{}

	fn := func(mountPoint string, route *Route) {
		paramPairs := parameters[route]

		// if route has : it has parameters
		if route.HasParams() {
			for tag, structs := range paramPairs {
				for _, stru := range structs {
					paths = append(paths, route.MustURLStruct(stru, tag))
				}
			}
		} else {
			paths = append(paths, route.MustURL())
		}
	}

	r.EachGETRoute(fn)
	return
}

func (r *Router) DynamicRoutes() (routes []*Route) {
	routes = []*Route{}
	for _, rt := range r.routes {
		if rt.HasParams() {
			routes = append(routes, rt)
		}
	}
	return routes
}

func (r *Router) StaticRoutePaths() (paths []string) {
	paths = []string{}
	for _, rt := range r.routes {
		if !rt.HasParams() {
			paths = append(paths, rt.MustURL())
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
