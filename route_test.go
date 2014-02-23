package router

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-on/method"
	"github.com/go-on/wrap"
	wr "github.com/go-on/wrap-contrib-testing/wrapstesting"
	wrr "github.com/go-on/wrap-contrib/wraps"
	. "launchpad.net/gocheck"
)

type routetest struct {
	path string
	body string
	code int
}

type routeSuite struct{}

var _ = Suite(&routeSuite{})

func makeRouter(mw ...wrap.Wrapper) *Router {
	corpus := []routetest{
		{"/a", "A", 200},
		{"/b", "B", 200},
		{"/x", "", 404},
		{"/a/x", "AX", 200},
		{"/a/b", "AB", 200},
		{"/b/x", "BX", 200},
		{"/:sth/x", "SthX", 200},
	}

	router := New(mw...)
	for _, r := range corpus {
		if r.code == 200 {
			router.MustHandle(r.path, method.GET, webwrite(r.body))
		}
	}
	return router
}

func (s *routeSuite) TestRouting(c *C) {
	corpus := []routetest{
		{"/a", "A", 200},
		{"/b", "B", 200},
		{"/x", "", 405},
		{"/a/x", "AX", 200},
		{"/b/x", "BX", 200},
		{"/z/x", "SthX", 200},
		{"/y", "", 405},
	}

	router := mount(makeRouter(), "/")

	for _, r := range corpus {
		rw, req := newTestRequest("GET", r.path)
		router.ServeHTTP(rw, req)
		assertResponse(c, rw, r.body, r.code)
	}
}

func (s *routeSuite) TestDoubleRoute(c *C) {
	r := New()
	r.MustHandle("/hu", method.GET, webwrite("hu"))

	defer func() {
		e := recover()
		c.Assert(e, Not(Equals), nil)
	}()

	r.MustHandle("/hu", method.GET, webwrite("ho"))

}

func (s *routeSuite) TestRoutingMiddlewareMounted(c *C) {
	corpus := []routetest{
		{"/mount/a", "#A#", 200},
		{"/mount/b", "#B#", 200},
		{"/mount/x", "", 405},
		{"/mount/a/x", "#AX#", 200},
		{"/mount/b/x", "#BX#", 200},
		{"/mount/z/x", "#SthX#", 200},
		{"/mount/y", "", 405},
		{"/a", "", 405},
		{"/z/x", "", 405},
	}

	router := mount(makeRouter(wrr.Around(webwrite("#"), webwrite("#"))), "/mount")
	for _, r := range corpus {
		rw, req := newTestRequest("GET", r.path)
		router.ServeHTTP(rw, req)
		assertResponse(c, rw, r.body, r.code)
	}
}

func (s *routeSuite) TestRoutingMiddleware(c *C) {
	corpus := []routetest{
		{"/a", "#A#", 200},
		{"/b", "#B#", 200},
		{"/x", "", 405},
		{"/a/x", "#AX#", 200},
		{"/b/x", "#BX#", 200},
		{"/z/x", "#SthX#", 200},
		{"/y", "", 405},
	}

	router := mount(makeRouter(wrr.Around(webwrite("#"), webwrite("#"))), "/")
	for _, r := range corpus {
		rw, req := newTestRequest("GET", r.path)
		router.ServeHTTP(rw, req)
		// fmt.Printf("body: %s code: %d\n", r.body, rw.Code)
		assertResponse(c, rw, r.body, r.code)
	}
}

func (s *routeSuite) TestRoutingMounted(c *C) {
	corpus := []routetest{
		{"/mount/a", "A", 200},
		{"/mount/b", "B", 200},
		{"/mount/x", "", 405},
		{"/mount/a/x", "AX", 200},
		{"/mount/b/x", "BX", 200},
		{"/mount/z/x", "SthX", 200},
		{"/mount/y", "", 405},
		{"/a", "", 405},
		{"/z/x", "", 405},
	}

	router := mount(makeRouter(), "/mount")

	for _, r := range corpus {
		rw, req := newTestRequest("GET", r.path)
		router.ServeHTTP(rw, req)
		assertResponse(c, rw, r.body, r.code)
	}
}

