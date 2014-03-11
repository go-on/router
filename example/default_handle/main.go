package main

import (
	"net/http"

	"github.com/go-on/router"
	"github.com/go-on/wrap-contrib/helper"
)

func main() {
	rt := router.New()

	rt.GET("/", helper.Write("root"))
	rt.GET("/hu", helper.Write("hu"))

	//http.Handle("/", rt)
	router.Mount("/", rt)

	err := http.ListenAndServe(":8080", rt)

	if err != nil {
		panic(err.Error())
	}

}
