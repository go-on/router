package router

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-on/method"
	"github.com/go-on/router/route"
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
		{"/a.html", "A", 200},
		{"/b.html", "B", 200},
		{"/x.html", "", 404},
		{"/a/x.html", "AX", 200},
		{"/a/b.html", "AB", 200},
		{"/b/x.html", "BX", 200},
		{"/:sth/x.html", "SthX", 200},
	}

	router := New()
	router.AddWrappers(mw...)
	for _, r := range corpus {
		if r.code == 200 {
			//router.GET(r.path, method.GET, webwrite(r.body))
			router.GET(r.path, webwrite(r.body))
		}
	}
	return router
}

func (s *routeSuite) TestRouting(c *C) {
	corpus := []routetest{
		{"/a.html", "A", 200},
		{"/b.html", "B", 200},
		{"/x.html", "", 405},
		{"/a/x.html", "AX", 200},
		{"/b/x.html", "BX", 200},
		{"/z/x.html", "SthX", 200},
		{"/y.html", "", 405},
	}

	router := mount(makeRouter(), "/")

	for _, r := range corpus {
		rw, req := newTestRequest("GET", r.path)
		router.ServeHTTP(rw, req)
		assertResponse(c, rw, r.body, r.code)
	}
}

func (s *routeSuite) TestRouter(c *C) {
	r := New()
	rt := r.GET("/hu", webwrite("hu"))
	c.Assert(rt.Router, Equals, r)
}

func (s *routeSuite) TestURLWrongParams(c *C) {
	r := New()
	rt := r.GET("/", webwrite("hu"))
	defer func() {
		e := recover()
		c.Assert(e, Not(Equals), nil)
	}()
	rt.URL("hu")
}

func (s *routeSuite) TestURLWrongParams2(c *C) {
	r := New()
	rt := r.GET("/:hi", webwrite("hu"))
	defer func() {
		e := recover()
		c.Assert(e, Not(Equals), nil)
	}()
	rt.MustURL("hu", "ho")
}

func (s *routeSuite) TestURLMapWrongParams(c *C) {
	r := New()
	rt := r.GET("/:hi", webwrite("hu"))
	defer func() {
		e := recover()
		c.Assert(e, Not(Equals), nil)
	}()
	rt.MustURLMap(map[string]string{"hu": "ho"})
}

func (s *routeSuite) TestURLStructWrongParams(c *C) {
	r := New()
	rt := r.GET("/", webwrite("hu"))
	_, err := URLStruct(rt, "hu", "ho")
	c.Assert(err, Not(Equals), nil)
}

func (s *routeSuite) TestMustURLStructWrongParams(c *C) {
	r := New()
	rt := r.GET("/", webwrite("hu"))
	defer func() {
		e := recover()
		c.Assert(e, Not(Equals), nil)
	}()
	MustURLStruct(rt, "hu", "ho")
}

func (s *routeSuite) TestNotExistingMethod(c *C) {
	r := New()
	r.POST("/", webwrite("hu"))
	router := mount(r, "/")

	rw, req := newTestRequest("GETTO", "/")
	router.ServeHTTP(rw, req)
	c.Assert(rw.Body.String(), Equals, "")
	c.Assert(rw.Code, Equals, 405)
}

