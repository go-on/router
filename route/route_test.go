package route

import (
	"net/http"

	"gopkg.in/go-on/method.v1"
	"github.com/gopherjs/gopherjs/js"

	"testing"
)

func TestURL(t *testing.T) {
	route1 := New("/route1", method.GET)

	Mount("/", route1)

	got := route1.MustURL()

	if got != "/route1" {
		t.Errorf("wrong URL: %#v, wanted: %#v", got, "/route1")
	}

	route3 := New("/route3", method.GET)

	Mount("/api/v1", route3)

	got = route3.MustURL()

	if got != "/api/v1/route3" {
		t.Errorf("wrong URL: %#v, wanted: %#v", got, "/api/v1/route3")
	}

	route2 := New("/route2/:param2", method.GET)
	Mount("/:param1", route2)

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
	route := New("/route/:param1/:param2", method.GET)
	Mount("/", route)
	defer func() {
		e := recover()
		if e == nil {
			t.Errorf("should report missing param")
		}
	}()

	route.MustURL("param1", "val1")
}

func TestURLMissingParamMap(t *testing.T) {
	route := New("/route/:param1/:param2", method.GET)
	Mount("/", route)
	// route.Router = PseudoRouter("/")
	defer func() {
		e := recover()
		if e == nil {
			t.Errorf("should report missing param")
		}
	}()

	route.MustURLMap(map[string]string{"param1": "val1"})
}

func TestURLMissingValue(t *testing.T) {
	route := New("/route/:param", method.GET)
	// route.Router = PseudoRouter("/")
	Mount("/", route)
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

func TestHXR(t *testing.T) {
	route := New("/", method.GET, method.POST, method.PATCH, method.PUT, method.DELETE, method.OPTIONS)

	Mount("/", route)

	aj := &XHRFuncs{}

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
	xhr = nil
	RegisterXHRService(aj)
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
}

func TestHasParams(t *testing.T) {
	route1 := New("/route/:param", method.GET)
	if !route1.HasParams() {
		t.Errorf("route1 has params")
	}

	route2 := New("/route", method.GET)
	if route2.HasParams() {
		t.Errorf("route2 has no params")
	}

}

func TestHasMethod(t *testing.T) {
	route := New("/route/:param", method.GET, method.POST)
	if !route.HasMethod(method.GET) {
		t.Errorf("route should have method %s", method.GET)
	}
	if !route.HasMethod(method.HEAD) {
		t.Errorf("route should have method %s", method.HEAD)
	}
	if !route.HasMethod(method.POST) {
		t.Errorf("route should have method %s", method.POST)
	}
	if route.HasMethod(method.PUT) {
		t.Errorf("route should not have method %s", method.PUT)
	}

}

func TestOptions(t *testing.T) {
	route := New("/", method.GET, method.POST, method.PATCH, method.PUT, method.DELETE)

	Mount("/", route)

	opts := Options(route)

	shouldHave := func(m method.Method) {
		for _, o := range opts {
			if o == m.String() {
				return
			}
		}
		t.Errorf("missing option: %s", m.String())
	}

	shouldHave(method.GET)
	shouldHave(method.POST)
	shouldHave(method.PATCH)
	shouldHave(method.PUT)
	shouldHave(method.DELETE)
	shouldHave(method.HEAD)
	shouldHave(method.OPTIONS)
}
