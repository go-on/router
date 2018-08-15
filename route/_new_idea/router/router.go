package router

import (
	// "gopkg.in/go-on/method.v1"
	"net/http"
	"path"
	"strings"
	"sync"
)

type Router struct {
	getRoutes    map[string]http.Handler
	onceGet      sync.Once
	postRoutes   map[string]http.Handler
	oncePost     sync.Once
	deleteRoutes map[string]http.Handler
	onceDelete   sync.Once
	patchRoutes  map[string]http.Handler
	oncePatch    sync.Once
	putRoutes    map[string]http.Handler
	oncePut      sync.Once
	subRouter    map[string]*Router
	onceSub      sync.Once
	mountPoint   string
	parent       *Router
}

func New() *Router {
	return &Router{}
}

func (r *Router) MountPoint() string {
	if r.parent != nil {
		return path.Join(r.parent.MountPoint(), r.mountPoint)
	}
	if r.mountPoint == "" {
		return "/"
	}
	return r.mountPoint
}

func (r *Router) SubRouter(prefix string, rt *Router) {
	if rt.mountPoint != "" {
		panic("cannot mount the same router twice")
	}

	if strings.ContainsRune(prefix, '/') {
		panic("prefix must not contain /")
	}

	if _, has := r.subRouter[prefix]; has {
		panic("route for prefix " + prefix + " already defined")
	}
	if _, has := r.getRoutes[prefix]; has {
		panic("route for prefix " + prefix + " already defined")
	}
	if _, has := r.postRoutes[prefix]; has {
		panic("route for prefix " + prefix + " already defined")
	}
	if _, has := r.putRoutes[prefix]; has {
		panic("route for prefix " + prefix + " already defined")
	}
	if _, has := r.patchRoutes[prefix]; has {
		panic("route for prefix " + prefix + " already defined")
	}
	if _, has := r.deleteRoutes[prefix]; has {
		panic("route for prefix " + prefix + " already defined")
	}

	r.onceSub.Do(func() { r.subRouter = map[string]*Router{} })
	r.subRouter[prefix] = rt
	rt.parent = r
	rt.mountPoint = prefix
}

type Getter interface {
	GET(http.ResponseWriter, *http.Request)
}

type Searcher interface {
	SEARCH(http.ResponseWriter, *http.Request)
}

type Poster interface {
	POST(http.ResponseWriter, *http.Request)
}

type Patcher interface {
	PATCH(http.ResponseWriter, *http.Request)
}

type Putter interface {
	PUT(http.ResponseWriter, *http.Request)
}

type Deleter interface {
	DELETE(http.ResponseWriter, *http.Request)
}

func (r *Router) RouteRessource(object interface{}, prefix string) (url func() string, urlParam func(string) string) {
	if o, ok := object.(Getter); ok {
		r.RouteFuncParam("GET", prefix, o.GET)
	}

	if o, ok := object.(Searcher); ok {
		r.RouteFunc("GET", prefix, o.SEARCH)
	}

	if o, ok := object.(Poster); ok {
		r.RouteFunc("POST", prefix, o.POST)
	}

	if o, ok := object.(Putter); ok {
		r.RouteFuncParam("PUT", prefix, o.PUT)
	}

	if o, ok := object.(Patcher); ok {
		r.RouteFuncParam("PATCH", prefix, o.PATCH)
	}

	if o, ok := object.(Deleter); ok {
		r.RouteFuncParam("DELETE", prefix, o.DELETE)
	}

	fn1 := func() string { return path.Join(r.MountPoint(), prefix) }
	fn2 := func(param string) string { return path.Join(r.MountPoint(), prefix, param) }
	return fn1, fn2
}

func (r *Router) RouteFunc(m string, prefix string, h http.HandlerFunc) func() string {
	return r.Route(m, prefix, h)
}

func (r *Router) Route(m string, prefix string, h http.Handler) func() string {
	if strings.ContainsRune(prefix, '/') {
		panic("prefix must not contain /")
	}

	r.route(m, prefix, h)
	return func() string {
		return path.Join(r.MountPoint(), prefix)
	}
}

func (r *Router) RouteFuncParam(m string, prefix string, h http.HandlerFunc) func(string) string {
	return r.RouteParam(m, prefix, h)
}

func (r *Router) RouteParam(m string, prefix string, h http.Handler) func(string) string {
	if strings.ContainsRune(prefix, '/') {
		panic("prefix must not contain /")
	}
	r.route(m, prefix+"/:", h)
	return func(param string) string {
		return path.Join(r.MountPoint(), prefix, param)
	}
}

func (r *Router) route(m string, prefix string, h http.Handler) {
	if len(r.subRouter) > 0 {
		if _, has := r.subRouter[prefix]; has {
			panic("route for prefix " + prefix + " already defined")
		}
	}

	var mp map[string]http.Handler

	switch m {
	case "GET":
		r.onceGet.Do(func() { r.getRoutes = map[string]http.Handler{} })
		mp = r.getRoutes
	case "POST":
		r.oncePost.Do(func() { r.postRoutes = map[string]http.Handler{} })
		mp = r.postRoutes
	case "PATCH":
		r.oncePatch.Do(func() { r.patchRoutes = map[string]http.Handler{} })
		mp = r.patchRoutes
	case "PUT":
		r.oncePut.Do(func() { r.putRoutes = map[string]http.Handler{} })
		mp = r.putRoutes
	case "DELETE":
		r.onceDelete.Do(func() { r.deleteRoutes = map[string]http.Handler{} })
		mp = r.deleteRoutes
	default:
		panic("unknown method " + m)
	}

	if _, has := mp[prefix]; has {
		panic("route for prefix " + prefix + " already defined")
	}

	mp[prefix] = h
}

func (r *Router) ServeHTTP(w http.ResponseWriter, rq *http.Request) {
	prefix := rq.URL.Path
	// println("rq.URL.Path " + rq.URL.Path)
	// println("MountPoint " + r.MountPoint())
	if r.mountPoint != "" {
		prefix = strings.TrimPrefix(prefix, r.MountPoint())
	}

	prefix = strings.TrimLeft(prefix, "/")
	prefix = strings.TrimRight(prefix, "/")

	r.serveRoute(prefix, w, rq)
}

func (r *Router) findHandler(rq *http.Request, prefix string) http.Handler {
	var m map[string]http.Handler
	switch rq.Method {
	case "GET":
		m = r.getRoutes
	case "POST":
		m = r.postRoutes
	case "PUT":
		m = r.putRoutes
	case "PATCH":
		m = r.patchRoutes
	case "DELETE":
		m = r.deleteRoutes
	}

	if handler, found := m[prefix]; found {
		return handler
	}
	pos := strings.LastIndex(prefix, "/")
	if pos == -1 || pos+1 >= len(prefix) {
		return nil
	}

	param := string(prefix[pos+1:])
	rq.URL.Fragment = param
	prefix = string(prefix[:pos] + "/:")

	return m[prefix]
}

// serve the route for a request
func (r *Router) serveRoute(prefix string, w http.ResponseWriter, rq *http.Request) {
	var pr = prefix
	if idx := strings.Index(prefix, "/"); idx > 0 {
		pr = prefix[:idx]
	}

	// println("serving with prefix " + prefix)
	if sub, found := r.subRouter[pr]; found {
		// println("serving with sub")
		sub.ServeHTTP(w, rq)
		return
	}

	var handler = r.findHandler(rq, prefix)

	if handler == nil {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`method not allowed`))
		return
	}
	handler.ServeHTTP(w, rq)
	return
}
