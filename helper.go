package router

import (
	"net/http"
	"net/url"
	"strings"

	"gopkg.in/go-on/wrap-contrib.v2/wraps"
	"gopkg.in/go-on/wrap.v2"
)

func GetRouteId(req *http.Request) (id string) {
	if i := strings.Index(req.URL.Fragment, "//0x"); i != -1 {
		return req.URL.Fragment[i:]
	}
	return
}

var slashB = []byte("/")[0]

func GetRouteParam(req *http.Request, key string) string {
	return GetURLParam(req.URL, key)
}

// since req.URL.Path has / unescaped so that originally escaped / are
// indistinguishable from escaped ones, we are save here, i.e. / is
// already handled as path splitted and no key or value has / in it
// also it is save to use req.URL.Fragment since that will never be transmitted
// by the request
func GetURLParam(u *url.URL, key string) (res string) {
	start, end := func() (start, end int) {
		var keyStart = 0
		var valStart = -1
		var inSlash bool
		for i := 0; i < len(u.Fragment); i++ {
			if u.Fragment[i] == slashB {
				if inSlash {
					break
				}
				inSlash = true
				if keyStart > -1 {
					if u.Fragment[keyStart:i] == key {
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
	return u.Fragment[start:end]
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
