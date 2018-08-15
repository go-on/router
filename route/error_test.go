package route

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/go-on/method"
)

func errorMustBe(err interface{}, class interface{}) string {
	classTy := reflect.TypeOf(class)
	if err == nil {
		return fmt.Sprintf("error must be of type %s but is nil", classTy)
	}

	errTy := reflect.TypeOf(err)
	if errTy.String() != classTy.String() {
		return fmt.Sprintf("error must be of type %s but is of type %s", classTy, errTy)
	}
	return ""
}

func TestUnknownMethod(t *testing.T) {

	defer func() {
		e := recover()
		errMsg := errorMustBe(e, ErrUnknownMethod{})

		if errMsg != "" {
			t.Error(errMsg)
			return
		}

		err := e.(ErrUnknownMethod)
		_ = err.Error()

		if err.Method.String() != "unknown" {
			t.Errorf("wrong method: %#v, expected: %v", err.Method, "unknown")
		}
	}()

	New("/route", method.Method("unknown"))
}

func TestErrPairParams(t *testing.T) {
	route := New("/route", method.GET)

	defer func() {
		e := recover()
		errMsg := errorMustBe(e, ErrPairParams{})

		if errMsg != "" {
			t.Error(errMsg)
			return
		}

		err := e.(ErrPairParams)
		_ = err.Error()
	}()

	route.MustURL("param1")
}

func TestErrMissingParams(t *testing.T) {
	route := New("/route/:name", method.GET)

	Mount("/a", route)

	defer func() {
		e := recover()
		errMsg := errorMustBe(e, ErrMissingParam{})

		if errMsg != "" {
			t.Error(errMsg)
			return
		}

		err := e.(ErrMissingParam)
		_ = err.Error()

		if err.Param != "name" {
			t.Errorf("wrong param: %#v, expected: %v", err.Param, "name")
		}

		if err.MountedPath != "/a/route/:name" {
			t.Errorf("wrong mountedPath: %#v, expected: %v", err.MountedPath, "/a/route/:name")
		}
	}()

	route.MustURL()
}

func TestDoubleMounted(t *testing.T) {
	route := New("/route/:name", method.GET)

	Mount("/a", route)

	defer func() {
		e := recover()
		errMsg := errorMustBe(e, &ErrDoubleMounted{})

		if errMsg != "" {
			t.Error(errMsg)
			return
		}

		err := e.(*ErrDoubleMounted)
		_ = err.Error()

		if err.Path != "/a" {
			t.Errorf("wrong Path: %#v, expected: %v", err.Path, "/a")
		}

		if err.Route != route {
			t.Errorf("wrong route: %#v, expected: %v", err.Route.DefinitionPath, route.DefinitionPath)
		}
	}()

	Mount("/b", route)
}

func TestRouteIsNil(t *testing.T) {
	var route *Route

	defer func() {
		e := recover()
		errMsg := errorMustBe(e, ErrRouteIsNil{})

		if errMsg != "" {
			t.Error(errMsg)
			return
		}

		err := e.(ErrRouteIsNil)
		_ = err.Error()

	}()

	route.MustURL()
}

/*
func TestHandlerAlreadyDefined(t *testing.T) {
	route := New("/route")
	route.SetHandlerForMethod(noop{}, method.GET)

	defer func() {
		e := recover()
		errMsg := errorMustBe(e, ErrHandlerAlreadyDefined{})

		if errMsg != "" {
			t.Error(errMsg)
			return
		}

		err := e.(ErrHandlerAlreadyDefined)
		_ = err.Error()

		if err.Method != method.GET {
			t.Errorf("wrong method: %#v, expected: %v", err.Method, method.GET)
		}
	}()

	route.SetHandlerForMethod(noop{}, method.GET)
}



// ErrPairParams

*/
