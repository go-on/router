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
