package router

import (
	"net/http"
	"strings"

	"github.com/go-on/method"

	"github.com/go-on/router/route"
)

func SetOPTIONSHandler(r *routeHandler) {
	optionsString := strings.Join(route.Options(r.Route), ",")
	r.Route.Methods[method.OPTIONS] = struct{}{}
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
	rt := r.newRouteHandler(path, method.GET)
	rt.GETHandler = handler
	return rt.Route
}

func (r *Router) POST(path string, handler http.Handler) *route.Route {
	mustNotBeRouter(handler)
	rt := r.newRouteHandler(path, method.POST)
	rt.POSTHandler = handler
	return rt.Route
}

func (r *Router) PUT(path string, handler http.Handler) *route.Route {
	mustNotBeRouter(handler)
	rt := r.newRouteHandler(path, method.PUT)
	rt.PUTHandler = handler
	return rt.Route
}

func (r *Router) PATCH(path string, handler http.Handler) *route.Route {
	mustNotBeRouter(handler)
	rt := r.newRouteHandler(path, method.PATCH)
	rt.PATCHHandler = handler
	return rt.Route
}

func (r *Router) DELETE(path string, handler http.Handler) *route.Route {
	mustNotBeRouter(handler)
	rt := r.newRouteHandler(path, method.DELETE)
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

func (r *Router) HandleRouteFunc(rt *route.Route, handler http.HandlerFunc) {
	r.HandleRoute(rt, handler)
}

func (r *Router) HandleRouteGET(rt *route.Route, handler http.Handler) {
	r.HandleRouteMethods(rt, handler, method.GET)
}

func (r *Router) HandleRoutePOST(rt *route.Route, handler http.Handler) {
	r.HandleRouteMethods(rt, handler, method.POST)
}

func (r *Router) HandleRoutePUT(rt *route.Route, handler http.Handler) {
	r.HandleRouteMethods(rt, handler, method.PUT)
}

func (r *Router) HandleRoutePATCH(rt *route.Route, handler http.Handler) {
	r.HandleRouteMethods(rt, handler, method.PATCH)
}

func (r *Router) HandleRouteDELETE(rt *route.Route, handler http.Handler) {
	r.HandleRouteMethods(rt, handler, method.DELETE)
}

func (r *Router) HandleRouteOPTIONS(rt *route.Route, handler http.Handler) {
	r.HandleRouteMethods(rt, handler, method.OPTIONS)
}

func (r *Router) HandleRouteGETFunc(rt *route.Route, handler http.HandlerFunc) {
	r.HandleRouteMethods(rt, handler, method.GET)
}

func (r *Router) HandleRoutePOSTFunc(rt *route.Route, handler http.HandlerFunc) {
	r.HandleRouteMethods(rt, handler, method.POST)
}

func (r *Router) HandleRoutePUTFunc(rt *route.Route, handler http.HandlerFunc) {
	r.HandleRouteMethods(rt, handler, method.PUT)
}

func (r *Router) HandleRoutePATCHFunc(rt *route.Route, handler http.HandlerFunc) {
	r.HandleRouteMethods(rt, handler, method.PATCH)
}

func (r *Router) HandleRouteDELETEFunc(rt *route.Route, handler http.HandlerFunc) {
	r.HandleRouteMethods(rt, handler, method.DELETE)
}

func (r *Router) HandleRouteOPTIONSFunc(rt *route.Route, handler http.HandlerFunc) {
	r.HandleRouteMethods(rt, handler, method.OPTIONS)
}
