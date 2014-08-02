package main

import (
	"github.com/go-on/lib/html"
	"github.com/go-on/router"
	"github.com/go-on/router/route"
	"github.com/go-on/router/route/routehtml"
	"github.com/go-on/router/tea"
	"github.com/go-on/wrap"
	"github.com/go-on/wrap-contrib/wraps"
)

var (
	paramName = "name"
	static    *router.FileServer
)

var helloRoute, errorRoute, betterRoute *route.Route

func main() {

	tea.USE(
		Context{},
		wrap.NextHandlerFunc(start),
		wraps.HTMLContentType,
	)

	static = tea.STATIC("/static", "./static")

	tea.POSTFunc("/with-param/:"+paramName, wrap.NoOp)

	errorRoute = tea.GET("/error", layout(routehtml.AHref(&helloRoute, nil, "should err"), html.PRE(stop)))

	helloRoute = tea.GET("/with-param/:"+paramName, layout(html.H1(headingParam), html.PRE(stop)))

	betterRoute = tea.GET("/no-params", layout("this page has no parameters"))

	tea.SERVE()
}
