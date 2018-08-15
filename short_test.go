package router

import (
	"testing"

	"github.com/go-on/method"
)

func TestGet(t *testing.T) {
	router := New()
	handlerFn := write("short").ServeHTTP
	router.GETFunc("/short", handlerFn)
	router.POSTFunc("/short", handlerFn)
	router.PUTFunc("/short", handlerFn)
	router.PATCHFunc("/short", handlerFn)
	router.DELETEFunc("/short", handlerFn)
	router.Mount("/", nil)

	corpus := map[method.Method]string{
		method.GET:    "GET short|",
		method.POST:   "POST short|",
		method.PUT:    "PUT short|",
		method.PATCH:  "PATCH short|",
		method.DELETE: "DELETE short|",
	}

	for meth, expected := range corpus {
		errMsg := assertResponse(meth, "/short", router, expected, 200)
		if errMsg != "" {
			t.Error(errMsg)
		}
	}
}
