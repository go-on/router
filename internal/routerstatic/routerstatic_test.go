package routerstatic

import (
	"testing"

	"github.com/go-on/router/internal/example/static/site"
	"github.com/go-on/router/route"
)

func TestTransformLink(t *testing.T) {
	corpus := map[string]string{
		transformLink("http://abc.de"): "http://abc.de",
		transformLink("/abc.de"):       "/abc.de",
		transformLink("/abc"):          "/abc.html",
		transformLink("/"):             "/index.html",
	}

	for got, expected := range corpus {
		if got != expected {
			t.Errorf("expected: %#v, got %#v", expected, got)
		}
	}
}

type resolver struct{}

func (rs resolver) Params(rt *route.Route) []map[string]string {
	switch rt {
	case site.DRoute:
		return []map[string]string{
			map[string]string{"a": "a0", "b": "b0", "d": "d0.html"},
			map[string]string{"a": "a1", "b": "b1", "d": "d1.html"},
			map[string]string{"a": "a2", "b": "b2", "d": "d2.html"},
		}
	default:
		panic("unhandled route: " + rt.DefinitionPath)
	}
}

func TestAllGetPaths(t *testing.T) {
	paths := AllGETPaths(site.Router, resolver{})

	if len(paths) != 6 {
		t.Errorf("must have 6 paths, but has %d", len(paths))
	}

	has := func(s string) bool {
		for _, p := range paths {
			if p == s {
				return true
			}
		}
		return false
	}

	checks := []string{
		"/",
		"/redirect",
		"/a.html",
		"/d/a0/x/b0/d0.html",
		"/d/a1/x/b1/d1.html",
		"/d/a2/x/b2/d2.html",
	}

	for _, c := range checks {
		if !has(c) {
			t.Errorf("missing path: %#v", c)
		}
	}

}

func TestRequestBody(t *testing.T) {
	site.Router.Mount("/", nil)
	tests := map[string]string{
		"/":                  "<!DOCTYPE html><html><head></head><body><header><ul><li><a href=\"/index.html\">Home</a></li><li><a href=\"/a.html\">a</a></li><li><a href=\"/d/a0/x/b0/d0.html\">d0</a></li><li><a href=\"/d/a1/x/b1/d1.html\">d1</a></li><li><a href=\"/d/a2/x/b2/d2.html\">d2</a></li></ul></header>index</body></html>",
		"/a.html":            "<!DOCTYPE html><html><head></head><body><header><ul><li><a href=\"/index.html\">Home</a></li><li><a href=\"/a.html\">a</a></li><li><a href=\"/d/a0/x/b0/d0.html\">d0</a></li><li><a href=\"/d/a1/x/b1/d1.html\">d1</a></li><li><a href=\"/d/a2/x/b2/d2.html\">d2</a></li></ul></header>A</body></html>",
		"/d/a0/x/b0/d0.html": "<!DOCTYPE html><html><head></head><body><header><ul><li><a href=\"/index.html\">Home</a></li><li><a href=\"/a.html\">a</a></li><li><a href=\"/d/a0/x/b0/d0.html\">d0</a></li><li><a href=\"/d/a1/x/b1/d1.html\">d1</a></li><li><a href=\"/d/a2/x/b2/d2.html\">d2</a></li></ul></header>a is &#34;a0&#34;, b is &#34;b0&#34;, d is &#34;d0.html&#34; whooho</body></html>",
		"/d/a1/x/b1/d1.html": "<!DOCTYPE html><html><head></head><body><header><ul><li><a href=\"/index.html\">Home</a></li><li><a href=\"/a.html\">a</a></li><li><a href=\"/d/a0/x/b0/d0.html\">d0</a></li><li><a href=\"/d/a1/x/b1/d1.html\">d1</a></li><li><a href=\"/d/a2/x/b2/d2.html\">d2</a></li></ul></header>a is &#34;a1&#34;, b is &#34;b1&#34;, d is &#34;d1.html&#34; whooho</body></html>",
		"/d/a2/x/b2/d2.html": "<!DOCTYPE html><html><head></head><body><header><ul><li><a href=\"/index.html\">Home</a></li><li><a href=\"/a.html\">a</a></li><li><a href=\"/d/a0/x/b0/d0.html\">d0</a></li><li><a href=\"/d/a1/x/b1/d1.html\">d1</a></li><li><a href=\"/d/a2/x/b2/d2.html\">d2</a></li></ul></header>a is &#34;a2&#34;, b is &#34;b2&#34;, d is &#34;d2.html&#34; whooho</body></html>",
		"/redirect":          "<!DOCTYPE html>\n<html>\n\t\t<head>\n\t\t\t<meta http-equiv=\"refresh\" content=\"15; url=/a.html\">\n\t\t\t<script language =\"JavaScript\">\n\t\t\t<!--\n\t\t\t\tdocument.location.href=\"/a.html\";\n\t\t\t// -->\n\t\t\t</script>\n\t\t</head>\n\t\t<body>\n\t\t<p>Please click <a href=\"/a.html\">here</a>, if you were not redirected automatically.</p>\n\t\t</body>\n</html>\n\t\t",
	}

	for path, expected := range tests {
		got, err := requestBody(site.App, path)

		if err != nil {
			t.Errorf("error: %s", err.Error())
		}
		if got != expected {
			t.Errorf("got: %#v expected: %#v", got, expected)
		}
	}

}
