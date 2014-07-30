package main

import (
	"go/build"
	"net/http"
	"path/filepath"

	"github.com/go-on/wrap-contrib/wraps"

	"github.com/go-on/router"
)

var relPath = "src/github.com/go-on/router/example/fileserver/static"
var static = filepath.Join(filepath.SplitList(build.Default.GOPATH)[0], relPath)

func main() {
	rtr := router.New()
	fs := rtr.FileServer("/files", static)
	url := fs.MustURL("/hiho.txt")
	rtr.GET("/", wraps.TextString(url))

	http.ListenAndServe(":8084", rtr.ServingHandler())
}
