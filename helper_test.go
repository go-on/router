package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/go-on/method"
)

type writeParam struct {
	text   string
	params []string
}

func (w writeParam) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "%s %s|", req.Method, w.text)
	for _, param := range w.params {
		fmt.Fprintf(rw, "%s:%s,", param, GetRouteParam(req, param))
	}
}

func writeParams(text string, params ...string) writeParam {
	return writeParam{text, params}
}

func write(text string) writeParam {
	return writeParam{text: text}
}

type writeString string

func (ww writeString) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, string(ww))
}

// Make a testing request
func newTestRequest(method, path string) (*httptest.ResponseRecorder, *http.Request) {
	request, err := http.NewRequest(method, path, nil)
	if err != nil {
		fmt.Printf("could not make request %s (%s): %s\n", path, method, err.Error())
	}
	recorder := httptest.NewRecorder()

	return recorder, request
}

func assertResponse(meth method.Method, path string, handler http.Handler, expectedbody string, code int) string {
	rec, req := newTestRequest(meth.String(), path)
	handler.ServeHTTP(rec, req)

	body := strings.TrimSpace(rec.Body.String())
	if body != expectedbody {
		return fmt.Sprintf("wrong body in %s %s, expected: %#v, got %#v", meth, path, expectedbody, body)
	}

	if rec.Code != code {
		return fmt.Sprintf("wrong status code in %s %s, expected: %d, got %d", meth, path, rec.Code, code)
	}

	return ""
}

func assertResponseHeader(meth method.Method, path string, handler http.Handler, headerKey string, expectedHeaderVal string) string {
	rec, req := newTestRequest(meth.String(), path)
	handler.ServeHTTP(rec, req)

	got := rec.Header().Get(headerKey)
	if got != expectedHeaderVal {
		return fmt.Sprintf("wrong header %s in %s %s, expected: %#v, got %#v", headerKey, meth, path, expectedHeaderVal, got)
	}
	return ""
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	w.Write([]byte(`not found`))
}

func mount(r *Router, mountpoint string) *Router {
	r.Mount(mountpoint, http.NewServeMux())
	return r
}
