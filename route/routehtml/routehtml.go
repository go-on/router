// Package routehtml provides shortcuts to deal with route and go-on/lib/html
package routehtml

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/go-on/lib/html/internal/element"

	"github.com/go-on/lib/html"
	"github.com/go-on/router/route"
)

func Params(params ...string) (res map[string]string) {
	if len(params)%2 != 0 {
		panic(route.ErrPairParams{})
	}

	res = map[string]string{}
	for i := 0; i < len(params); i += 2 {
		res[params[i]] = params[i+1]
	}
	return
}

type Tag struct {
	rt     **route.Route
	params map[string]string
	data   []interface{}
	tag    string
	pc     uintptr
	file   string
	line   int
	ok     bool
}

func (t *Tag) Add(data ...interface{}) {
	t.data = append(t.data, data...)
}

func newTag(r **route.Route, params map[string]string, data []interface{}, tag string) *Tag {
	t := &Tag{rt: r, params: params, data: data, tag: tag}
	t.pc, t.file, t.line, t.ok = runtime.Caller(2)
	return t
}

func AHref(r **route.Route, params map[string]string, data ...interface{}) *Tag {
	return newTag(r, params, data, "a")
}

func FORM(r **route.Route, params map[string]string, data ...interface{}) *Tag {
	return newTag(r, params, data, "form")
}

func FormPost(r **route.Route, params map[string]string, data ...interface{}) *Tag {
	return newTag(r, params, data, "form-post")
}

func FormGet(r **route.Route, params map[string]string, data ...interface{}) *Tag {
	return newTag(r, params, data, "form-get")
}

func FormPut(r **route.Route, params map[string]string, data ...interface{}) *Tag {
	return newTag(r, params, data, "form-put")
}

func FormPatch(r **route.Route, params map[string]string, data ...interface{}) *Tag {
	return newTag(r, params, data, "form-patch")
}

func FormDelete(r **route.Route, params map[string]string, data ...interface{}) *Tag {
	return newTag(r, params, data, "form-delete")
}

func FormPostMultipart(r **route.Route, params map[string]string, data ...interface{}) *Tag {
	return newTag(r, params, data, "form-post-multipart")
}

func JsSrc(r **route.Route, params map[string]string, data ...interface{}) *Tag {
	return newTag(r, params, data, "js-src")
}

func CssHref(r **route.Route, params map[string]string, data ...interface{}) *Tag {
	return newTag(r, params, data, "css-href")
}

func ImgSrc(r **route.Route, params map[string]string, data ...interface{}) *Tag {
	return newTag(r, params, data, "img-src")
}

type ErrParams struct {
	message string
	*Tag
}

func (p *ErrParams) Error() string {
	return fmt.Sprintf("%s in %s:%d", p.message, p.Tag.file, p.Tag.line)
}

type ErrRouteIsNil struct {
	*Tag
}

func (p *ErrRouteIsNil) Error() string {
	return fmt.Sprintf("route is nil in %s:%d", p.Tag.file, p.Tag.line)
}

func (t *Tag) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if *t.rt == nil {
		panic(&ErrRouteIsNil{t})
	}
	url, err := (*t.rt).URLMap(t.params)
	if err != nil {
		panic(&ErrParams{err.Error(), t})
	}

	// url := (*t.rt).MustURLMap(t.params)
	var elem *element.Element
	switch t.tag {
	case "a":
		elem = html.AHref(url, t.data...)
	case "form":
		data := append(t.data, html.Action_(url))
		elem = html.FORM(data...)
	case "form-post":
		elem = html.FormPost(url, t.data...)
	case "form-get":
		elem = html.FormGet(url, t.data...)
	case "form-put":
		elem = html.FormPut(url, t.data...)
	case "form-patch":
		elem = html.FormPatch(url, t.data...)
	case "form-delete":
		elem = html.FormDelete(url, t.data...)
	case "form-post-multipart":
		elem = html.FormPostMultipart(url, t.data...)
	case "js-src":
		elem = html.JsSrc(url, t.data...)
	case "css-href":
		elem = html.CssHref(url, t.data...)
	case "img-src":
		elem = html.ImgSrc(url, t.data...)
	}

	elem.ServeHTTP(rw, req)
}
