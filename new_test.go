package router

import (
	"testing"

	"github.com/go-on/router/route"
)

func TestNewVariant(t *testing.T) {
	a := route.NewRoute("a.html")
	a.GETHandler = webwrite("A-GET")
	a.POSTHandler = webwrite("A-POST")

	// fmt.Printf("route a is %p\n", a)

	rtr := New()
	rtr.MustRegisterRoute2(a)
	rtr.Mount("/", nil)

	rec, req := newTestRequest("GET", "/a.html")
	rtr.ServeHTTP(rec, req)

	body := rec.Body.String()

	if body != "A-GET" {
		t.Errorf("expected A-GET, got: %#v", body)
	}

	rec, req = newTestRequest("POST", "/a.html")
	rtr.ServeHTTP(rec, req)

	body = rec.Body.String()

	if body != "A-POST" {
		t.Errorf("expected A-POST, got: %#v", body)
	}
	// assertResponse(c, rw, "A", 200)
}
