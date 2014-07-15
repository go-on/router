package router

import (
	"net/http"

	"github.com/go-on/router/route"
)

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
	rt := r.newRoute(path)
	rt.GETHandler = handler
	return rt
}

func (r *Router) POSTFunc(path string, handler http.HandlerFunc) *route.Route {
	rt := r.newRoute(path)
	rt.POSTHandler = handler
	return rt
}

func (r *Router) PUTFunc(path string, handler http.HandlerFunc) *route.Route {
	rt := r.newRoute(path)
	rt.PUTHandler = handler
	return rt
}

func (r *Router) PATCHFunc(path string, handler http.HandlerFunc) *route.Route {
	rt := r.newRoute(path)
	rt.PATCHHandler = handler
	return rt
}

func (r *Router) DELETEFunc(path string, handler http.HandlerFunc) *route.Route {
	rt := r.newRoute(path)
	rt.DELETEHandler = handler
	return rt
}
