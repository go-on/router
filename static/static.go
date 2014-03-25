package static

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/go-on/router"
)

// map[string][]interface{} is tag => []struct
func PathsByStruct(r *router.Router, parameters map[*router.Route]map[string][]interface{}) (paths []string) {
	paths = []string{}

	fn := func(mountPoint string, route *router.Route) {
		paramPairs := parameters[route]

		// if route has : it has parameters
		if r.HasParams() {
			for tag, structs := range paramPairs {
				for _, stru := range structs {
					paths = append(paths, route.MustURLStruct(stru, tag))
				}
			}
		} else {
			paths = append(paths, route.MustURL())
		}
	}

	r.EachRoute(fn)
	return
}

func savePath(server http.Handler, p, targetDir string) error {
	req, err := http.NewRequest("GET", p, nil)

	if err != nil {
		return err
	}

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != 200 {
		return fmt.Errorf("Status: %d, Body: %s", rec.Code, rec.Body.String())
	}

	path := filepath.Join(targetDir, p)
	if p[len(p)-1:] == "/" {
		path = filepath.Join(targetDir, p, "index.html")
	}

	os.MkdirAll(filepath.Dir(path), os.FileMode(0755))
	err = ioutil.WriteFile(path, rec.Body.Bytes(), os.FileMode(0644))

	if err != nil {
		fmt.Printf("can't write %#v\n", rec.Body.String())
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
