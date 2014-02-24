package router

import (
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
	c.Assert(rec.Header().Get("Etag"), Equals, "8bf623a6e5c2a6cd374f3f18c3d51145")
	c.Assert(rec.Code, Equals, 200)
	c.Assert(rec.Body.String(), Equals, "A")
}

func (s *routerSuite) TestRouterIfNoneMatchDoesMatch(c *C) {
	rec, req := newTestRequest("GET", "/a")
	req.Header.Set("If-None-Match", `"8bf623a6e5c2a6cd374f3f18c3d51145"`)
	mkEtaggedRouter().ServeHTTP(rec, req)
	c.Assert(rec.Code, Equals, 304)
	c.Assert(rec.Header().Get("Etag"), Equals, "8bf623a6e5c2a6cd374f3f18c3d51145")
	c.Assert(rec.Body.String(), Equals, "")
}

func (s *routerSuite) TestRouterIfNoneMatchDoesNotMatch(c *C) {
	rec, req := newTestRequest("GET", "/a")
	req.Header.Set("If-None-Match", `"nix"`)
	mkEtaggedRouter().ServeHTTP(rec, req)
	c.Assert(rec.Header().Get("Etag"), Equals, "8bf623a6e5c2a6cd374f3f18c3d51145")
	c.Assert(rec.Code, Equals, 200)
	c.Assert(rec.Body.String(), Equals, "A")
}

func (s *routerSuite) TestRouterIfMatchDoesMatch(c *C) {
	rec, req := newTestRequest("PATCH", "/a")
	req.Header.Set("If-Match", `"8bf623a6e5c2a6cd374f3f18c3d51145"`)
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
	c.Assert(rec.Header().Get("Etag"), Equals, "a723816c4715934926760bd61b8e6fae")

	router2 := NewETagged()
	router2.MustHandle("/a", method.GET, webwrite("abcdA"))
	mount(router2, "/")

	rec, req = newTestRequest("GET", "/a")
	router2.ServeHTTP(rec, req)

	c.Assert(rec.Code, Equals, 200)
	c.Assert(rec.Body.String(), Equals, "abcdA")
	c.Assert(rec.Header().Get("Etag"), Equals, "a723816c4715934926760bd61b8e6fae")

}
