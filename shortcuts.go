package router

import (
	"net/http"
	"strings"

	"github.com/go-on/method"

	"github.com/go-on/router/route"
)

func SetOPTIONSHandler(r *routeHandler) {
	optionsString := strings.Join(route.Options(r.Route), ",")
	r.OPTIONSHandler = http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Allow", optionsString)
	})
}

/*
// if server is nil, the default server is used
func (ø *Router) ListenAndServe(addr string, server *http.Server) error {
	if server == nil {
		return http.ListenAndServe(addr, ø.ServingHandler())
	}
	server.Addr = addr
	server.Handler = ø.ServingHandler()
	return server.ListenAndServe()
}

func (ø *Router) ListenAndServeTLS(addr string, certFile string, keyFile string, server *http.Server) error {
	if server == nil {
		return http.ListenAndServeTLS(addr, certFile, keyFile, ø.ServingHandler())
	}
	server.Addr = addr
	server.Handler = ø.ServingHandler()
	return server.ListenAndServeTLS(certFile, keyFile)
}
*/

func (r *Router) GET(path string, handler http.Handler) *route.Route {
	mustNotBeRouter(handler)
	rt := r.newRoute(path, method.GET)
	rt.GETHandler = handler
	return rt.Route
}

func (r *Router) POST(path string, handler http.Handler) *route.Route {
	mustNotBeRouter(handler)
	rt := r.newRoute(path, method.POST)
	rt.POSTHandler = handler
	return rt.Route
}

func (r *Router) PUT(path string, handler http.Handler) *route.Route {
	mustNotBeRouter(handler)
	rt := r.newRoute(path, method.PUT)
	rt.PUTHandler = handler
	return rt.Route
}

func (r *Router) PATCH(path string, handler http.Handler) *route.Route {
	mustNotBeRouter(handler)
	rt := r.newRoute(path, method.PATCH)
	rt.PATCHHandler = handler
	return rt.Route
}

func (r *Router) DELETE(path string, handler http.Handler) *route.Route {
	mustNotBeRouter(handler)
	rt := r.newRoute(path, method.DELETE)
	rt.DELETEHandler = handler
	return rt.Route
}

func (r *Router) GETFunc(path string, handler http.HandlerFunc) *route.Route {
	return r.GET(path, handler)
}

func (r *Router) POSTFunc(path string, handler http.HandlerFunc) *route.Route {
	return r.POST(path, handler)
}

func (r *Router) PUTFunc(path string, handler http.HandlerFunc) *route.Route {
	return r.PUT(path, handler)
}

func (r *Router) PATCHFunc(path string, handler http.HandlerFunc) *route.Route {
	return r.PATCH(path, handler)
}

func (r *Router) DELETEFunc(path string, handler http.HandlerFunc) *route.Route {
	return r.DELETE(path, handler)
}
