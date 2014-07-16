package router

import (
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"regexp"

	"code.google.com/p/go-html-transform/html/transform"
	"github.com/go-on/router/route"
	"github.com/go-on/wrap-contrib/helper"
	// "net/http/httptest"
	"os"
	"path/filepath"
	"strings"
)

var staticRedirectTemplate = `<!DOCTYPE html>
<html>
		<head>
			<meta http-equiv="refresh" content="15; url=%s">
			<script language ="JavaScript">
			<!--
				document.location.href="%s";
			// -->
			</script>
		</head>
		<body>
		<p>Please click <a href="%s">here</a>, if you were not redirected automatically.</p>
		</body>
</html>
		`

// transformLink transforms relative links, that do not have a  fileextension and
// adds a .html to them
func transformLink(in string) (out string) {
	if in == "/" {
		return "index.html"
	}

	if !strings.HasPrefix(in, "/") {
		return in
	}

	if filepath.Ext(in) != "" {
		return in
	}

	return in + ".html"
}

func staticRedirect(location string) string {
	location = html.EscapeString(location)
	return fmt.Sprintf(staticRedirectTemplate, location, location, location)
}

var htmlContentType = regexp.MustCompile("html")

func savePath(server http.Handler, p, targetDir string) error {
	req, err := http.NewRequest("GET", p, nil)
	if req.Body != nil {
		defer req.Body.Close()
	}

	if err != nil {
		return err
	}

	// rec := httptest.NewRecorder()
	buf := helper.NewResponseBuffer(nil)

	server.ServeHTTP(buf, req)
	loc := buf.Header().Get("Location")
	if loc != "" {
		buf.Header().Set("Location", transformLink(loc))
	}

	contentType := buf.Header().Get("Content-Type")
	var body = buf.Buffer.String()

	if contentType == "" || htmlContentType.MatchString(contentType) {

		// if contentType

		x, err := transform.NewFromReader(&buf.Buffer)
		if err != nil {
			return err
		}
		err = x.Apply(transform.CopyAnd(transform.TransformAttrib("href", transformLink)), "a")

		if err != nil {
			return err
		}

		body = x.String()
	}

	/*
		buf.WriteHeadersTo(rw)
		buf.WriteCodeTo(rw)
		fmt.Fprint(rw, x.String())
		server.ServeHTTP(rec, req)
	*/

	switch buf.Code {
	case 301, 302:
		// fmt.Printf("got status %d for static file, location: %s\n", rec.Code, rec.Header().Get("Location"))
		body = staticRedirect(buf.Header().Get("Location"))
	case 200, 0:
		// body = buf.Buffer.Bytes()
	default:
		return fmt.Errorf("Status: %d, Body: %s", buf.Code, body)
	}

	/*
		if rec.Code != 200 {
			return fmt.Errorf("Status: %d, Body: %s", rec.Code, rec.Body.String())
		}
	*/

	if p != "" {
		p = transformLink(p)
	}

	path := filepath.Join(targetDir, p)
	if p[len(p)-1:] == "/" {
		path = filepath.Join(targetDir, p, "index.html")
	}

	os.MkdirAll(filepath.Dir(path), os.FileMode(0755))
	err = ioutil.WriteFile(path, []byte(body), os.FileMode(0644))

	if err != nil {
		fmt.Printf("can't write %s\n", body)
		return err
	}
	return nil
}

// DumpPaths calls the given paths on the given server and writes them to the target
// directory. The target directory must exist
func DumpPaths(server http.Handler, paths []string, targetDir string) (errors map[string]error) {
	errors = map[string]error{}

	d, e := os.Stat(targetDir)

	if e != nil {
		errors[""] = fmt.Errorf("%#v does not exist", targetDir)
		return
	}

	if !d.IsDir() {
		errors[""] = fmt.Errorf("%#v is no dir", targetDir)
		return
	}

	for _, p := range paths {
		// TODO maybe run savePath in goroutines that return an error channel
		// and collect all of them
		err := savePath(server, p, targetDir)

		if err != nil {
			errors[p] = err
		}
	}
	return
}

func (r *Router) EachRoute(fn func(mountPoint string, route *route.Route)) {
	for mP, rt := range r.routes {
		fn(mP, rt)
	}
}

func (r *Router) EachGETRoute(fn func(mountPoint string, route *route.Route)) {
	for mP, rt := range r.routes {
		if rt.GETHandler != nil {
			fn(mP, rt)
		}
	}
}

// the paths of all get routes
func (r *Router) AllGETPaths(paramSolver RouteParameter) (paths []string) {
	paths = []string{}
	fn := func(mountPoint string, rt *route.Route) {

		if rt.HasParams() {
			paramsArr := paramSolver.Params(rt)

			for _, params := range paramsArr {
				paths = append(paths, rt.MustURLMap(params))
			}

		} else {
			paths = append(paths, rt.MustURL())
		}
	}

	r.EachGETRoute(fn)
	return paths
}

// saves the results of all get routes
func (r *Router) SavePages(paramSolver RouteParameter, mainHandler http.Handler, targetDir string) map[string]error {
	return DumpPaths(mainHandler, r.AllGETPaths(paramSolver), targetDir)
}

func (r *Router) MustSavePages(paramSolver RouteParameter, mainHandler http.Handler, targetDir string) {
	errs := r.SavePages(paramSolver, mainHandler, targetDir)
	for _, err := range errs {
		panic(err.Error())
	}
}

// map[string][]interface{} is tag => []struct
func (r *Router) GETPathsByStruct(parameters map[*route.Route]map[string][]interface{}) (paths []string) {
	paths = []string{}

	fn := func(mountPoint string, route *route.Route) {
		paramPairs := parameters[route]

		// if route has : it has parameters
		if route.HasParams() {
			for tag, structs := range paramPairs {
				for _, stru := range structs {
					paths = append(paths, MustURLStruct(route, stru, tag))
				}
			}
		} else {
			paths = append(paths, route.MustURL())
		}
	}

	r.EachGETRoute(fn)
	return
}

func (r *Router) DynamicRoutes() (routes []*route.Route) {
	routes = []*route.Route{}
	for _, rt := range r.routes {
		if rt.HasParams() {
			routes = append(routes, rt)
		}
	}
	return routes
}

func (r *Router) StaticRoutePaths() (paths []string) {
	paths = []string{}
	for _, rt := range r.routes {
		if !rt.HasParams() {
			paths = append(paths, rt.MustURL())
		}
	}
	return paths
}
