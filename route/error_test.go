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

func TestDoubleRegisteredAjax(t *testing.T) {
	ajax = nil
	aj := &PseudoAjaxHandler{}
	RegisterAjaxHandler(aj)

	defer func() {
		e := recover()
		errMsg := errorMustBe(e, ErrAjaxAlreadyRegistered{})

		if errMsg != "" {
			t.Error(errMsg)
			return
		}

		_ = e.(ErrAjaxAlreadyRegistered).Error()
	}()

	RegisterAjaxHandler(ajax)
}

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

func TestUnknownMethod(t *testing.T) {
	route := New("/route")

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

	route.SetHandlerForMethod(noop{}, method.Method("unknown"))
}

// ErrPairParams

func TestErrPairParams(t *testing.T) {
	route := New("/route")

	defer func() {
		e := recover()
		errMsg := errorMustBe(e, ErrPairParams{})

		if errMsg != "" {
			t.Error(errMsg)
			return
		}

		err := e.(ErrPairParams)
		_ = err.Error()
		/*
			if err.Method.String() != "unknown" {
				t.Errorf("wrong method: %#v, expected: %v", err.Method, "unknown")
			}
		*/
	}()

	route.MustURL("param1")
	// route.SetHandlerForMethod(noop{}, method.Method("unknown"))
}
