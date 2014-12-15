package main

import (
	"net/http"

	"gopkg.in/go-on/method.v1"
	"gopkg.in/go-on/router.v2"
	"gopkg.in/go-on/router.v2/internal/example/gopherjs/routes"
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
