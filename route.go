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

func HasParams(r *route.Route) bool {
	return strings.ContainsRune(r.DefinitionPath, ':')
}

func optionsString(r *route.Route) string {
	allow := []string{method.OPTIONS.String()}

	if r.GETHandler != nil {
		allow = append(allow, method.GET.String())
		allow = append(allow, method.HEAD.String())
	}

	if r.POSTHandler != nil {
		allow = append(allow, method.POST.String())
	}

	if r.DELETEHandler != nil {
		allow = append(allow, method.DELETE.String())
	}

	if r.PATCHHandler != nil {
		allow = append(allow, method.PATCH.String())
	}

	if r.PUTHandler != nil {
		allow = append(allow, method.PUT.String())
	}

	return strings.Join(allow, ",")
}

func SetOPTIONSHandler(r *route.Route) {
	options := optionsString(r)
	r.OPTIONSHandler = http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Allow", options)
	})

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
	// println("MountedPath: ", ø.MountedPath)
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
