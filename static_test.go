package router

import (
	"os"
	"path/filepath"

	. "launchpad.net/gocheck"
)

type staticSuite struct{}

var _ = Suite(&staticSuite{})

func containsString(a []string, s string) bool {
	for _, as := range a {
		if as == s {
			return true
		}
	}
	return false
}

func (s *staticSuite) TestRouting(c *C) {
	router := mount(makeRouter(), "/")
	paramSolver := RouteParameterFunc(func(r *Route) []map[string]string {
		if r.Route() == "/:sth/x.html" {
			return []map[string]string{
				map[string]string{
					"sth": "hu",
				},
			}
		}

		return nil
	})

	paths := router.AllGETPaths(paramSolver)

	c.Assert(len(paths), Equals, 6)

	c.Assert(containsString(paths, "/a.html"), Equals, true)
	c.Assert(containsString(paths, "/b.html"), Equals, true)
	c.Assert(containsString(paths, "/a/x.html"), Equals, true)
	c.Assert(containsString(paths, "/a/b.html"), Equals, true)
	c.Assert(containsString(paths, "/b/x.html"), Equals, true)
	c.Assert(containsString(paths, "/hu/x.html"), Equals, true)

	tmpDir := filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "go-on", "router", "temp")

	os.RemoveAll(tmpDir)

	os.MkdirAll(tmpDir, 0644)

	// errs :=
	// DumpPaths(router, paths, tmpDir)

	// c.Assert(len(errs), Equals, 0)

}
