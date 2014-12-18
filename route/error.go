package route

import "gopkg.in/go-on/method.v1"

// ErrPairParams is raised if a variadic parameter group has no pairs.
type ErrPairParams struct{}

func (ErrPairParams) Error() string {
	return "number of params must be even (pairs of key, value)"
}

// ErrMissingParam is raised if a route URL parameter is missing.
type ErrMissingParam struct {
	Param       string
	MountedPath string
}

func (e ErrMissingParam) Error() string {
	return "parameter " + e.Param + " is missing for route " + e.MountedPath
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

// ErrDoubleMounted is raised if the route already has been mounted.
type ErrDoubleMounted struct {
	Path  string
	Route *Route
}

func (e *ErrDoubleMounted) Error() string {
	return "route " + e.Route.DefinitionPath + " is already mounted at " + e.Path
}
