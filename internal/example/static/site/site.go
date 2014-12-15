package site

import (
	"fmt"
	"net/http"

	. "github.com/go-on/lib/html"

	"gopkg.in/go-on/router.v2"
)

func cHandler(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "c is %#v", router.GetRouteParam(req, "c"))
}

func dHandler(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "a is %#v, b is %#v, d is %#v whooho",
		router.GetRouteParam(req, "a"),
		router.GetRouteParam(req, "b"),
		router.GetRouteParam(req, "d"),
	)
}

func redirect(rw http.ResponseWriter, req *http.Request) {
	http.Redirect(rw, req, ARoute.MustURL(), 302)
}

func menu(rw http.ResponseWriter, req *http.Request) {
	// filepath.Rel(req.URL.String(), targpath)
	UL(
		LI(
			AHref(HomeRoute.MustURL(), "Home"),
		),
		LI(
			AHref(ARoute.MustURL(), "a"),
		),
		LI(
			AHref(DRoute.MustURL("a", "a0", "b", "b0", "d", "d0.html"), "d0"),
		),
		LI(
			AHref(DRoute.MustURL("a", "a1", "b", "b1", "d", "d1.html"), "d1"),
		),
		LI(
			AHref(DRoute.MustURL("a", "a2", "b", "b2", "d", "d2.html"), "d2"),
		),
	).WriteTo(rw)
}

type write string

func (s write) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprint(rw, s)
}

var (
	Router    = router.New()
	HomeRoute = Router.GET("/", write("index"))
	ARoute    = Router.GET("/a.html", write("A"))
	DRoute    = Router.GET("/d/:a/x/:b/:d", http.HandlerFunc(dHandler))
	Redirect  = Router.GETFunc("/redirect", redirect)
	App       = HTML5(
		HTML(
			BODY(
				HEADER(menu),
				Router,
			),
		),
	)
)
