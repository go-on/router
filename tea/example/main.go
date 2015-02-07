package main

import (
	"gopkg.in/go-on/lib.v3/html"
	"gopkg.in/go-on/router.v2"
	"gopkg.in/go-on/router.v2/route"
	"gopkg.in/go-on/router.v2/route/routehtml"
	"gopkg.in/go-on/router.v2/tea/t"
	"gopkg.in/go-on/wrap.v2"
	"gopkg.in/go-on/wrap-contrib.v2/wraps"
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
