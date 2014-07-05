package router

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-on/lib/internal/meta"
	"github.com/go-on/method"
	"github.com/go-on/router/route"
)

/*
type MountPather interface {
	MountPath() string
}
*/

/*
a Router should only know its subroutes and have a pointer to its parent router
*/

// TODO: have an inner Route that is decoupled from handler and router,
// that has only a mounter/pather that resolvs to a mount path and
// method.Methods (method package should get rid of net/http dependency)
// this inner Route should be used to define real routes at the first place.
// the mounter/pather should be used by the Router
/*
type Route struct {
	Handlers         map[method.Method]http.Handler
	RessourceOptions string
	MountedPath      string
	OriginalPath     string
	//router           *Router
	Router MountPather
}

func NewRoute(path string) *Route {
	// fmt.Println("creating route for path", path, "router", router.Path())
	r := &Route{}
	// r.Router = router
	r.MountedPath = path
	r.OriginalPath = path
	r.Handlers = map[method.Method]http.Handler{}
	return r
}

func (r *Route) SetHandler(m method.Method, h http.Handler) {
	r.Handlers[m] = h
}
*/

/*
func (r *Route) Router() *Router {
	return r.router
}
*/

/*
func Route(r *Route) string {
	return r.OriginalPath
}
*/

func HasParams(r *route.Route) bool {
	return strings.ContainsRune(r.OriginalPath, ':')
}

/*
func (r *Route) AddHandler(handler http.Handler, v method.Method) error {
	_, has := r.Handlers[v]
	if has {
		return fmt.Errorf("handler for method %s already defined", v)
	}
	r.Handlers[v] = handler
	return nil
}
*/

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

type OptionsServer struct {
	*route.Route
}

// serves the OPTIONS
func (r *OptionsServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if r.RessourceOptions == "" {
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
			for vb, _ := range r.Handlers {
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

		r.RessourceOptions = strings.Join(allow, ",")
	}
	rw.Header().Set("Allow", r.RessourceOptions)
}

func getHandler(r *route.Route, v string) http.Handler {
	v = strings.TrimSpace(strings.ToUpper(v))
	ver, ok := method.StringToMethod[v]
	if !ok {
		return nil
	}

	h, exists := r.Handlers[ver]
	if !exists {
		if ver == method.OPTIONS {
			return &OptionsServer{r}
		}

		for vb, h := range r.Handlers {
			if vb&ver != 0 {
				return h
			}
		}
		return nil
	}
	return h
}

// params are key/value pairs
func URL(ø *route.Route, params ...string) (string, error) {
	// fmt.Printf("params: %#v\n", params)
	if len(params)%2 != 0 {
		panic("number of params must be even (pairs of key, value)")
	}
	vars := map[string]string{}
	for i := 0; i < len(params); i += 2 {
		vars[params[i]] = params[i+1]
	}

	// fmt.Printf("%s => url vars: %#v\n", ø.path, vars)

	return URLMap(ø, vars)
}

// params are key/values
func URLMap(ø *route.Route, params map[string]string) (string, error) {
	segments := splitPath(ø.MountedPath)

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

	if ø.Router.MountPath() == "/" {
		return "/" + strings.Join(segments, "/"), nil
	} else {
		return ø.Router.MountPath() + "/" + strings.Join(segments, "/"), nil
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

	return URLMap(ø, params)
}

func MustURLMap(ø *route.Route, params map[string]string) string {
	u, err := URLMap(ø, params)
	if err != nil {
		panic(err.Error())
	}
	return u
}

func MustURLStruct(ø *route.Route, paramStruct interface{}, tagKey string) string {
	u, err := URLStruct(ø, paramStruct, tagKey)
	if err != nil {
		panic(err.Error())
	}
	return u
}

func MustURL(ø *route.Route, params ...string) string {
	u, err := URL(ø, params...)
	if err != nil {
		panic(err.Error())
	}
	return u
}
