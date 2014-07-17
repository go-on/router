package main

import (
	"net/http"

	"github.com/go-on/router"
	"github.com/go-on/router/internal/example/gopherjs/routes"
)

var Router = router.New()

func getArticle(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte("get article handler for " + req.FormValue(":id")))
}

func main() {
	// method.GET, http.HandlerFunc(getArticle)
	routes.GetArticle.GETHandler = http.HandlerFunc(getArticle)
	Router.MustAddRoute(routes.GetArticle)
	Router.Mount(routes.AdminMountPoint, nil)
	http.ListenAndServe(":8086", Router.ServingHandler())
}
