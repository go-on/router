package main

import (
	"net/http"

	"gopkg.in/go-on/wrap.v2"
	"gopkg.in/go-on/wrap-contrib.v2/wraps"

	"gopkg.in/go-on/router.v2"
	// "gopkg.in/go-on/wrap-contrib.v2/helper"
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
