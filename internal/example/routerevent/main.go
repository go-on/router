package main

import (
	"fmt"
	"net/http"

	event "gopkg.in/go-on/router.v2/internal/routerevent"
)

var UserGetMail = event.New("user.GetMail")

func EMailAddress(ev event.Event) {
	name := ev.Param().(string)
	ev.Return(name + "@example.com")
}

func init() {
	UserGetMail.Handle(EMailAddress)
}

func printEMail(rw http.ResponseWriter, req *http.Request) {
	ev := UserGetMail.Throw(req.URL.Query().Get("name"))
	email := ev.Returned()
	fmt.Fprintf(rw, "email is: %s", email.(string))

}

func main() {
	event.Watch()
	http.ListenAndServe(":8081", http.HandlerFunc(printEMail))
}
