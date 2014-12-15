package router

import (
	"net/http"
	"strings"

	"gopkg.in/go-on/wrap.v2"
	"gopkg.in/go-on/wrap-contrib.v2/wraps"
)

func GetRouteId(req *http.Request) (id string) {
	if i := strings.Index(req.URL.Fragment, "//0x"); i != -1 {
		return req.URL.Fragment[i:]
	}
	return
}

var slashB = []byte("/")[0]

// since req.URL.Path has / unescaped so that originally escaped / are
// indistinguishable from escaped ones, we are save here, i.e. / is
// already handled as path splitted and no key or value has / in it
// also it is save to use req.URL.Fragment since that will never be transmitted
// by the request
func GetRouteParam(req *http.Request, key string) (res string) {
	start, end := func() (start, end int) {
		var keyStart = 0
		var valStart = -1
		var inSlash bool
		for i := 0; i < len(req.URL.Fragment); i++ {
			if req.URL.Fragment[i] == slashB {
				if inSlash {
					break
				}
				inSlash = true
				if keyStart > -1 {
					if req.URL.Fragment[keyStart:i] == key {
						valStart = i + 1
					}
					keyStart = -1
					continue
				}

				keyStart = i + 1

				if valStart > -1 {
					return valStart, i
				}
				continue
			}
			inSlash = false
		}
		return -1, -1
	}()

	if start == -1 {
		return
	}
	return req.URL.Fragment[start:end]
}

func NewETagged(wrappers ...wrap.Wrapper) (r *Router) {
	r = newRouter()
	wrappers = append(
		wrappers,
		wraps.IfNoneMatch,
		wraps.IfMatch(r),
		wraps.ETag,
	)
	r.wrapper = append(r.wrapper, wrappers...)
	return
}
