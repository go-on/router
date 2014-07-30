package tea

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/cupcake/mannersagain"
	"github.com/go-on/router"
	"github.com/go-on/router/route"
	"github.com/go-on/wrap"
	"github.com/go-on/wrap-contrib/wraps"
)

var DEVELOPMENT = true

var (
	Router        = router.New()
	nonFileServer = router.NewETagged()
	middlewares   = []wrap.Wrapper{}
)

func USE(middleware ...wrap.Wrapper) {
	middlewares = append(middlewares, middleware...)
}

func ROUTE_PARAM(req *http.Request, name string) string {
	return router.GetRouteParam(req, name)
}

func STATIC(path, directory string) *router.FileServer {
	return Router.FileServer(path, directory)
}

func GET(path string, fn http.Handler) *route.Route {
	return nonFileServer.GET(path, fn)
}

func POST(path string, fn http.Handler) *route.Route {
	return nonFileServer.POST(path, fn)
}

func PUT(path string, fn http.Handler) *route.Route {
	return nonFileServer.PUT(path, fn)
}

func PATCH(path string, fn http.Handler) *route.Route {
	return nonFileServer.PATCH(path, fn)
}

func DELETE(path string, fn http.Handler) *route.Route {
	return nonFileServer.DELETE(path, fn)
}

func GETFunc(path string, fn http.HandlerFunc) *route.Route {
	return nonFileServer.GETFunc(path, fn)
}

func POSTFunc(path string, fn http.HandlerFunc) *route.Route {
	return nonFileServer.POSTFunc(path, fn)
}

func PUTFunc(path string, fn http.HandlerFunc) *route.Route {
	return nonFileServer.PUTFunc(path, fn)
}

func PATCHFunc(path string, fn http.HandlerFunc) *route.Route {
	return nonFileServer.PATCHFunc(path, fn)
}

func DELETEFunc(path string, fn http.HandlerFunc) *route.Route {
	return nonFileServer.DELETEFunc(path, fn)
}

func mkHandler() http.Handler {
	if DEVELOPMENT {
		Router.GETFunc("/_tea-launcheditor", launchEditor)
		Router.GETFunc("/_tea-418", teapot)
	}
	nonFileServer.Mount("/", nil)
	nonFsStack := []wrap.Wrapper{}
	nonFsStack = append(nonFsStack, middlewares...)
	nonFsStack = append(nonFsStack, wrap.Handler(nonFileServer))

	mw := []wrap.Wrapper{}
	mw = append(mw,
		wraps.CatchFunc(CATCHER),
		wraps.Fallback(
			[]int{http.StatusMethodNotAllowed},
			wrap.New(nonFsStack...),
			Router.ServingHandler(),
		),
		wrap.HandlerFunc(FALLBACK),
	)
	return wrap.New(mw...)
}

func SERVE() {
	pid := os.Getpid()
	wd, err := os.Getwd()
	if err != nil {
		panic("can't get working directory " + err.Error())
	}

	ioutil.WriteFile(filepath.Join(wd, "main.pid"), []byte(fmt.Sprintf("%d", pid)), 0644)
	handler := mkHandler()
	port := 8080
	for i := port; i < port+10; i++ {
		err := mannersagain.ListenAndServe(fmt.Sprintf(":%d", i), handler)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stdout, err)
		}
		return
	}
}

func SERVE_ADDRESS(address string) {
	handler := mkHandler()
	err := mannersagain.ListenAndServe(address, handler)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
	return
}
