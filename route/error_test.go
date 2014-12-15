package route

import (
	"fmt"
	"reflect"
	"testing"

	"gopkg.in/go-on/method.v1"
	"github.com/gopherjs/gopherjs/js"
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

func TestDoubleRegisteredXHRService(t *testing.T) {
	xhr = nil
	aj := &XHRFuncs{}
	RegisterXHRService(aj)

	defer func() {
		e := recover()
		errMsg := errorMustBe(e, ErrXHRServiceAlreadyRegistered{})

		if errMsg != "" {
			t.Error(errMsg)
			return
		}

		_ = e.(ErrXHRServiceAlreadyRegistered).Error()
	}()

	RegisterXHRService(aj)
}

func TestNotRegisteredXHRService(t *testing.T) {
	xhr = nil
	rt := New("/", method.GET)
	Mount("/", rt)
	defer func() {
		e := recover()
		errMsg := errorMustBe(e, ErrXHRServiceNotRegistered{})

		if errMsg != "" {
			t.Error(errMsg)
			return
		}

		_ = e.(ErrXHRServiceNotRegistered).Error()
	}()

	rt.Get(func(js.Object) {})
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

		if err.param != "name" {
			t.Errorf("wrong param: %#v, expected: %v", err.param, "name")
		}

		if err.mountedPath != "/a/route/:name" {
			t.Errorf("wrong mountedPath: %#v, expected: %v", err.mountedPath, "/a/route/:name")
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

func testMethodNotDefined(has method.Method, hasNot method.Method, t *testing.T) {
	xhr = nil
	route := New("/route/:name", has)

	Mount("/a", route)

	x := &XHRFuncs{}
	RegisterXHRService(x)

	defer func() {
		e := recover()
		errMsg := errorMustBe(e, &ErrMethodNotDefined{})

		if errMsg != "" {
			t.Error(errMsg)
			return
		}

		err := e.(*ErrMethodNotDefined)
		_ = err.Error()

		if err.Method != hasNot {
			t.Errorf("wrong method: %#v, expected: %v", err.Method.String(), hasNot.String())
		}

		if err.Route != route {
			t.Errorf("wrong route: %#v, expected: %v", err.Route.DefinitionPath, route.DefinitionPath)
		}
	}()

	switch hasNot {
	case method.GET:
		route.Get(nil)
	case method.POST:
		route.Post(nil, nil)
	case method.PUT:
		route.Put(nil, nil)
	case method.PATCH:
		route.Patch(nil, nil)
	case method.DELETE:
		route.Delete(nil)
	case method.OPTIONS:
		route.Options(nil)
	}
}

func TestMethodNotDefined(t *testing.T) {
	testMethodNotDefined(method.POST, method.GET, t)
	testMethodNotDefined(method.POST, method.PUT, t)
	testMethodNotDefined(method.POST, method.PATCH, t)
	testMethodNotDefined(method.POST, method.DELETE, t)
	testMethodNotDefined(method.POST, method.OPTIONS, t)
	testMethodNotDefined(method.GET, method.POST, t)
	testMethodNotDefined(method.OPTIONS, method.GET, t)
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
