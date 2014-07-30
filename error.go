package router

import (
	"fmt"

	"github.com/go-on/router/route"

	"github.com/go-on/method"
)

type ErrDoubleRegistration struct {
	DefinitionPath string
}

func (e ErrDoubleRegistration) Error() string {
	return fmt.Sprintf("path %#v already registered by another route", e.DefinitionPath)
}

type ErrNotMounted struct{}

func (e ErrNotMounted) Error() string {
	return "router is not mounted"
}

type ErrInvalidMountPath struct {
	Path   string
	Reason string
}

func (e ErrInvalidMountPath) Error() string {
	return fmt.Sprintf("mount path %#v is invalid: %s", e.Path, e.Reason)
}

type ErrDoubleMounted struct {
	Path string
}

type ErrRouterNotAllowed struct{}

func (e ErrRouterNotAllowed) Error() string {
	return "handler must not be a *Router, use Handle() or Mount() instead"
}

func (e ErrDoubleMounted) Error() string {
	return fmt.Sprintf("router is already mounted at %#v", e.Path)
}

type ErrHandlerAlreadyDefined struct {
	method.Method
}

type ErrUnknownMethod struct {
	method.Method
}

func (e ErrUnknownMethod) Error() string {
	return "unknown method " + e.Method.String()
}

func (e ErrHandlerAlreadyDefined) Error() string {
	return "handler for " + e.Method.String() + " already defined"
}

type ErrMethodNotDefinedForRoute struct {
	method.Method
	Route *route.Route
}

func (e *ErrMethodNotDefinedForRoute) Error() string {
	return "method " + e.Method.String() + " is not defined for route " + e.Route.DefinitionPath
}

type ErrMissingHandler struct {
	methods []method.Method
	Route   *route.Route
}

func (e *ErrMissingHandler) Error() string {
	return fmt.Sprintf("route %s has no handler defined for the methods %v", e.Route.DefinitionPath, e.methods)
}
