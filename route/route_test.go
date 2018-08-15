package route

import (
	"net/http"

	"github.com/go-on/method"

	"testing"
)

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

func TestURLMissingParamMap(t *testing.T) {
	rt := New("/route/:param1/:param2", method.GET)
	Mount("/", rt)
	// route.Router = PseudoRouter("/")
	defer func() {
		e := recover()
		if e == nil {
			t.Errorf("should report missing param")
		}
	}()

	rt.MustURLMap(map[string]string{"param1": "val1"})
}

func TestURLMissingParam(t *testing.T) {
	rt := New("/route/:param1/:param2", method.GET)
	Mount("/", rt)
	defer func() {
		e := recover()
		if e == nil {
			t.Errorf("should report missing param")
		}
	}()

	rt.MustURL("param1", "val1")
}

func TestURLMissingValue(t *testing.T) {
	rt := New("/route/:param", method.GET)
	// Router = PseudoRouter("/")
	Mount("/", rt)
	defer func() {
		e := recover()
		if e == nil {
			t.Errorf("should report missing value")
		}
	}()

	rt.MustURL("param")
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
