package routerrest

import (
	"net/http"

	"gopkg.in/go-on/method.v1"

	"github.com/go-on/router"
	"github.com/go-on/router/route"
)

type GetHandler interface {
	Get(http.ResponseWriter, *http.Request)
	ItemPath() string
}

type IndexHandler interface {
	Index(http.ResponseWriter, *http.Request)
	ListPath() string
}

type PostHandler interface {
	Post(http.ResponseWriter, *http.Request)
	ListPath() string
}

type DeleteHandler interface {
	Delete(http.ResponseWriter, *http.Request)
	ItemPath() string
}

type PatchHandler interface {
	Patch(http.ResponseWriter, *http.Request)
	ItemPath() string
}

func Register(rt *router.Router, handler interface{}) (m map[method.Method]*route.Route) {
	m = map[method.Method]*route.Route{}

	if get, ok := handler.(GetHandler); ok {
		m[method.GET] = rt.GET(get.ItemPath(), http.HandlerFunc(get.Get))
	}

	if post, ok := handler.(PostHandler); ok {
		m[method.POST] = rt.POST(post.ListPath(), http.HandlerFunc(post.Post))
	}

	if patch, ok := handler.(PatchHandler); ok {
		m[method.PATCH] = rt.PATCH(patch.ItemPath(), http.HandlerFunc(patch.Patch))
	}

	if del, ok := handler.(DeleteHandler); ok {
		m[method.DELETE] = rt.DELETE(del.ItemPath(), http.HandlerFunc(del.Delete))
	}

	if index, ok := handler.(IndexHandler); ok {
		m[method.Method(-1)] = rt.GET(index.ListPath(), http.HandlerFunc(index.Index))
	}
	return
}
