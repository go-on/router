package main

import (
	"fmt"
	"net/http"

	. "github.com/go-on/lib/html"
	"github.com/go-on/lib/html/element"
	. "github.com/go-on/lib/types"
	"github.com/go-on/router"
	"github.com/go-on/router/route/routehtml"
	"github.com/go-on/router/tea/t"
)

var noDecoration = Style{"text-decoration", "none"}

func menu(w http.ResponseWriter, r *http.Request) {
	m := UL()

	entries := []*element.Element{
		LI(routehtml.AHref(&betterRoute, nil, "without params")),
		LI(routehtml.AHref(&helloRoute, routehtml.Params(paramName, "<world>"), "with params")),
		LI(routehtml.AHref(&errorRoute, nil, "with error")),
		LI(AHref("/", "no route")),
	}

	var no = -1
	switch router.GetRouteId(r) {
	case betterRoute.Id:
		no = 0
	case helloRoute.Id:
		no = 1
	case errorRoute.Id:
		no = 2
	}

	if no != -1 {
		entries[no].Add(Class("active"))
	}

	for _, e := range entries {
		m.Add(e)
	}
	m.ServeHTTP(w, r)
}

func layout(body ...interface{}) http.Handler {
	return HTML5(
		HTML(
			HEAD(CssHref(static.MustURL("styles.css"), Media_("screen"))),
			BODY(
				NAV(menu),
				DIV(Class("main"), DIV(body...)),
			),
		),
	)
}

func headingParam(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello! The parameter is: %s", EscapeHTML(t.RouteParam(r, paramName)))
}
