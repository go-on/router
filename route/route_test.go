package route

import (
	"fmt"
	"net/http"

	"github.com/go-on/method"
	"github.com/gopherjs/gopherjs/js"

	"testing"
)

func TestURL(t *testing.T) {
	route1 := New("/route1")

	route1.Router = PseudoRouter("/")

	got := route1.MustURL()

	if got != "/route1" {
		t.Errorf("wrong URL: %#v, wanted: %#v", got, "/route1")
	}

	route1.Router = PseudoRouter("/api/v1")

	got = route1.MustURL()

	if got != "/api/v1/route1" {
		t.Errorf("wrong URL: %#v, wanted: %#v", got, "/api/v1/route1")
	}

	route2 := New("/route2/:param2")
	route2.Router = PseudoRouter("/:param1")

	got = route2.MustURL("param1", "val1", "param2", "val2")

	if got != "/val1/route2/val2" {
		t.Errorf("wrong URL: %#v, wanted: %#v", got, "/val1/route2/val2")
	}

	got = route2.MustURLMap(map[string]string{"param1": "v1", "param2": "v2"})

	if got != "/v1/route2/v2" {
		t.Errorf("wrong URL: %#v, wanted: %#v", got, "/v1/route2/v2")
	}

}

func TestURLMissingParam(t *testing.T) {
	route := New("/route/:param1/:param2")
	route.Router = PseudoRouter("/")
	defer func() {
		e := recover()
		if e == nil {
			t.Errorf("should report missing param")
		}
	}()

	route.MustURL("param1", "val1")
}

func TestURLMissingParamMap(t *testing.T) {
	route := New("/route/:param1/:param2")
	route.Router = PseudoRouter("/")
	defer func() {
		e := recover()
		if e == nil {
			t.Errorf("should report missing param")
		}
	}()

	route.MustURLMap(map[string]string{"param1": "val1"})
}

func TestURLMissingValue(t *testing.T) {
	route := New("/route/:param")
	route.Router = PseudoRouter("/")
	defer func() {
		e := recover()
		if e == nil {
			t.Errorf("should report missing value")
		}
	}()

	route.MustURL("param")
}

type noop struct{}

var allMethods = []method.Method{
	method.GET,
	method.POST,
	method.PUT,
	method.PATCH,
	method.DELETE,
	method.OPTIONS,
}

func (noop) ServeHTTP(rw http.ResponseWriter, req *http.Request) {}

func TestRoute(t *testing.T) {
	route := New("/route")
	route.Router = PseudoRouter("/fake")

	h := noop{}
	route.SetHandlerForMethods(h, allMethods...)

	for _, meth := range allMethods {
		if route.Handler(meth) != h {
			t.Errorf("wrong %s handler", meth)
		}
	}

	if route.Handler(method.Method("unknown")) != nil {
		t.Errorf("unknown method should not return handler")
	}

	route2 := route.Clone()

	for _, meth := range allMethods {
		if route2.Handler(meth) != h {
			t.Errorf("wrong clone %s handler", meth)
		}
	}

	if route2.Router != route.Router {
		t.Errorf("wrong clone Router")
	}

	if route2.DefinitionPath != route.DefinitionPath {
		t.Errorf("wrong clone DefinitionPath")
	}

	opts := Options(route)

	has := func(meth method.Method) bool {
		for _, m := range opts {
			if m == meth.String() {
				return true
			}
		}
		return false
	}

	for _, meth := range allMethods {
		if !has(meth) {
			t.Errorf("options missing %s handler", meth)
		}
	}

	if !has(method.HEAD) {
		t.Errorf("options missing %s handler", method.HEAD)
	}

	num := 0

	err := route2.EachHandler(func(hd http.Handler) error {
		if hd != h {
			return fmt.Errorf("wrong handler")
		}
		num++
		return nil
	})

	if err != nil {
		t.Error(err.Error())
	}

	if num != 6 {
		t.Errorf("wrong number of handlers in EachHandler: %d, expected: %d", num, 6)
	}

	for _, meth := range allMethods {
		func() {

			defer func() {
				e := recover()
				if e == nil {
					if route.Handler(meth) != h {
						t.Errorf("missing error for double definition of%s", meth)
					}
				}
			}()

			route.SetHandlerForMethod(h, meth)
		}()

	}

	aj := &PseudoAjaxHandler{}

	methCalled := []method.Method{}

	aj.GET = func(url string, callback func(js.Object)) {
		methCalled = append(methCalled, method.GET)
	}

	aj.POST = func(url string, data interface{}, callback func(js.Object)) {
		methCalled = append(methCalled, method.POST)
	}

	aj.PUT = func(url string, data interface{}, callback func(js.Object)) {
		methCalled = append(methCalled, method.PUT)
	}

	aj.PATCH = func(url string, data interface{}, callback func(js.Object)) {
		methCalled = append(methCalled, method.PATCH)
	}

	aj.DELETE = func(url string, callback func(js.Object)) {
		methCalled = append(methCalled, method.DELETE)
	}

	aj.OPTIONS = func(url string, callback func(js.Object)) {
		methCalled = append(methCalled, method.OPTIONS)
	}
	ajax = nil
	RegisterAjaxHandler(aj)
	expectedMethCalled := 0

	route.Get(nil)
	expectedMethCalled++
	if len(methCalled) != expectedMethCalled {
		t.Errorf("ajax %s not called", method.GET)
	}

	route.Post(nil, nil)
	expectedMethCalled++
	if len(methCalled) != expectedMethCalled {
		t.Errorf("ajax %s not called", method.POST)
	}

	route.Put(nil, nil)
	expectedMethCalled++
	if len(methCalled) != expectedMethCalled {
		t.Errorf("ajax %s not called", method.PUT)
	}

	route.Patch(nil, nil)
	expectedMethCalled++
	if len(methCalled) != expectedMethCalled {
		t.Errorf("ajax %s not called", method.PATCH)
	}

	route.Delete(nil)
	expectedMethCalled++
	if len(methCalled) != expectedMethCalled {
		t.Errorf("ajax %s not called", method.DELETE)
	}

	route.Options(nil)
	expectedMethCalled++
	if len(methCalled) != expectedMethCalled {
		t.Errorf("ajax %s not called", method.OPTIONS)
	}

	/*
		defer func() {
			e := recover()
			if e == nil {
				t.Errorf("should not second registration of ajax handler")
			}
		}()
		RegisterAjaxHandler(aj)
	*/
}

/*
func TestUnknownMethod(t *testing.T) {
	route := NewRoute("/route")
	if route.Handler(method.Method("unknown")) != nil {
		t.Errorf("should find not handler for unknown method")
	}

	defer func() {
		e := recover()
		if e == nil {
			t.Errorf("should not allow registration of unknown handler")
		}
	}()

	route.SetHandlerForMethod(noop{}, method.Method("unknown"))
}
*/

func TestHasParams(t *testing.T) {
	route1 := New("/route/:param")
	if !route1.HasParams() {
		t.Errorf("route1 has params")
	}

	route2 := New("/route")
	if route2.HasParams() {
		t.Errorf("route2 has no params")
	}

}
