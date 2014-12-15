package main

import (
	"gopkg.in/go-on/router.v2/internal/example/gopherjs/routes"
	"github.com/gopherjs/gopherjs/js"
)

func getElementById(id string) js.Object {
	return js.Global.Get("window").Get("document").Get("getElementById").Invoke(id)
}

func setInnerHTML(o js.Object, html string) {
	o.Set("innerHTML", html)
}

func setPath() {
	setInnerHTML(
		getElementById("content"),
		"created url via gopherjs: "+routes.Article.MustURL(routes.Id_, "23242"),
	)
}

func main() {
	js.Global.Get("jQuery").Invoke(setPath)
}
