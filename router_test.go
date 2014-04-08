package router

import (
	"fmt"
	"net/http"

	"github.com/go-on/method"
	"github.com/go-on/wrap-contrib/wraps"
	. "launchpad.net/gocheck"
)

type routerSuite struct{}

var _ = Suite(&routerSuite{})

func mkEtaggedRouter() *Router {
	router := NewETagged()
	router.MustHandle("/a", method.GET, webwrite("A"))
	router.MustHandle("/a", method.PATCH, webwrite("OK"))
	mount(router, "/")
	return router
}

func (s *routerSuite) TestRouterEtag(c *C) {
	rec, req := newTestRequest("GET", "/a")
	mkEtaggedRouter().ServeHTTP(rec, req)
	c.Assert(rec.Header().Get("Etag"), Equals, "7fc56270e7a70fa81a5935b72eacbe29")
	c.Assert(rec.Code, Equals, 200)
	c.Assert(rec.Body.String(), Equals, "A")
}

func (s *routerSuite) TestRoute(c *C) {
	r := New()
	rt := r.GET("/hu", webwrite("hu"))

	c.Assert(r.Route("/hu"), Equals, rt)
}

func (s *routerSuite) TestParent(c *C) {
	p := New()

	ch := New()
	p.GET("/ch", ch)

	c.Assert(ch.Parent(), Equals, p)
}

func (s *routerSuite) TestSubmountAlreadyMounted(c *C) {
	p := New()

	ch := New()
	mount(ch, "/hu")
	defer func() {
		e := recover()
		c.Assert(e, Not(Equals), nil)
	}()
	p.GET("/ch", ch)
}

func (s *routerSuite) TestTrace(c *C) {
	r := New()
	r.TRACE("/ho", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("hu", "ho")
	}))
	mount(r, "/")

	rw, req := newTestRequest("TRACE", "/ho")

	r.ServeHTTP(rw, req)
	c.Assert(rw.Header().Get("hu"), Equals, "ho")
}

func (s *routerSuite) TestEachRoute(c *C) {
	r := New()
	a := r.GET("/ho", webwrite("ho"))
	b := r.POST("/hu", webwrite("hu"))

	mount(r, "/hi")

	rts := map[string]*Route{}

	r.EachRoute(func(mountPoint string, route *Route) {
		// fmt.Println(mountPoint)
		rts[mountPoint] = route
	})

	c.Assert(rts["/ho"], Equals, a)
	c.Assert(rts["/hu"], Equals, b)
}

func (s *routerSuite) TestOptions(c *C) {
	r := New()
	r.OPTIONS("/ho", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("hu", "ho")
	}))
	mount(r, "/")

	rw, req := newTestRequest("OPTIONS", "/ho")

	r.ServeHTTP(rw, req)
	c.Assert(rw.Header().Get("hu"), Equals, "ho")

}

func (s *routerSuite) TestHead(c *C) {
	r := New()
	r.HEAD("/hu", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("hu", "ho")
	}))

	mount(r, "/hi")

	rw, req := newTestRequest("HEAD", "/hi/hu")

	r.ServeHTTP(rw, req)
	c.Assert(rw.Header().Get("hu"), Equals, "ho")

}

func (s *routerSuite) TestDelete(c *C) {
	r := New()
	r.DELETE("/hu", webwrite("deleted"))

	mount(r, "/")

	rw, req := newTestRequest("DELETE", "/hu")

	r.ServeHTTP(rw, req)
	c.Assert(rw.Body.String(), Equals, "deleted")
}

func (s *routerSuite) TestAddHandleAfterMount(c *C) {
	r := New()

	mount(r, "/hu")

	_, err := r.Handle("/ho", method.GET, webwrite("ho"))
	c.Assert(err, Not(Equals), nil)
}

func (s *routerSuite) TestMountPoint(c *C) {
	r := New()

	mount(r, "/hu")

	c.Assert(r.MountPoint(), Equals, "/hu")
}

func (s *routerSuite) TestDoubleMount(c *C) {
	r := New()

	mount(r, "/hu")
	defer func() {
		e := recover()
		c.Assert(e, Not(Equals), nil)
	}()
	mount(r, "/ho")

}

func (s *routerSuite) TestNotMounted(c *C) {
	r := New()
	r.GET("/hu", webwrite("hu"))

	rw, req := newTestRequest("GET", "/hu")

	defer func() {
		e := recover()
		c.Assert(e, Not(Equals), nil)
	}()
	r.ServeHTTP(rw, req)
}

