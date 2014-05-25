package router

import (
	"code.google.com/p/go-html-transform/html/transform"
	"fmt"
	"github.com/go-on/wrap-contrib/helper"
	"html"
	"io/ioutil"
	"net/http"
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

// transformLink transforms relative links, that do not have a  fileextension and
// adds a .html to them
func transformLink(in string) (out string) {
	if in == "/" {
		return "index.html"
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

func savePath(server http.Handler, p, targetDir string) error {
	req, err := http.NewRequest("GET", p, nil)
	if req.Body != nil {
		defer req.Body.Close()
	}

	if err != nil {
		return err
	}

	// rec := httptest.NewRecorder()
	buf := helper.NewResponseBuffer(nil)

	server.ServeHTTP(buf, req)
	loc := buf.Header().Get("Location")
	if loc != "" {
		buf.Header().Set("Location", transformLink(loc))
	}

	x, err := transform.NewFromReader(&buf.Buffer)
	if err != nil {
		return err
	}
	err = x.Apply(transform.CopyAnd(transform.TransformAttrib("href", transformLink)), "a")

	if err != nil {
		return err
	}

	var body = x.String()

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
		return fmt.Errorf("Status: %d, Body: %s", buf.Code, body)
	}

	/*
		if rec.Code != 200 {
			return fmt.Errorf("Status: %d, Body: %s", rec.Code, rec.Body.String())
		}
	*/

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
