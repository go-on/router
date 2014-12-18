package main

import (
	"fmt"
	"net/http"

	"gopkg.in/go-on/method.v1"
	"gopkg.in/go-on/router.v2"
	"gopkg.in/go-on/router.v2/route"
)

var (
	Car        = route.New("/car/:car_id", method.GET, method.POST, method.PUT)
	Cars       = route.New("/car", method.GET)
	MountPoint = "/api/v1"
)

func init() {
	route.Mount(MountPoint, Car, Cars)
}

var Router = router.New()

func carHandler(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "car: %s, method: %s", router.GetRouteParam(req, "car_id"), req.Method)
}

func postCarHandler(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "car: %s, method: POST", router.GetRouteParam(req, "car_id"))
}

func carsHandler(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "all cars")
}

func main() {
	Router.HandleRouteFunc(Car, carHandler, method.GET, method.PUT)
	Router.HandleRouteFunc(Car, postCarHandler, method.POST)
	Router.HandleRouteFunc(Cars, carsHandler, method.GET)
	Router.Mount(MountPoint, nil)

	http.ListenAndServe(":8080", Router.ServingHandler())
}