func (s *routerSuite) TestRouterIfNoneMatchDoesMatch(c *C) {
	rec, req := newTestRequest("GET", "/a")
	req.Header.Set("If-None-Match", `"7fc56270e7a70fa81a5935b72eacbe29"`)
	mkEtaggedRouter().ServeHTTP(rec, req)
	c.Assert(rec.Code, Equals, 304)
	c.Assert(rec.Header().Get("Etag"), Equals, "7fc56270e7a70fa81a5935b72eacbe29")
	c.Assert(rec.Body.String(), Equals, "")
}

func (s *routerSuite) TestRouterIfNoneMatchDoesNotMatch(c *C) {
	rec, req := newTestRequest("GET", "/a")
	req.Header.Set("If-None-Match", `"nix"`)
	mkEtaggedRouter().ServeHTTP(rec, req)
	c.Assert(rec.Header().Get("Etag"), Equals, "7fc56270e7a70fa81a5935b72eacbe29")
	c.Assert(rec.Code, Equals, 200)
	c.Assert(rec.Body.String(), Equals, "A")
}

func (s *routerSuite) TestRouterIfMatchDoesMatch(c *C) {
	rec, req := newTestRequest("PATCH", "/a")
	req.Header.Set("If-Match", `"7fc56270e7a70fa81a5935b72eacbe29"`)
	mkEtaggedRouter().ServeHTTP(rec, req)
	c.Assert(rec.Header().Get("Etag"), Equals, "")
	c.Assert(rec.Code, Equals, 200)
	c.Assert(rec.Body.String(), Equals, "OK")
}

func (s *routerSuite) TestRouterIfMatchDoesNotMatch(c *C) {
	rec, req := newTestRequest("PATCH", "/a")
	req.Header.Set("If-Match", `"nix"`)
	mkEtaggedRouter().ServeHTTP(rec, req)
	c.Assert(rec.Code, Equals, 412)
	c.Assert(rec.Body.String(), Equals, "")
}

func (s *routerSuite) TestRouterEtaggedWithCustomWrappers(c *C) {

	router := NewETagged()
	router.AddWrappers(wraps.Before(webwrite("a")), wraps.Before(webwrite("b")))
	router.AddWrappers(wraps.Before(webwrite("c")), wraps.Before(webwrite("d")))

	router.MustHandle("/a", method.GET, webwrite("A"))
	mount(router, "/")

	rec, req := newTestRequest("GET", "/a")
	router.ServeHTTP(rec, req)

	c.Assert(rec.Code, Equals, 200)
	c.Assert(rec.Body.String(), Equals, "abcdA")
	c.Assert(rec.Header().Get("Etag"), Equals, "90c5f685703be163a3894ba83b6b57a2")

	router2 := NewETagged()
	router2.MustHandle("/a", method.GET, webwrite("abcdA"))
	mount(router2, "/")

	rec, req = newTestRequest("GET", "/a")
	router2.ServeHTTP(rec, req)

	c.Assert(rec.Code, Equals, 200)
	c.Assert(rec.Body.String(), Equals, "abcdA")
	c.Assert(rec.Header().Get("Etag"), Equals, "90c5f685703be163a3894ba83b6b57a2")

}

func (s *routeSuite) TestAllGETPaths(c *C) {

	router := New()
	router.GET("/a", webwrite("a"))
	rt := router.GET("/:b/:c/d", http.HandlerFunc(
		func(rw http.ResponseWriter, req *http.Request) {
			fmt.Fprintf(rw, "b: %s c: c", req.FormValue(":b"), req.FormValue(":c"))
		}))

	solver := RouteParameterFunc(func(route *Route) []map[string]string {
		if route == rt {
			return []map[string]string{
				map[string]string{
					"b": "b1",
					"c": "c1",
				},
				map[string]string{
					"b": "b2",
					"c": "c2",
				},
			}
		}
		return nil
	})

	paths := router.AllGETPaths(solver)

	c.Assert(len(paths), Equals, 3)
	c.Assert(paths[0], Equals, "/a")
	c.Assert(paths[1], Equals, "/b1/c1/d")
	c.Assert(paths[2], Equals, "/b2/c2/d")

	// fmt.Printf("paths: %v", paths)

}
