package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"github.com/go-on/router/example/static/site"

	"github.com/go-on/router"
)

type resolver struct{}

func (rs resolver) Params(rt *router.Route) []map[string]string {
	switch rt {
	case site.DRoute:
		return []map[string]string{
			map[string]string{"a": "a0", "b": "b0", "d": "d0.html"},
			map[string]string{"a": "a1", "b": "b1", "d": "d1.html"},
			map[string]string{"a": "a2", "b": "b2", "d": "d2.html"},
		}
	default:
		panic("unhandled route: " + rt.Route())
	}
}

func main() {
	router.Mount("/", site.Router)

	gopath := os.Getenv("GOPATH")
	dir := filepath.Join(gopath, "src", "github.com", "go-on", "router", "example", "static", "result")

	os.RemoveAll(dir)
	os.Mkdir(dir, os.FileMode(0755))

	fmt.Println("dump paths")
	site.Router.MustSavePages(resolver{}, site.App, dir)

	fmt.Println("running static fileserver at localhost:8080")

	err := http.ListenAndServe(":8080", http.FileServer(http.Dir(dir)))

	if err != nil {
		panic(err.Error())
	}
}