func (s *routeSuite) TestRoutingMiddlewareMounted(c *C) {
	corpus := []routetest{
		{"/mount/a.html", "#A#", 200},
		{"/mount/b.html", "#B#", 200},
		{"/mount/x.html", "", 405},
		{"/mount/a/x.html", "#AX#", 200},
		{"/mount/b/x.html", "#BX#", 200},
		{"/mount/z/x.html", "#SthX#", 200},
		{"/mount/y.html", "", 405},
		{"/a.html", "", 405},
		{"/z/x.html", "", 405},
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
		{"/a.html", "#A#", 200},
		{"/b.html", "#B#", 200},
		{"/x.html", "", 405},
		{"/a/x.html", "#AX#", 200},
		{"/b/x.html", "#BX#", 200},
		{"/z/x.html", "#SthX#", 200},
		{"/y.html", "", 405},
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
		{"/mount/a.html", "A", 200},
		{"/mount/b.html", "B", 200},
		{"/mount/x.html", "", 405},
		{"/mount/a/x.html", "AX", 200},
		{"/mount/b/x.html", "BX", 200},
		{"/mount/z/x.html", "SthX", 200},
		{"/mount/y.html", "", 405},
		{"/a.html", "", 405},
		{"/z/x.html", "", 405},
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
		{"/outer/a.html", "A", 200},
		{"/outer/b.html", "B", 200},
		{"/outer/x.html", "", 405},
		{"/outer/a/x.html", "AX", 200},
		{"/outer/b/x.html", "BX", 200},
		{"/outer/z/x.html", "SthX", 200},
		{"/outer/y.html", "", 405},
		{"/a.html", "", 405},
		{"/z/x.html", "", 405},

		{"/outer/inner/a.html", "A", 200},
		{"/outer/inner/b.html", "B", 200},
		{"/outer/inner/a/x.html", "AX", 200},
		{"/outer/inner/b/x.html", "BX", 200},
		{"/outer/inner/z/x.html", "SthX", 200},
		{"/outer/inner/y.html", "", 405},
		{"/inner/a.html", "", 405},
		{"/inner/z/x.html", "", 405},
	}
	inner := makeRouter()
	outer := makeRouter()
	outer.GET("/inner", inner)

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
		{"/outer/a.html", "#A#", 200},
		{"/outer/b.html", "#B#", 200},
		{"/outer/x.html", "", 405},
		{"/outer/a/x.html", "#AX#", 200},
		{"/outer/b/x.html", "#BX#", 200},
		{"/outer/z/x.html", "#SthX#", 200},
		{"/outer/y.html", "", 405},
		{"/a.html", "", 405},
		{"/z/x.html", "", 405},
		{"z.html", "", 405},

		{"/outer/inner/a.html", "#~A~#", 200},
		{"/outer/inner/b.html", "#~B~#", 200},
		{"/outer/inner/a/x.html", "#~AX~#", 200},
		{"/outer/inner/b/x.html", "#~BX~#", 200},
		{"/outer/inner/z/x.html", "#~SthX~#", 200},
		{"/outer/inner/y.html", "", 405},
		{"/inner/a.html", "", 405},
		{"/inner/z/x.html", "", 405},
	}

	inner := makeRouter(wrr.Around(webwrite("~"), webwrite("~")))
	outer := makeRouter(wrr.Around(webwrite("#"), webwrite("#")))
	outer.GET("/inner", inner)

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
	r.POST("/a.html", webwrite("A-POST"))
	router := mount(r, "/")
	router.SetOPTIONSHandlers()

	rw, req := newTestRequest("GET", "/a.html")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "A", 200)

	rw, req = newTestRequest("POST", "/a.html")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "A-POST", 200)

	rw, req = newTestRequest("OPTIONS", "/a.html")
	router.ServeHTTP(rw, req)
	allow := rw.HeaderMap.Get("Allow")
	// println(allow)
	c.Assert(strings.Contains(allow, "OPTIONS"), Equals, true)
	c.Assert(strings.Contains(allow, "GET"), Equals, true)
	c.Assert(strings.Contains(allow, "POST"), Equals, true)
	c.Assert(strings.Contains(allow, "HEAD"), Equals, true)
}

func (s *routeSuite) TestRoutingHandlerAndSubroutes(c *C) {
	inner := New()
	inner.AddWrappers(wrr.Around(webwrite("~"), webwrite("~")))
	inner.POST("/b.html", webwrite("B-POST"))
	inner2 := New()
	inner2.AddWrappers(wrr.Around(webwrite("~"), webwrite("~")))
	inner2.POST("/b.html", webwrite("B-POST"))

	outer := New()
	outer.AddWrappers(wrr.Around(webwrite("#"), webwrite("#")))
	outer.POST("/a", inner)
	outer.POST("/other", inner2)

	outer.SetAllOPTIONSHandlers()

	//	fmt.Println(outer.Inspect(0))
	router := mount(outer, "/mount")
	// fmt.Println(router.Inspect(0))

	rw, req := newTestRequest("POST", "/mount/a/b.html")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "#~B-POST~#", 200)

	rw, req = newTestRequest("POST", "/mount/other/b.html")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "#~B-POST~#", 200)

	rw, req = newTestRequest("OPTIONS", "/mount/other/b.html")
	router.ServeHTTP(rw, req)

	allow := rw.HeaderMap.Get("Allow")
	c.Assert(strings.Contains(allow, "OPTIONS"), Equals, true)
	c.Assert(strings.Contains(allow, "GET"), Equals, false)
	c.Assert(strings.Contains(allow, "POST"), Equals, true)
	c.Assert(strings.Contains(allow, "HEAD"), Equals, false)
}

func (s *routeSuite) TestRoutingHandlerCombined(c *C) {
	inner := New()
	inner.AddWrappers(wrr.Around(webwrite("~"), webwrite("~")))
	inner.GET("/", webwrite("INNER-ROOT"))
	inner.HandleMethods("/a.html", webwrite("A-INNER-GET-POST"), method.GET, method.POST)

	outer := New()
	outer.AddWrappers(wrr.FilterBody(method.PATCH), wrr.Around(webwrite("#"), webwrite("#")))
	outer.HandleMethods("/a.html", webwrite("A-OUTER-GET-POST"), method.GET, method.POST)

	outer.HandleMethods("/inner", inner, method.GET, method.POST)

	outer.SetAllOPTIONSHandlers()

	_ = fmt.Println
	//	fmt.Println(outer.Inspect(0))
	router := mount(outer, "/mount")
	// fmt.Println(router.Inspect(0))
	rw, req := newTestRequest("GET", "/mount/a.html")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "#A-OUTER-GET-POST#", 200)

	rw, req = newTestRequest("POST", "/mount/a.html")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "#A-OUTER-GET-POST#", 200)

	rw, req = newTestRequest("GET", "/mount/inner")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "#~INNER-ROOT~#", 200)

	rw, req = newTestRequest("OPTIONS", "/mount/a.html")
	router.ServeHTTP(rw, req)

	allow := rw.HeaderMap.Get("Allow")
	c.Assert(strings.Contains(allow, "OPTIONS"), Equals, true)
	c.Assert(strings.Contains(allow, "GET"), Equals, true)
	c.Assert(strings.Contains(allow, "POST"), Equals, true)
	c.Assert(strings.Contains(allow, "HEAD"), Equals, true)
	assertResponse(c, rw, "", 200)

	rw, req = newTestRequest("GET", "/mount/inner/a.html")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "#~A-INNER-GET-POST~#", 200)

	rw, req = newTestRequest("POST", "/mount/inner/a.html")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "#~A-INNER-GET-POST~#", 200)

	rw, req = newTestRequest("PATCH", "/mount/inner/a.html")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "", 405)

}

