package site

import (
	"fmt"
	. "github.com/go-on/lib/html"
	"net/http"

	"github.com/go-on/router"
)

func cHandler(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "c is %#v", req.FormValue(":c"))
}

func dHandler(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "a is %#v, b is %#v, d is %#v whooho", req.FormValue(":a"), req.FormValue(":b"), req.FormValue(":d"))
}

func menu(rw http.ResponseWriter, req *http.Request) {
	// filepath.Rel(req.URL.String(), targpath)
	UL(
		LI(
			AHref(router.MustURL(HomeRoute), "Home"),
		),
		LI(
			AHref(router.MustURL(ARoute), "a"),
		),
		LI(
			AHref(router.MustURL(DRoute, "a", "a0", "b", "b0", "d", "d0.html"), "d0"),
		),
		LI(
			AHref(router.MustURL(DRoute, "a", "a1", "b", "b1", "d", "d1.html"), "d1"),
		),
		LI(
			AHref(router.MustURL(DRoute, "a", "a2", "b", "b2", "d", "d2.html"), "d2"),
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
	App       = HTML5(
		HTML(
			BODY(
				HEADER(menu),
				Router,
			),
		),
	)
)