func (s *routeSuite) TestRoutingSubroutes(c *C) {
	corpus := []routetest{
		{"/outer/a", "A", 200},
		{"/outer/b", "B", 200},
		{"/outer/x", "", 405},
		{"/outer/a/x", "AX", 200},
		{"/outer/b/x", "BX", 200},
		{"/outer/z/x", "SthX", 200},
		{"/outer/y", "", 405},
		{"/a", "", 405},
		{"/z/x", "", 405},

		{"/outer/inner/a", "A", 200},
		{"/outer/inner/b", "B", 200},
		{"/outer/inner/a/x", "AX", 200},
		{"/outer/inner/b/x", "BX", 200},
		{"/outer/inner/z/x", "SthX", 200},
		{"/outer/inner/y", "", 405},
		{"/inner/a", "", 405},
		{"/inner/z/x", "", 405},
	}
	inner := makeRouter()
	outer := makeRouter()
	outer.MustHandle("/inner", method.GET, inner)

	router := mount(outer, "/outer")
	_ = router
	for _, r := range corpus {
		rw, req := newTestRequest("GET", r.path)
		router.ServeHTTP(rw, req)
		assertResponse(c, rw, r.body, r.code)
	}
}

func (s *routeSuite) TestRoutingMiddlewareSubroutes(c *C) {
	corpus := []routetest{
		{"/outer/a", "#A#", 200},
		{"/outer/b", "#B#", 200},
		{"/outer/x", "", 405},
		{"/outer/a/x", "#AX#", 200},
		{"/outer/b/x", "#BX#", 200},
		{"/outer/z/x", "#SthX#", 200},
		{"/outer/y", "", 405},
		{"/a", "", 405},
		{"/z/x", "", 405},
		{"z", "", 405},

		{"/outer/inner/a", "#~A~#", 200},
		{"/outer/inner/b", "#~B~#", 200},
		{"/outer/inner/a/x", "#~AX~#", 200},
		{"/outer/inner/b/x", "#~BX~#", 200},
		{"/outer/inner/z/x", "#~SthX~#", 200},
		{"/outer/inner/y", "", 405},
		{"/inner/a", "", 405},
		{"/inner/z/x", "", 405},
	}

	inner := makeRouter(wrr.Around(webwrite("~"), webwrite("~")))
	outer := makeRouter(wrr.Around(webwrite("#"), webwrite("#")))
	outer.MustHandle("/inner", method.GET, inner)

	router := mount(outer, "/outer")
	//fmt.Println(router.Inspect(0))
	_ = router
	for _, r := range corpus {
		rw, req := newTestRequest("GET", r.path)
		router.ServeHTTP(rw, req)
		assertResponse(c, rw, r.body, r.code)
	}
}

func (s *routeSuite) TestRoutingVerbs(c *C) {
	r := makeRouter()
	r.MustHandle("/a", method.POST, webwrite("A-POST"))
	router := mount(r, "/")

	rw, req := newTestRequest("GET", "/a")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "A", 200)

	rw, req = newTestRequest("POST", "/a")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "A-POST", 200)

	rw, req = newTestRequest("OPTIONS", "/a")
	router.ServeHTTP(rw, req)
	allow := rw.HeaderMap.Get("Allow")
	c.Assert(strings.Contains(allow, "OPTIONS"), Equals, true)
	c.Assert(strings.Contains(allow, "GET"), Equals, true)
	c.Assert(strings.Contains(allow, "POST"), Equals, true)
	c.Assert(strings.Contains(allow, "HEAD"), Equals, true)
}

func (s *routeSuite) TestRoutingHandlerAndSubroutes(c *C) {
	inner := New(wrr.Around(webwrite("~"), webwrite("~")))
	inner.MustHandle("/b", method.POST, webwrite("B-POST"))
	inner2 := New(wrr.Around(webwrite("~"), webwrite("~")))
	inner2.MustHandle("/b", method.POST, webwrite("B-POST"))

	outer := New(wrr.Around(webwrite("#"), webwrite("#")))
	outer.MustHandle("/a", method.POST, inner)
	outer.MustHandle("/other", method.POST, inner2)

	//	fmt.Println(outer.Inspect(0))
	router := mount(outer, "/mount")
	// fmt.Println(router.Inspect(0))

	rw, req := newTestRequest("POST", "/mount/a/b")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "#~B-POST~#", 200)

	rw, req = newTestRequest("POST", "/mount/other/b")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "#~B-POST~#", 200)

	rw, req = newTestRequest("OPTIONS", "/mount/other/b")
	router.ServeHTTP(rw, req)

	allow := rw.HeaderMap.Get("Allow")
	c.Assert(strings.Contains(allow, "OPTIONS"), Equals, true)
	c.Assert(strings.Contains(allow, "GET"), Equals, false)
	c.Assert(strings.Contains(allow, "POST"), Equals, true)
	c.Assert(strings.Contains(allow, "HEAD"), Equals, false)
}

