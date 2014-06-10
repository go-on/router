package main

import (
	. "github.com/go-on/html"
	. "github.com/go-on/html/tag"
	"github.com/go-on/router"
	"github.com/go-on/template"
	"net/http"
)

var (
	header = Text("header").Placeholder()
	body   = Html("body").Placeholder()
	layout = HTML5(
		HEAD(
			TITLE(header),
		),
		BODY(
			NAV(
				A(Attr("href", "/"), "navigate to /"), BR(),
				A(Attr("href", "/app"), "navigate to /app"), BR(),
				A(Attr("href", "/other"), "navigate to /other"), BR(),
			),
			H1(header),
			PRE(body),
		),
	).Compile("layout")
)

type App struct {
	URL     string
	Setters []template.Setter
}

func (m *App) setURL(req *http.Request) {
	m.URL = req.URL.Path
}

func (m *App) setHeader() {
	m.Setters = append(m.Setters, header.Setf("header at <%#v>", m.URL))
}

func (m *App) setBody() {
	m.Setters = append(m.Setters, body.Set(PRE("body at", m.URL)))
}

func (m *App) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	m.setURL(req)
	m.setHeader()
	m.setBody()
	layout.Replace(m.Setters...).WriteTo(rw)

	// you can also make a switch on req.Method or switch req.URL.Fragment for subroutes
}

func NewApp() http.Handler {
	return &App{}
}

var Router = router.New()

func main() {
	appRouterFunc := router.RouterFunc(NewApp)

	Router.GET("/", layout)
	Router.GET("/app", appRouterFunc)
	Router.GET("/other", appRouterFunc)
	// or Router.MustHandle("/", method.GET|method.POST, appRouterFunc)

	router.MustMount("/", Router)

	http.ListenAndServe(":8085", nil)
}
