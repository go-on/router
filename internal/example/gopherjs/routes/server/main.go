package main

import (
	"net/http"

	"github.com/go-on/method"
	"github.com/go-on/router"
	"github.com/go-on/router/internal/example/gopherjs/routes"
)

var Router = router.New()

func getArticle(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte(
		"get article handler for " +
			router.GetRouteParam(req, routes.Id_),
	))
}

func main() {
	Router.HandleRouteMethodsFunc(routes.Article, getArticle, method.GET)
	Router.Mount(routes.ADMIN, nil)
	http.ListenAndServe(":8086", Router.ServingHandler())
}
