package router

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-on/method"
	"github.com/go-on/router/route"
)

// FileServer serves the files from the given directory under the given path
func (r *Router) FileServer(path, dir string) *FileServer {
	rt := r.newRouteHandler(path, method.GET)
	fs := NewFileServer(http.FileServer(http.Dir(dir)), dir, rt.Route)
	rt.GETHandler = fs
	return fs
}

func NewFileServer(fs http.Handler, dir string, route *route.Route) *FileServer {
	return &FileServer{
		fs:    fs,
		Dir:   dir,
		route: route,
	}
}

type FileServer struct {
	fs    http.Handler
	Dir   string
	route *route.Route
	http.Handler
}

// TODO: rename if to mount
func (fs *FileServer) SetHandler() {
	fs.Handler = http.StripPrefix(fs.route.MountedPath(), fs.fs)
}

func (fs *FileServer) URL(relativePath string) (string, error) {
	_, err := os.Stat(filepath.Join(fs.Dir, relativePath))
	if err != nil {
		return "", err
	}
	return filepath.Join(fs.route.MountedPath(), relativePath), nil
}

func (fs *FileServer) MustURL(relativePath string) string {
	url, err := fs.URL(relativePath)
	if err != nil {
		panic(err)
	}
	return url
}
