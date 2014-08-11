package t

import (
	"fmt"
	"go/build"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	. "github.com/go-on/lib/html"
	"github.com/go-on/lib/html/internal/element"
	"github.com/go-on/lib/types"
	"github.com/go-on/wrap-contrib/wraps"
	"github.com/metakeule/backtrace"
)

var (
	gopathsrc     = filepath.Join(filepath.SplitList(os.Getenv("GOPATH"))[0], "src")
	goroot        = filepath.Join(build.Default.GOROOT, "src", "pkg")
	gorootDefault = "/usr/local/go/src/pkg"
)

func init() {
	fmt.Printf("GOROOT: %#v\n", goroot)
	fmt.Printf("GOPATH: %#v\n", gopathsrc)
}

func stripGoPath(s string) string {
	if strings.HasPrefix(s, gopathsrc) {
		return strings.Replace(s, gopathsrc, "$GOPATH", 1)
	}
	return s
}

func stripGoRoot(s string) string {
	if strings.HasPrefix(s, goroot) {
		return strings.Replace(s, goroot, "$GOROOT", 1)
	}
	return s
}

func stripGoRootDefault(s string) string {
	if strings.HasPrefix(s, gorootDefault) {
		return strings.Replace(s, gorootDefault, "$GOROOT", 1)
	}
	return s
}

var fileRegExp = regexp.MustCompile("(/[^:]+:[0-9]+)")

func urlForFile(s string) string {

	b := fileRegExp.ReplaceAllFunc([]byte(s), func(in []byte) []byte {

		stripped := stripGoRootDefault(stripGoRoot(stripGoPath(string(in))))
		path := "/_tea-launcheditor?file=" + url.QueryEscape(string(in))

		return []byte(
			AHref("#", stripped,
				OnClick_(fmt.Sprintf(`window.open(%#v); return false;`, path)),
			).String(),
		)

	})

	return string(b)
}

func mkTableForTrace(trace []backtrace.FootPrint) *element.Element {
	table := TABLE(THEAD(TR(TH("Function"), TH("File"))))
	tbody := TBODY()

	for _, tr := range trace {
		path := fmt.Sprintf("%s:%d", tr.File, tr.Line)
		tbody.Add(TR(TD(tr.Function), TD(types.HTMLString(urlForFile(path)))))
	}

	table.Add(tbody)
	return table
}

func launchEditor(rw http.ResponseWriter, req *http.Request) {

	wraps.HTMLContentType.SetContentType(rw)

	editor := os.Getenv("EDITOR")
	if editor == "" {
		rw.WriteHeader(http.StatusInternalServerError)
		layout("500 - can't launch editor", CODE("Can't launch editor, please set $EDITOR environment variable before running cot")).ServeHTTP(rw, req)
		return
	}
	file := req.URL.Query().Get("file")
	cmd := exec.Command(editor, file)
	out, err := cmd.CombinedOutput()
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		layout("500 - can't open file in editor", CODE(fmt.Sprintf("error while running %s %s:\n%s", editor, file, out))).ServeHTTP(rw, req)
		return
	}

	layout("Window closing", SCRIPT(types.JavaScriptString("window.close();"))).ServeHTTP(rw, req)
}

var ignoreFileSuffixes = []string{
	"runtime/proc.c",
	"runtime/panic.c",
	"wraps/catch.go",
	"tea/tea.go",
	"tea/backtrace.go",
	"tea/fallback.go",
	"tea/styles.go",
}

var ignoreFuncPrefixes = []string{
	"github.com/go-on",
	"net/http",
}

var Catcher func(recovered interface{}, rw http.ResponseWriter, req *http.Request) = defaultCatcher

func defaultCatcher(recovered interface{}, rw http.ResponseWriter, req *http.Request) {
	wraps.HTMLContentType.SetContentType(rw)

	btComplete := backtrace.BackTrace()

	btReduced := backtrace.Filter(btComplete, func(index int, fp backtrace.FootPrint) bool {
		for _, suff := range ignoreFileSuffixes {
			if strings.HasSuffix(fp.File, suff) {
				return false
			}
		}
		for _, pref := range ignoreFuncPrefixes {
			if strings.HasPrefix(fp.Function, pref) {
				return false
			}
		}
		return true
	})

	tblReduced, tblComplete := mkTableForTrace(btReduced), mkTableForTrace(btComplete)

	rw.WriteHeader(http.StatusInternalServerError)

	message := fmt.Sprintf("%s", recovered)

	if ty := fmt.Sprintf("%T", recovered); ty != "string" {
		message += " (" + ty + ")"
	}

	msg := types.HTMLString(urlForFile(message))

	layout(
		"500 Don't panic - your server already did",
		H1("500 Don't panic - your server already did it for you..."),
		P("Click on any file to open it in your editor."),
		CODE(msg), tblReduced,
		H2("Backtrace"), DIV(types.Class("complete-backtrace"), tblComplete),
	).ServeHTTP(rw, req)
}
