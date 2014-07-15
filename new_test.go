package router

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/go-on/router/route"
)

type writeParam struct {
	text   string
	params []string
}

func (w writeParam) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "%s %s|", req.Method, w.text)
	for _, param := range w.params {
		fmt.Fprintf(rw, "%s:%s,", param, GetRouteParam(req, param))
	}
}

func writeParams(text string, params ...string) writeParam {
	return writeParam{text, params}
}

func write(text string) writeParam {
	return writeParam{text: text}
}

func TestNewVariant(t *testing.T) {
	a := route.NewRoute("/:x/a.html")
	a.GETHandler = writeParams("A", "x")
	a.POSTHandler = writeParams("A", "x")

	// fmt.Printf("route a is %p\n", a)

	rtr := New()
	rtr.MustAdd(a)
	rtr.Mount("/", nil)

	rec, req := newTestRequest("GET", "/y/a.html")
	rtr.ServeHTTP(rec, req)

	body := rec.Body.String()

	exp := "GET A|x:y,"
	if body != exp {
		t.Errorf("expected %#v, got: %#v", exp, body)
	}

	rec, req = newTestRequest("POST", "/z/a.html")
	rtr.ServeHTTP(rec, req)

	body = rec.Body.String()
	exp = "POST A|x:z,"
	if body != exp {
		t.Errorf("expected %#v, got: %#v", exp, body)
	}
	// assertResponse(c, rw, "A", 200)
}

func TestNewVariant2(t *testing.T) {
	a := route.NewRoute("/a.html")
	a.GETHandler = write("A")
	a.POSTHandler = write("A")

	b := route.NewRoute("/:sth/x.html")
	b.GETHandler = writeParams("B", "sth")

	// fmt.Printf("route a is %p\n", a)

	rtr := New()
	rtr.MustAdd(a)
	rtr.MustAdd(b)
	rtr.Mount("/", nil)

	rec, req := newTestRequest("GET", "/a.html")
	rtr.ServeHTTP(rec, req)

	body := rec.Body.String()
	exp := "GET A|"
	if body != exp {
		t.Errorf("expected %#v, got: %#v", exp, body)
	}

	rec, req = newTestRequest("POST", "/a.html")
	rtr.ServeHTTP(rec, req)

	body = rec.Body.String()
	exp = "POST A|"
	if body != exp {
		t.Errorf("expected %#v, got: %#v", exp, body)
	}

	rec, req = newTestRequest("GET", "/x/x.html")
	rtr.ServeHTTP(rec, req)

	body = rec.Body.String()
	exp = "GET B|sth:x,"
	if body != exp {
		t.Errorf("expected %#v, got: %#v", exp, body)
	}
	// assertResponse(c, rw, "A", 200)
}

func TestNewMounted(t *testing.T) {
	a := route.NewRoute("/:x/:p/a/:b")
	a.GETHandler = writeParams("A", "x", "p", "b")
	a.POSTHandler = writeParams("A", "x", "p", "b")

	// fmt.Printf("route a is %p\n", a)

	rtr := New()
	rtr.MustAdd(a)
	rtr.Mount("/ho", nil)

	rec, req := newTestRequest("GET", "/ho/y/f/a/q")
	rtr.ServeHTTP(rec, req)

	body := rec.Body.String()
	exp := "GET A|x:y,p:f,b:q,"
	if body != exp {
		t.Errorf("expected %#v, got: %#v", exp, body)
	}

	rec, req = newTestRequest("POST", "/ho/z/g/a/r")
	rtr.ServeHTTP(rec, req)

	body = rec.Body.String()
	exp = "POST A|x:z,p:g,b:r,"
	if body != exp {
		t.Errorf("expected %#v, got: %#v", exp, body)
	}
	// assertResponse(c, rw, "A", 200)
}

func TestNewSub(t *testing.T) {
	zero := route.NewRoute("/zero")
	a := route.NewRoute("/:x/:p/a/:b")
	a.GETHandler = writeParams("A", "x", "p", "b")

	// fmt.Printf("route a is %p\n", a)

	rtr := New()
	rtr.MustAdd(a)
	zero.GETHandler = rtr

	outer := New()
	outer.MustAdd(zero)
	outer.Mount("/ho", nil)

	rec, req := newTestRequest("GET", "/ho/zero/y/f/a/q")
	outer.ServeHTTP(rec, req)

	body := rec.Body.String()
	exp := "GET A|x:y,p:f,b:q,"
	if body != exp {
		t.Errorf("expected %#v, got: %#v", exp, body)
	}

}

/*
func TestFindParams(t *testing.T) {
	corpus := map[string]string{
		"abc/def/key/val1/mno/pqr//0xfun":  "val1",
		"key/val2/mno/pqr//0xfun":          "val2",
		"abc/def/xyz/val1/key/val3//0xfun": "val3",
	}

	for k, v := range corpus {
		start, end := findParam(k, "key")
		got := ""
		if start > -1 && end > -1 {
			got = k[start:end]
		}
		if got != v {
			t.Errorf("expected: %#v, got: %#v", v, got)
		}
	}

}
*/
