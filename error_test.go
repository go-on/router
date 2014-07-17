package router

import (
	"testing"

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
	route1 := route.NewRoute("/double")
	route2 := route.NewRoute("/double")
	router := New()
	router.MustAddRoute(route1)

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

	router.MustAddRoute(route2)
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
