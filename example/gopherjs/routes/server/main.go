package main

import (
	"github.com/go-on/method"
	"github.com/go-on/router"
	"github.com/go-on/router/example/gopherjs/routes"
	"net/http"
)

var Router = router.New()

func getArticle(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte("get article handler for " + req.FormValue(":id")))
}

func main() {
	Router.MustRegisterRoute(routes.GetArticle, method.GET, http.HandlerFunc(getArticle))
	router.MustMount(routes.AdminMountPoint, Router)
	http.ListenAndServe(":8086", nil)
}
