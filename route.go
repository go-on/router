package router

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-on/meta"
	"github.com/go-on/method"
)

/*
a Router should only know its subroutes and have a pointer to its parent router
*/

type Route struct {
	handler          map[method.Method]http.Handler
	ressourceOptions string
	path             string
	router           *Router
}

func newRoute(router *Router, path string) *Route {
	r := &Route{}
	r.router = router
	r.path = path
	r.handler = map[method.Method]http.Handler{}
	return r
}

func (r *Route) Router() *Router {
	return r.router
}

func (r *Route) addHandler(handler http.Handler, v method.Method) error {
	_, has := r.handler[v]
	if has {
		return fmt.Errorf("handler for method %s already defined", v)
	}
	r.handler[v] = handler
	return nil
}

/*
func (r *Route) inspect(indent int) string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%s", strings.Repeat("\t", indent))
	for v, _ := range r.handler {
		fmt.Fprintf(&buf, "%s ", v)
	}
	return buf.String()
}
*/

// serves the OPTIONS
func (r *Route) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if r.ressourceOptions == "" {
		opts := map[method.Method]bool{method.OPTIONS: true}
		allow := []string{}
		all := []method.Method{
			method.GET,
			method.POST,
			method.DELETE,
			// method.HEAD,
			method.PATCH,
			method.PUT,
		}

		for _, m := range all {
			for vb, _ := range r.handler {
				if vb&m != 0 {
					opts[m] = true
				}
			}
		}

		if opts[method.GET] {
			opts[method.HEAD] = true
		}

		for m, ok := range opts {
			if ok {
				allow = append(allow, m.String())
			}
		}

		r.ressourceOptions = strings.Join(allow, ",")
	}
	rw.Header().Set("Allow", r.ressourceOptions)
}

func (r *Route) getHandler(v string) http.Handler {
	v = strings.TrimSpace(strings.ToUpper(v))
	ver, ok := method.StringToMethod[v]
	if !ok {
		return nil
	}

	if ver == method.OPTIONS {
		return r
	}

	h, exists := r.handler[ver]
	if !exists {
		for vb, h := range r.handler {
			if vb&ver != 0 {
				return h
			}
		}
		return nil
	}
	return h
}

// params are key/value pairs
func (ø *Route) URL(params ...string) (string, error) {
	if len(params)%2 != 0 {
		panic("number of params must be even (pairs of key, value)")
	}
	vars := map[string]string{}
	for i := 0; i < len(params)/2; i += 2 {
		vars[params[i]] = params[i+1]
	}

	return ø.URLMap(vars)
}

// params are key/values
func (ø *Route) URLMap(params map[string]string) (string, error) {
	segments := splitPath(ø.path)

	for i := range segments {
		wc, wcName := isWildcard(segments[i])
		if wc {
			repl, ok := params[wcName]
			if !ok {
				return "", fmt.Errorf("missing parameter for %s", wcName)
			}
			segments[i] = repl
		}
	}

	if ø.router.Path() == "/" {
		return "/" + strings.Join(segments, "/"), nil
	} else {
		return ø.router.Path() + "/" + strings.Join(segments, "/"), nil
	}
}

var strTy = reflect.TypeOf("")

func (ø *Route) URLStruct(paramStruct interface{}, tagKey string) (string, error) {
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

func (ø *Route) MustURLMap(params map[string]string) string {
	u, err := ø.URLMap(params)
	if err != nil {
		panic(err.Error())
	}
	return u
}

func (ø *Route) MustURLStruct(paramStruct interface{}, tagKey string) string {
	u, err := ø.URLStruct(paramStruct, tagKey)
	if err != nil {
		panic(err.Error())
	}
	return u
}

func (ø *Route) MustURL(params ...string) string {
	u, err := ø.URL(params...)
	if err != nil {
		panic(err.Error())
	}
	return u
}
