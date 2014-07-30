package router

import (
	"testing"

	"github.com/go-on/method"

	"github.com/go-on/router/route"
)

func TestDoubleMount(t *testing.T) {
	router := New()
	router.Mount("/", nil)

	defer func() {
		e := recover()
		errMsg := errorMustBe(e, ErrDoubleMounted{})

		if errMsg != "" {
			t.Error(errMsg)
			return
		}

		err := e.(ErrDoubleMounted)
		_ = err.Error()

		if err.Path != "/" {
			t.Errorf("wrong path: %#v, expected: %v", err.Path, "/")
		}
	}()

	router.Mount("/double", nil)
}

func TestDoubleMountSub(t *testing.T) {
	sub := New()
	router := New()
	sub.Mount("/", router)

	defer func() {
		e := recover()
		errMsg := errorMustBe(e, ErrDoubleMounted{})

		if errMsg != "" {
			t.Error(errMsg)
			return
		}

		err := e.(ErrDoubleMounted)
		_ = err.Error()

		if err.Path != "/" {
			t.Errorf("wrong path: %#v, expected: %v", err.Path, "/")
		}
	}()

	sub.Mount("/double", router)
}

func TestInvalidMountPathSub(t *testing.T) {
	sub := New()
	router := New()

	defer func() {
		e := recover()
		errMsg := errorMustBe(e, ErrInvalidMountPath{})

		if errMsg != "" {
			t.Error(errMsg)
			return
		}

		err := e.(ErrInvalidMountPath)
		_ = err.Error()

		if err.Path != "/:invalid" {
			t.Errorf("wrong path: %#v, expected: %v", err.Path, "/:invalid")
		}
	}()

	sub.Mount("/:invalid", router)
}

func TestNotMounted(t *testing.T) {
	router := New()

	defer func() {
		e := recover()
		errMsg := errorMustBe(e, ErrNotMounted{})

		if errMsg != "" {
			t.Error(errMsg)
			return
		}

		err := e.(ErrNotMounted)
		_ = err.Error()
	}()

	router.ServeHTTP(nil, nil)
}

func TestInvalidMountPath(t *testing.T) {
	router := New()

	defer func() {
		e := recover()
		errMsg := errorMustBe(e, ErrInvalidMountPath{})

		if errMsg != "" {
			t.Error(errMsg)
			return
		}

		err := e.(ErrInvalidMountPath)
		_ = err.Error()

		if err.Path != "/:invalid" {
			t.Errorf("wrong path: %#v, expected: %v", err.Path, "/:invalid")
		}
	}()

	router.Mount("/:invalid", nil)
}

func TestDoubleRegistration(t *testing.T) {
	route1 := route.New("/double", method.GET)
	route2 := route.New("/double", method.POST)
	router := New()
	router.mustAddRoute(newRouteHandler(route1))

	defer func() {
		e := recover()
		errMsg := errorMustBe(e, ErrDoubleRegistration{})

		if errMsg != "" {
			t.Error(errMsg)
			return
		}

		err := e.(ErrDoubleRegistration)
		_ = err.Error()

		if err.DefinitionPath != "/double" {
			t.Errorf("wrong definition path: %#v, expected: %v", err.DefinitionPath, "/double")
		}
	}()

	router.mustAddRoute(newRouteHandler(route2))
	// router.Mount("/", nil)
}

func TestDoubleRegistrationSub(t *testing.T) {
	sub := New()
	sub.path = "/first"
	router := New()

	defer func() {
		e := recover()
		errMsg := errorMustBe(e, ErrDoubleRegistration{})

		if errMsg != "" {
			t.Error(errMsg)
			return
		}

		err := e.(ErrDoubleRegistration)
		_ = err.Error()

		if err.DefinitionPath != "/first" {
			t.Errorf("wrong definition path: %#v, expected: %v", err.DefinitionPath, "/first")
		}
	}()

	sub.Mount("/double", router)
}

func TestRouterNotAllowed(t *testing.T) {
	sub := New()
	router := New()

	defer func() {
		e := recover()
		errMsg := errorMustBe(e, ErrRouterNotAllowed{})

		if errMsg != "" {
			t.Error(errMsg)
			return
		}

		err := e.(ErrRouterNotAllowed)
		_ = err.Error()
	}()

	router.GET("/sub", sub)
}

func TestMethodNotDefinedForRoute(t *testing.T) {
	rt := route.New("/", method.POST)
	router := New()

	defer func() {
		e := recover()
		errMsg := errorMustBe(e, &ErrMethodNotDefinedForRoute{})

		if errMsg != "" {
			t.Error(errMsg)
			return
		}

		err := e.(*ErrMethodNotDefinedForRoute)
		_ = err.Error()

		if err.Method != method.GET {
			t.Errorf("wrong method: %#v, expected: %v", err.Method, method.GET)
		}

		if err.Route != rt {
			t.Errorf("wrong route: %v, expected: %v", err.Route, rt)
		}
	}()

	router.HandleRouteMethods(rt, writeString("a"), method.GET, method.POST)
}

func TestErrMissingHandler(t *testing.T) {
	rt := route.New("/missing_handler", method.GET, method.POST)
	router := New()
	router.HandleRouteMethods(rt, writeString("a"), method.POST)

	defer func() {
		e := recover()
		errMsg := errorMustBe(e, &ErrMissingHandler{})

		if errMsg != "" {
			t.Error(errMsg)
			return
		}

		err := e.(*ErrMissingHandler)
		_ = err.Error()

		if len(err.methods) != 1 {
			t.Errorf("wrong number of methods: %d, expected: 1", len(err.methods))
		}

		if err.methods[0] != method.GET {
			t.Errorf("wrong method: %#v, expected: %v", err.methods[0], method.GET)
		}

		if err.Route != rt {
			t.Errorf("wrong route: %v, expected: %v", err.Route, rt)
		}
	}()

	router.Mount("/", nil)
}