func (s *routeSuite) TestRoutingHandlerCombined(c *C) {
	inner := New(wrr.Around(webwrite("~"), webwrite("~")))
	inner.MustHandle("/", method.GET, webwrite("INNER-ROOT"))
	inner.MustHandle("/a", method.GET|method.POST, webwrite("A-INNER-GET-POST"))

	outer := New(wr.FilterBody(method.PATCH), wrr.Around(webwrite("#"), webwrite("#")))
	outer.MustHandle("/a", method.GET|method.POST, webwrite("A-OUTER-GET-POST"))

	outer.MustHandle("/inner", method.GET|method.POST, inner)

	_ = fmt.Println
	//	fmt.Println(outer.Inspect(0))
	router := mount(outer, "/mount")
	// fmt.Println(router.Inspect(0))
	rw, req := newTestRequest("GET", "/mount/a")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "#A-OUTER-GET-POST#", 200)

	rw, req = newTestRequest("POST", "/mount/a")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "#A-OUTER-GET-POST#", 200)

	rw, req = newTestRequest("GET", "/mount/inner")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "#~INNER-ROOT~#", 200)

	rw, req = newTestRequest("OPTIONS", "/mount/a")
	router.ServeHTTP(rw, req)

	allow := rw.HeaderMap.Get("Allow")
	c.Assert(strings.Contains(allow, "OPTIONS"), Equals, true)
	c.Assert(strings.Contains(allow, "GET"), Equals, true)
	c.Assert(strings.Contains(allow, "POST"), Equals, true)
	c.Assert(strings.Contains(allow, "HEAD"), Equals, true)
	assertResponse(c, rw, "", 200)

	rw, req = newTestRequest("GET", "/mount/inner/a")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "#~A-INNER-GET-POST~#", 200)

	rw, req = newTestRequest("POST", "/mount/inner/a")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "#~A-INNER-GET-POST~#", 200)

	rw, req = newTestRequest("PATCH", "/mount/inner/a")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "", 405)

}

func (s *routeSuite) TestRoutingSubRouteRoot(c *C) {
	admin := New()
	admin.MustHandle("/", method.GET, webwrite("ADMIN"))
	index := New()
	index.MustHandle("/admin", method.GET, admin)

	router := mount(index, "/index")

	rw, req := newTestRequest("GET", "/index/admin/")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "ADMIN", 200)

	rw, req = newTestRequest("GET", "/index/admin")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "ADMIN", 200)
}

func (s *routeSuite) TestMassiveRoutes(c *C) {
	var inner = New()
	var route *Route

	for i := 0; i < 10000; i++ {
		route = inner.MustHandle(fmt.Sprintf("/r%d", i), method.GET, webwrite(fmt.Sprintf("r%d", i)))
	}

	index := New()
	index.MustHandle("/admin", method.GET, inner)

	// fmt.Println("mounting flat")
	router := mount(index, "/index")

	// fmt.Println("new request")
	_ = route
	c.Assert(route.MustURL(), Equals, "/index/admin/r9999")
	// fmt.Println("serving")

	// for i := 0; i < 20000; i++ {
	rw, req := newTestRequest("GET", "/index/admin/r9998")
	router.ServeHTTP(rw, req)
	c.Assert(rw.Code, Equals, 200)
	assertResponse(c, rw, "r9998", 200)
	// }
}

