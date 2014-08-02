package routerstatic

import (
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"

	"github.com/go-on/method"

	"github.com/go-on/wrap"

	"code.google.com/p/go-html-transform/html/transform"
	"github.com/go-on/lib/internal/meta"
	"github.com/go-on/router"
	"github.com/go-on/router/route"

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

type ParameterFunc func(*route.Route) []map[string]string

func (rpf ParameterFunc) Params(rt *route.Route) []map[string]string {
	return rpf(rt)
}

type Parameter interface {
	Params(*route.Route) []map[string]string
}

// transformLink transforms relative links, that do not have a  fileextension and
// adds a .html to them
func transformLink(in string) (out string) {
	if in == "/" {
		return "/index.html"
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

func requestBody(server http.Handler, p string) (string, error) {
	req, err := http.NewRequest("GET", p, nil)
	if req.Body != nil {
		defer req.Body.Close()
	}

	if err != nil {
		return "", err
	}

	// rec := httptest.NewRecorder()
	buf := wrap.NewBuffer(nil)

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
			return "", err
		}
		err = x.Apply(transform.CopyAnd(transform.TransformAttrib("href", transformLink)), "a")

		if err != nil {
			return "", err
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
		return "", fmt.Errorf("Status: %d, Body: %s", buf.Code, body)
	}
	return body, nil
	/*
		if rec.Code != 200 {
			return fmt.Errorf("Status: %d, Body: %s", rec.Code, rec.Body.String())
		}
	*/
}

func savePath(server http.Handler, p, targetDir string) error {
	body, err := requestBody(server, p)
	if err != nil {
		return err
	}

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

// the paths of all get routes
func AllGETPaths(r *router.Router, paramSolver Parameter) (paths []string) {
	paths = []string{}
	fn := func(mountPoint string, rt *route.Route) {
		if rt.HasMethod(method.GET) {
			if rt.HasParams() {
				paramsArr := paramSolver.Params(rt)

				for _, params := range paramsArr {
					paths = append(paths, rt.MustURLMap(params))
				}

			} else {
				paths = append(paths, rt.MustURL())
			}
		}
	}

	r.EachRoute(fn)
	return paths
}

// saves the results of all get routes
func SavePages(r *router.Router, paramSolver Parameter, mainHandler http.Handler, targetDir string) map[string]error {
	return DumpPaths(mainHandler, AllGETPaths(r, paramSolver), targetDir)
}

func MustSavePages(r *router.Router, paramSolver Parameter, mainHandler http.Handler, targetDir string) {
	errs := SavePages(r, paramSolver, mainHandler, targetDir)
	for _, err := range errs {
		panic(err.Error())
	}
}

var strTy = reflect.TypeOf("")

func URLStruct(ø *route.Route, paramStruct interface{}, tagKey string) (string, error) {
	val := reflect.ValueOf(paramStruct)
	params := map[string]string{}
	stru, err := meta.StructByValue(val)
	if err != nil {
		return "", err
	}

	fn := func(field *meta.Field, tagVal string) {
		params[tagVal] = field.Value.Convert(strTy).String()
	}

	stru.EachTag(tagKey, fn)

	return ø.URLMap(params)
}

func MustURLStruct(ø *route.Route, paramStruct interface{}, tagKey string) string {
	u, err := URLStruct(ø, paramStruct, tagKey)
	if err != nil {
		panic(err.Error())
	}
	return u
}

// map[string][]interface{} is tag => []struct
func GETPathsByStruct(r *router.Router, parameters map[*route.Route]map[string][]interface{}) (paths []string) {
	paths = []string{}

	fn := func(mountPoint string, route *route.Route) {
		if route.HasMethod(method.GET) {
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
	}

	r.EachRoute(fn)
	return
}

func DynamicRoutes(r *router.Router) (routes []*route.Route) {
	routes = []*route.Route{}
	r.EachRoute(func(s string, rt *route.Route) {
		if rt.HasParams() {
			routes = append(routes, rt)
		}
	})
	return routes
}

func StaticRoutePaths(r *router.Router) (paths []string) {
	paths = []string{}
	r.EachRoute(func(s string, rt *route.Route) {
		if !rt.HasParams() {
			paths = append(paths, rt.MustURL())
		}
	})
	return paths
}
