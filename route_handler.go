package router

import (
	"github.com/go-on/router/route"

	"net/http"

	"github.com/go-on/method"
)

type routeHandler struct {
	GETHandler     http.Handler
	POSTHandler    http.Handler
	PUTHandler     http.Handler
	PATCHHandler   http.Handler
	DELETEHandler  http.Handler
	OPTIONSHandler http.Handler
	*route.Route
}

func newRouteHandler(rt *route.Route) *routeHandler {
	rh := &routeHandler{Route: rt}
	return rh
}

/*
func (r *routeHandler) Clone() *routeHandler {
	rh := &routeHandler{
		GETHandler:     r.GETHandler,
		POSTHandler:    r.POSTHandler,
		PUTHandler:     r.PUTHandler,
		PATCHHandler:   r.PATCHHandler,
		DELETEHandler:  r.DELETEHandler,
		OPTIONSHandler: r.OPTIONSHandler,
		Route:          r.Route,
	}
	// rt.Id = fmt.Sprintf("//%p", rt)
	return rh
}
*/

func (r *routeHandler) MissingHandler() (missing []method.Method) {
	// fmt.Printf("running MissingHandler for %s\n", r.Route.DefinitionPath)
	for m := range r.Route.Methods {
		// fmt.Printf("checking %s\n", m)
		if r.Handler(m) == nil {
			missing = append(missing, m)
		}
	}
	// fmt.Printf("MissingHandler for %s: %v\n", r.Route.DefinitionPath, missing)
	return
}

func (r *routeHandler) Handler(meth method.Method) http.Handler {
	switch meth {
	case method.GET:
		return r.GETHandler
	case method.POST:
		return r.POSTHandler
	case method.PUT:
		return r.PUTHandler
	case method.PATCH:
		return r.PATCHHandler
	case method.DELETE:
		return r.DELETEHandler
	case method.OPTIONS:
		return r.OPTIONSHandler
	}
	return nil
}

func (rt *routeHandler) SetHandlerForMethod(handler http.Handler, m method.Method) {
	switch m {
	case method.GET:
		if rt.GETHandler != nil {
			panic(ErrHandlerAlreadyDefined{m})
		}
		rt.GETHandler = handler
	case method.PUT:
		if rt.PUTHandler != nil {
			panic(ErrHandlerAlreadyDefined{m})
		}
		rt.PUTHandler = handler
	case method.POST:
		if rt.POSTHandler != nil {
			panic(ErrHandlerAlreadyDefined{m})
		}
		rt.POSTHandler = handler
	case method.DELETE:
		if rt.DELETEHandler != nil {
			panic(ErrHandlerAlreadyDefined{m})
		}
		rt.DELETEHandler = handler
	case method.PATCH:
		if rt.PATCHHandler != nil {
			panic(ErrHandlerAlreadyDefined{m})
		}
		rt.PATCHHandler = handler
	case method.OPTIONS:
		if rt.OPTIONSHandler != nil {
			panic(ErrHandlerAlreadyDefined{m})
		}
		rt.OPTIONSHandler = handler
	default:
		panic(ErrUnknownMethod{m})
	}
}

func (rt *routeHandler) SetHandlerForMethods(handler http.Handler, method1 method.Method, furtherMethods ...method.Method) {
	methods := append(furtherMethods, method1)
	for _, m := range methods {
		rt.SetHandlerForMethod(handler, m)
	}
}

func (r *routeHandler) EachHandler(fn func(http.Handler) error) error {

	if r.GETHandler != nil {
		err := fn(r.GETHandler)
		if err != nil {
			return err
		}
	}

	if r.POSTHandler != nil {
		err := fn(r.POSTHandler)
		if err != nil {
			return err
		}
	}

	if r.PUTHandler != nil {
		err := fn(r.PUTHandler)
		if err != nil {
			return err
		}
	}

	if r.PATCHHandler != nil {
		err := fn(r.PATCHHandler)
		if err != nil {
			return err
		}
	}

	if r.DELETEHandler != nil {
		err := fn(r.DELETEHandler)
		if err != nil {
			return err
		}
	}

	if r.OPTIONSHandler != nil {
		err := fn(r.OPTIONSHandler)
		if err != nil {
			return err
		}
	}
	return nil
}
