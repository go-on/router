package route

import "github.com/go-on/method"

// TODO: add the mountedPath to every error message

type ErrAjaxAlreadyRegistered struct{}

func (ErrAjaxAlreadyRegistered) Error() string {
	return "ajax handler already registered"
}

type ErrHandlerAlreadyDefined struct {
	method.Method
}

func (e ErrHandlerAlreadyDefined) Error() string {
	return "handler for " + e.Method.String() + " already defined"
}

type ErrUnknownMethod struct {
	method.Method
}

func (e ErrUnknownMethod) Error() string {
	return "unknown method " + e.Method.String()
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
