package router

// FileServer serves the files from the given directory under the given path

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-on/router/route"
)

func (r *Router) FileServer(path, dir string) *FileServer {
	rt := r.newRoute(path)
	fs := &FileServer{
		fs:    http.FileServer(http.Dir(dir)),
		Dir:   dir,
		route: rt,
	}
	rt.GETHandler = fs
	return fs
}

type FileServer struct {
	fs    http.Handler
	Dir   string
	route *route.Route
	http.Handler
}

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
