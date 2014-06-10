package main

import (
	"github.com/go-on/wrap"
	"github.com/go-on/wrap-contrib/wraps"
	"net/http"

	"github.com/go-on/router"
	"github.com/go-on/wrap-contrib/helper"
)

func missing(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(404)
	rw.Write([]byte("not found"))
}

func main() {
	rt := router.New()

	rt.GET("/", helper.Write("root"))
	rt.GET("/missing", http.HandlerFunc(missing))
	rt.GET("/hu", helper.Write("hu"))

	//http.Handle("/", rt)
	router.Mount("/", rt)

	wrapper := wrap.New(
		wraps.Fallback(
			[]int{405}, // ignore 405 method not allowed status code
			rt,
			helper.Write("fallback"),
		),
	)

	err := http.ListenAndServe(":8087", wrapper)

	if err != nil {
		panic(err.Error())
	}

}
