package main

import (
	"github.com/go-on/lib/html"
	"github.com/go-on/router"
	"github.com/go-on/router/route"
	"github.com/go-on/router/route/routehtml"
	"github.com/go-on/router/tea/t"
	"gopkg.in/go-on/wrap-contrib.v2/wraps"
	"gopkg.in/go-on/wrap.v2"
)

var (
	paramName = "name"
	static    *router.FileServer
)

var helloRoute, errorRoute, betterRoute *route.Route

func main() {

	t.Use(
		Context{},
		wrap.NextHandlerFunc(start),
		wraps.HTMLContentType,
	)

	static = t.Static("/static", "./static")

	t.POSTFunc("/with-param/:"+paramName, wrap.NoOp)

	errorRoute = t.GET("/error", layout(routehtml.AHref(&helloRoute, nil, "should err"), html.PRE(stop)))

	helloRoute = t.GET("/with-param/:"+paramName, layout(html.H1(headingParam), html.PRE(stop)))

	betterRoute = t.GET("/no-params",
		layout(
			"this page has no parameters and links to ",
			routehtml.AHref(&helloRoute, routehtml.Params(paramName, "Peter"), "hello Peter"),
		),
	)

	t.Serve()
}
