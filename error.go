package router

import "fmt"

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

func (e ErrDoubleMounted) Error() string {
	return fmt.Sprintf("router is already mounted at %#v", e.Path)
}