func (s *routeSuite) TestRoutingSubRouteRoot(c *C) {
	admin := New()
	admin.GET("/", webwrite("ADMIN"))
	index := New()
	index.GET("/admin", admin)

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
	var rt *route.Route

	for i := 0; i < 10000; i++ {
		rt = inner.GET(fmt.Sprintf("/r%d", i), webwrite(fmt.Sprintf("r%d", i)))
	}

	index := New()
	index.GET("/admin", inner)

	// fmt.Println("mounting flat")
	router := mount(index, "/index")

	// fmt.Println("new request")
	_ = rt
	c.Assert(rt.MustURL(), Equals, "/index/admin/r9999")
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
	inner.GET("/admin", webwrite("ADMIN"))
	var route *route.Route
	// var r2 *Route

	for i := 0; i < 10001; i++ {
		//for i := 0; i < 100; i++ {
		//for i := 0; i < 500; i++ {
		r = New()
		r.GET(fmt.Sprintf("/i%d", i), inner)
		route = r.GET(fmt.Sprintf("/r%d", i), webwrite(fmt.Sprintf("r%d", i)))
		inner = r

		// route = r.GET(fmt.Sprintf("/a%d", i), method.GET, webwrite(fmt.Sprintf("a%d", i)))

		//	fmt.Print(".")
		// fmt.Printf("/r%d\n", i)
	}

	index := New()
	index.GET("/admin", inner)

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
	// req.ParseForm()
	// fmt.Printf("%#v\n", req.PostForm)
	vv.x = GetRouteParam(req, "x")
	vv.y = GetRouteParam(req, "y")
}

func (s *routeSuite) TestVars(c *C) {
	vv := &v{}
	r := New()
	rt := r.GET("/a/:x/c/:y", vv)
	router := mount(r, "/r")
	rw, req := newTestRequest("GET", "/r/a/b/c/d")
	router.ServeHTTP(rw, req)
	//assertResponse(c, rw, "ADMIN", 200)
	c.Assert(vv.x, Equals, "b")
	c.Assert(vv.y, Equals, "d")
	// println(rt.Id)
	c.Assert(GetRouteId(req), Equals, rt.Id)
}

/*
func (s *routeSuite) TestVarsSubrouterPanic(c *C) {
	vv := &v{}
	r := New()
	sub := New()
	sub.GET("/:y", vv)
	defer func() {
		e := recover()
		c.Assert(e, Not(Equals), nil)
	}()

	r.GET("/a/:x/sub", sub)
}
*/

type ctx struct {
	App  string `var:"app"`
	path string
	http.ResponseWriter
}

func (c *ctx) SetPath(w http.ResponseWriter, r *http.Request) {
	c.path = r.URL.Path
}

func (c *ctx) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("app: " + c.App + " path: " + c.path))
}

func mkCtx(wr http.ResponseWriter, req *http.Request) http.ResponseWriter {
	return &ctx{ResponseWriter: wr}
}

func (s *routeSuite) TestVarsSetStruct(c *C) {
	// ct := &ctx{}
	r := New()
	r.AddWrappers(
		wr.Context(mkCtx),
		//	wrr.Before(wr.HandlerMethod((*ctx).SetVars)),
		wrr.Before(wr.HandlerMethod((*ctx).SetPath)),
	)
	r.GET("/app/:app/hiho.html", wr.HandlerMethod((*ctx).ServeHTTP))

	router := mount(r, "/r")
	rw, req := newTestRequest("GET", "/r/app/X/hiho.html")
	router.ServeHTTP(rw, req)
	// fmt.Printf("c.App: %s\n", ct.App)
	assertResponse(c, rw, "app:  path: /r/app/X/hiho.html", 200)
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
	index1.GET("/admin1", admin1)
	index2 := New()
	index2.GET("/admin2", admin2)

	mount(index1, "/index1")
	mount(index2, "/index2")

	url1 := route1.MustURL()
	// return
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
	url5 := MustURLStruct(route2, &str1, "urltest")
	c.Assert(url5, Equals, "/index1/admin1/q/z")
}
