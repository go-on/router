package router

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/go-on/router/route"

	"gopkg.in/go-on/method.v1"
)

var allMethods = []method.Method{method.GET, method.POST, method.PATCH, method.PUT, method.DELETE}

func TestRouteHandler(t *testing.T) {
	h := writeString("hu")
	rt := route.New("/hu", method.GET, method.POST, method.PATCH, method.PUT, method.DELETE)

	rh := newRouteHandler(rt)
	rh.SetHandlerForMethods(h, method.GET, method.POST, method.PATCH, method.PUT, method.DELETE)

	for _, meth := range allMethods {
		if rh.Handler(meth) != h {
			t.Errorf("wrong %s handler", meth)
		}
	}

	if rh.Handler(method.Method("unknown")) != nil {
		t.Errorf("unknown method should not return handler")
	}

	num := 0

	err := rh.EachHandler(func(hd http.Handler) error {
		if hd != h {
			return fmt.Errorf("wrong handler")
		}
		num++
		return nil
	})

	if err != nil {
		t.Error(err.Error())
	}

	if num != 5 {
		t.Errorf("wrong number of handlers in EachHandler: %d, expected: %d", num, 6)
	}

	for _, meth := range allMethods {
		func() {

			defer func() {
				e := recover()
				if e == nil {
					if rh.Handler(meth) != h {
						t.Errorf("missing error for double definition of%s", meth)
					}
				}
			}()

			rh.SetHandlerForMethod(h, meth)
		}()

	}

}
