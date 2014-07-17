package router

import (
	"go/build"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-on/method"
)

var gopath = build.Default.GOPATH

func findGoPathForPackage(pkg string) string {
	paths := filepath.SplitList(gopath)

	for _, path := range paths {
		p, err := os.Stat(filepath.Join(gopath, "src", pkg))
		if err == nil && p.IsDir() {
			return path
		}
	}
	return ""
}

var pkg = "github.com/go-on/router"

func TestFileServer(t *testing.T) {
	gpath := findGoPathForPackage(pkg)

	if gpath == "" {
		panic("cannot find GOPATH dir for package " + pkg)
	}

	dir := filepath.Join(gpath, "src", pkg, "internal", "_fileserver")

	router1 := New()
	fs1 := router1.FileServer("/files", dir)

	router1.Mount("/", nil)
	errMsg := assertResponse(method.GET, "/files/file.txt", router1, "filecontent", 200)

	if errMsg != "" {
		t.Error(errMsg)
	}

	errMsg = assertResponse(method.GET, "/files/not-there.txt", router1, "404 page not found", 404)

	if errMsg != "" {
		t.Error(errMsg)
	}

	router2 := New()
	fs2 := router2.FileServer("/files", dir)

	router2.Mount("/app", nil)
	errMsg = assertResponse(method.GET, "/app/files/file.txt", router2, "filecontent", 200)

	if errMsg != "" {
		t.Error(errMsg)
	}

	got := fs1.MustURL("file.txt")
	expected := "/files/file.txt"
	if got != expected {
		t.Errorf("wrong file url for fileserver fs1: %#v, expected: %#v", got, expected)
	}

	got = fs2.MustURL("file.txt")
	expected = "/app/files/file.txt"
	if got != expected {
		t.Errorf("wrong file url for fileserver fs2: %#v, expected: %#v", got, expected)
	}

	defer func() {
		e := recover()
		var err *os.PathError
		errMsg := errorMustBe(e, err)

		if errMsg != "" {
			t.Error(errMsg)
			return
		}
	}()

	fs2.MustURL("not-there.txt")
}
