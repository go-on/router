package router

import (
	"github.com/go-on/method"
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
