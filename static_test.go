package router

import (
	"os"
	"path/filepath"

	. "launchpad.net/gocheck"
)

type staticSuite struct{}

var _ = Suite(&staticSuite{})

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
	c.Assert(paths[0], Equals, "/a.html")
	c.Assert(paths[1], Equals, "/b.html")
	c.Assert(paths[2], Equals, "/a/x.html")
	c.Assert(paths[3], Equals, "/a/b.html")
	c.Assert(paths[4], Equals, "/b/x.html")
	c.Assert(paths[5], Equals, "/hu/x.html")

	tmpDir := filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "go-on", "router", "temp")

	os.RemoveAll(tmpDir)

	os.MkdirAll(tmpDir, 0644)

	// errs :=
	// DumpPaths(router, paths, tmpDir)

	// c.Assert(len(errs), Equals, 0)

}
