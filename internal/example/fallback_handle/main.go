package main

import (
	"net/http"

	"github.com/go-on/wrap"
	"github.com/go-on/wrap-contrib/wraps"

	"github.com/go-on/router"
	// "github.com/go-on/wrap-contrib/helper"
)

func missing(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(404)
	rw.Write([]byte("not found"))
}

func main() {
	rt := router.New()

	rt.GET("/", wraps.String("root"))
	rt.GETFunc("/missing", missing)
	rt.GET("/hu", wraps.String("hu"))

	//http.Handle("/", rt)
	rt.Mount("/", nil)

	wrapper := wrap.New(
		wraps.Fallback(
			[]int{405}, // ignore 405 method not allowed status code
			rt,
			wraps.String("fallback"),
		),
	)

	err := http.ListenAndServe(":8087", wrapper)

	if err != nil {
		panic(err.Error())
	}

}