func (s *routeSuite) TestMassiveRoutingNested(c *C) {
	var inner = New()
	var r *Router
	inner.MustHandle("/admin", method.GET, webwrite("ADMIN"))
	var route *Route
	// var r2 *Route

	for i := 0; i < 10001; i++ {
		//for i := 0; i < 100; i++ {
		//for i := 0; i < 500; i++ {
		r = New()
		r.MustHandle(fmt.Sprintf("/i%d", i), method.GET, inner)
		route = r.MustHandle(fmt.Sprintf("/r%d", i), method.GET, webwrite(fmt.Sprintf("r%d", i)))
		inner = r

		// route = r.MustHandle(fmt.Sprintf("/a%d", i), method.GET, webwrite(fmt.Sprintf("a%d", i)))

		//	fmt.Print(".")
		// fmt.Printf("/r%d\n", i)
	}

	index := New()
	index.MustHandle("/admin", method.GET, inner)

	// fmt.Println("mounting nested")
	router := mount(index, "/index")

	//fmt.Println(router.MustURL(r2))
	// fmt.Println("new request")

	c.Assert(route.MustURL(), Equals, "/index/admin/r10000")
	// fmt.Println("serving")

	// for i := 0; i < 20000; i++ {
	rw, req := newTestRequest("GET", "/index/admin/i10000/r9999")
	router.ServeHTTP(rw, req)
	c.Assert(rw.Code, Equals, 200)
	assertResponse(c, rw, "r9999", 200)
	// }
}

type v struct {
	x string
	y string
}

func (vv *v) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	vars := &Vars{}
	wr.UnWrap(rw, &vars)
	vv.x = vars.Get("x")
	vv.y = vars.Get("y")
}

func (s *routeSuite) TestVars(c *C) {
	vv := &v{}
	r := New()
	r.MustHandle("/a/:x/c/:y", method.GET, vv)
	router := mount(r, "/r")
	rw, req := newTestRequest("GET", "/r/a/b/c/d")
	router.ServeHTTP(rw, req)
	//assertResponse(c, rw, "ADMIN", 200)
	c.Assert(vv.x, Equals, "b")
	c.Assert(vv.y, Equals, "d")
}

type ctx struct {
	App  string `var:"app"`
	path string
	http.ResponseWriter
}

func (c *ctx) SetPath(w http.ResponseWriter, r *http.Request) {
	c.path = r.URL.Path
}

func (c *ctx) SetVars(vars *Vars, w http.ResponseWriter, r *http.Request) {
	vars.SetStruct(c, "var")
}

func (c *ctx) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("app: " + c.App + " path: " + c.path))
}

func (s *routeSuite) TestVarsSetStruct(c *C) {
	ct := &ctx{}
	r := New(
		wrr.Before(wr.HandlerMethod(ct.SetPath)),
		wrr.Before(wr.HandlerMethod(ct.SetVars)),
		wr.Context(ct))
	r.MustHandle("/app/:app/hiho", method.GET, wr.HandlerMethod(ct.ServeHTTP))

	router := mount(r, "/r")
	rw, req := newTestRequest("GET", "/r/app/X/hiho")
	router.ServeHTTP(rw, req)
	// fmt.Printf("c.App: %s\n", ct.App)
	assertResponse(c, rw, "app: X path: /r/app/X/hiho", 200)
}

type uStr1 struct {
	Y string `urltest:"y"`
}

func (s *routeSuite) TestURL(c *C) {
	admin1 := New()
	route1 := admin1.GET("/x", webwrite("ADMIN-X"))
	route2 := admin1.POST("/:y/z", webwrite("ADMIN-Z"))
	admin2 := New()
	route3 := admin2.PUT("/x", webwrite("ADMIN-X"))
	route4 := admin2.PATCH("/:y/z", webwrite("ADMIN-Z"))
	index1 := New()
	index1.MustHandle("/admin1", method.GET, admin1)
	index2 := New()
	index2.MustHandle("/admin2", method.GET, admin2)

	mount(index1, "/index1")
	mount(index2, "/index2")

	url1 := route1.MustURL()
	c.Assert(url1, Equals, "/index1/admin1/x")
	url2 := route3.MustURL()
	c.Assert(url2, Equals, "/index2/admin2/x")

	url3 := route2.MustURL("y", "p")
	c.Assert(url3, Equals, "/index1/admin1/p/z")

	url4 := route4.MustURL("y", "p")
	c.Assert(url4, Equals, "/index2/admin2/p/z")

	_, err := route2.URL()
	c.Assert(err, NotNil)

	str1 := uStr1{"q"}
	url5 := route2.MustURLStruct(&str1, "urltest")
	c.Assert(url5, Equals, "/index1/admin1/q/z")
}
