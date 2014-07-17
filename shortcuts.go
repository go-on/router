package router

import (
	"net/http"

	"github.com/go-on/router/route"
)

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

func (r *Router) GET(path string, handler http.Handler) *route.Route {
	rt := r.newRoute(path)
	rt.GETHandler = handler
	return rt
}

func (r *Router) POST(path string, handler http.Handler) *route.Route {
	rt := r.newRoute(path)
	rt.POSTHandler = handler
	return rt
}

func (r *Router) PUT(path string, handler http.Handler) *route.Route {
	rt := r.newRoute(path)
	rt.PUTHandler = handler
	return rt
}

func (r *Router) PATCH(path string, handler http.Handler) *route.Route {
	rt := r.newRoute(path)
	rt.PATCHHandler = handler
	return rt
}

func (r *Router) DELETE(path string, handler http.Handler) *route.Route {
	rt := r.newRoute(path)
	rt.DELETEHandler = handler
	return rt
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
