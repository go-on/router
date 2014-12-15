package main

import (
	"go/build"
	"net/http"
	"path/filepath"

	"gopkg.in/go-on/wrap-contrib.v2/wraps"

	"gopkg.in/go-on/router.v2"
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
