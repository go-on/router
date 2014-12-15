package t

import (
	"fmt"
	"net/http"

	. "github.com/go-on/lib/html"
	"github.com/go-on/lib/types"
	"github.com/go-on/method"
	"github.com/go-on/router/route"
	"gopkg.in/go-on/wrap-contrib.v2/wraps"
)

var createCode = `t.GETFunc(%#v, func (w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("new route for GET %v"))
})`

func listOfRoutes() http.Handler {
	routesDefined := UL(types.Class("routes-defined"))

	nonFileServer.EachRoute(func(mountpath string, rt *route.Route) {
		for m := range rt.Methods {
			if m == method.GET {
				routesDefined.Add(LI(AHref(rt.MountedPath(), fmt.Sprintf("%s %s", m, rt.MountedPath()))))
			} else {
				routesDefined.Add(LI(fmt.Sprintf("%s %s", m, rt.MountedPath())))
			}
		}
	})
	return routesDefined
}

var Fallback func(rw http.ResponseWriter, req *http.Request) = defaultFallback

func defaultFallback(rw http.ResponseWriter, req *http.Request) {
	wraps.HTMLContentType.SetContentType(rw)

	rw.WriteHeader(http.StatusMethodNotAllowed)

	if req.Method != method.GET.String() {
		return
	}

	layout(
		"405 This route is not defined yet",
		H1("405 This route is not defined yet..."), "To create it, add the following code",
		CODE(PRE(fmt.Sprintf(createCode, req.URL.Path, req.URL.Path))),
		H2("Already there"), listOfRoutes(),
	).ServeHTTP(rw, req)

}

func teapot(rw http.ResponseWriter, req *http.Request) {
	wraps.HTMLContentType.SetContentType(rw)
	rw.WriteHeader(http.StatusTeapot)
	layout(
		"418 I am not a coffee pot",
		H1("418 Tea is ready - how about the pot?"),
		P("HTCPCP/1.0 was not meant to be implemented by tea. So maybe you switch?"),
		ImgSrc("http://www.htcpcp.net/img/Error%20418%20htcpcp%20teapot_R1.jpg", Width_("450px")),
		DIV("Image from ", AHref("http://www.htcpcp.net", "http://www.htcpcp.net", TargetBlank_)),
	).ServeHTTP(rw, req)
}
