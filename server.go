package router

// this handler should be used for the top level router

import (
	"net/http"

	"github.com/go-on/wrap"

	"github.com/go-on/wrap-contrib/wraps"
)

// Serve serves the request on the top level. It handles method override and path cleaning
// and then serves via the corresponding http.Handler of the route or passes it to a given wrapper
//
// Serve will selfmount the router under / if it is not already mounted
func (ø *Router) Serve(wrapper wrap.Wrapper) http.Handler {
	if ø.mountPoint == "" {
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
	if wrapper != nil {
		stack = append(stack, wrapper)
	}
	stack = append(stack, wrap.Handler(ø))
	return wrap.New(stack...)
}

// if server is nil, the default server is used
func (ø *Router) ListenAndServe(addr string, server *http.Server) error {
	if server == nil {
		return http.ListenAndServe(addr, ø.Serve(nil))
	}
	server.Addr = addr
	server.Handler = ø.Serve(nil)
	return server.ListenAndServe()
}

func (ø *Router) ListenAndServeTLS(addr string, certFile string, keyFile string, server *http.Server) error {
	if server == nil {
		return http.ListenAndServeTLS(addr, certFile, keyFile, ø.Serve(nil))
	}
	server.Addr = addr
	server.Handler = ø.Serve(nil)
	return server.ListenAndServeTLS(certFile, keyFile)
}
