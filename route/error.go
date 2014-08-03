package route

import "github.com/go-on/method"

// ErrXHRServiceAlreadyRegistered is raised if the XHRService has already been registered.
type ErrXHRServiceAlreadyRegistered struct{}

func (ErrXHRServiceAlreadyRegistered) Error() string {
	return "XHR handler already registered"
}

// ErrXHRServiceNotRegistered is raised if no XHRService has been registered.
type ErrXHRServiceNotRegistered struct{}

func (ErrXHRServiceNotRegistered) Error() string {
	return "XHR handler not registered"
}

// ErrPairParams is raised if a variadic parameter group has no pairs.
type ErrPairParams struct{}

func (ErrPairParams) Error() string {
	return "number of params must be even (pairs of key, value)"
}

// ErrMissingParam is raised if a route URL parameter is missing.
type ErrMissingParam struct {
	param       string
	mountedPath string
}

func (e ErrMissingParam) Error() string {
	return "parameter " + e.param + " is missing for route " + e.mountedPath
}

// ErrRouteIsNil is raised if a route is not yet defined.
type ErrRouteIsNil struct{}

func (e ErrRouteIsNil) Error() string {
	return "route is nil"
}

// ErrUnknownMethod is raised if the given http method is not known.
type ErrUnknownMethod struct {
	method.Method
}

func (e ErrUnknownMethod) Error() string {
	return "unknown method " + e.Method.String()
}

// ErrMethodNotDefined is raised if the given http method is not defined for the given route
type ErrMethodNotDefined struct {
	method.Method
	Route *Route
}

func (e *ErrMethodNotDefined) Error() string {
	return "method " + e.Method.String() + " is not defined for route " + e.Route.DefinitionPath
}

// ErrDoubleMounted is raised if the route already has been mounted.
type ErrDoubleMounted struct {
	Path  string
	Route *Route
}

func (e *ErrDoubleMounted) Error() string {
	return "route " + e.Route.DefinitionPath + " is already mounted at " + e.Path
}
