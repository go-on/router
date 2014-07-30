package route

import "github.com/go-on/method"

type ErrXHRServiceAlreadyRegistered struct{}

func (ErrXHRServiceAlreadyRegistered) Error() string {
	return "XHR handler already registered"
}

type ErrXHRServiceNotRegistered struct{}

func (ErrXHRServiceNotRegistered) Error() string {
	return "XHR handler not registered"
}

type ErrPairParams struct{}

func (ErrPairParams) Error() string {
	return "number of params must be even (pairs of key, value)"
}

type ErrMissingParam struct {
	param       string
	mountedPath string
}

func (e ErrMissingParam) Error() string {
	return "parameter " + e.param + " is missing for route " + e.mountedPath
}

type ErrRouteIsNil struct{}

func (e ErrRouteIsNil) Error() string {
	return "route is nil"
}

type ErrUnknownMethod struct {
	method.Method
}

func (e ErrUnknownMethod) Error() string {
	return "unknown method " + e.Method.String()
}

type ErrMethodNotDefined struct {
	method.Method
	Route *Route
}

func (e *ErrMethodNotDefined) Error() string {
	return "method " + e.Method.String() + " is not defined for route " + e.Route.DefinitionPath
}

type ErrDoubleMounted struct {
	Path  string
	Route *Route
}

func (e *ErrDoubleMounted) Error() string {
	return "route " + e.Route.DefinitionPath + " is already mounted at " + e.Path
}
