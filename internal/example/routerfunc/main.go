package main

import (
	"fmt"
	"net/http"

	"github.com/go-on/router/route"

	"github.com/go-on/method"
	"github.com/go-on/router"
)

type Person struct {
	Id       string
	EMail    string
	prepared bool
}

func (p *Person) Load(req *http.Request) {
	p.Id = req.FormValue(":person_id")
}

func (p Person) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	p.Load(req)

	switch req.URL.Fragment {
	case ArticlesRoute.DefinitionPath:
		switch req.Method {
		case method.GET.String():
			(&Article{&p, ""}).GETList(rw, req)
		case method.POST.String():
			(&Article{&p, ""}).POST(rw, req)
		}
	case ArticleRoute.DefinitionPath:
		(&Article{&p, ""}).GET(rw, req)
	case CommentsRoute.DefinitionPath:
		a := &Article{&p, ""}
		a.Load(req)
		(&Comment{a}).GETList(rw, req)
	default:
	}
}

type Article struct {
	*Person
	Id string
}

func (a *Article) Load(req *http.Request) {
	a.Id = req.FormValue(":article_id")
}

func (a *Article) GETList(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "articles for person with id %s", a.Person.Id)
}

func (a *Article) GET(rw http.ResponseWriter, req *http.Request) {
	a.Load(req)
	fmt.Fprintf(rw, "article with id %s for person with id %s", a.Id, a.Person.Id)
}

func (a *Article) POST(rw http.ResponseWriter, req *http.Request) {
	a.Load(req)
	fmt.Fprintf(rw, "new article with title: %#v", req.FormValue("title"))
}

type Comment struct {
	*Article
}

func (c *Comment) GETList(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "comments for article %s of person with id %s", c.Article.Id, c.Article.Person.Id)
}

var (
	personRouter  = router.New()
	ArticlesRoute *route.Route
	ArticleRoute  *route.Route
	CommentsRoute *route.Route
	//personRouterFunc = router.RouterFunc(func() http.Handler { return &Person{} })
	// CommentsRoute    = personRouter.GET("/:person_id/article/:article_id/comment", personRouterFunc)
	// ArticleRoute     = personRouter.GET("/:person_id/article/:article_id", personRouterFunc)
	// ArticlesRoute    = personRouter.HandleMethods("/:person_id/article", personRouterFunc, method.GET, method.POST)
	fn = Person{}.ServeHTTP

	mainRouter = router.New()
)

func init() {
	CommentsRoute = personRouter.GETFunc("/:person_id/article/:article_id/comment", fn)
	ArticleRoute = personRouter.GETFunc("/:person_id/article/:article_id", fn)
	ArticlesRoute = personRouter.HandleMethods("/:person_id/article", http.HandlerFunc(fn), method.GET, method.POST)
}

func main() {
	mainRouter.Handle("/person", personRouter)
	mainRouter.Mount("/", http.DefaultServeMux)
	err := http.ListenAndServe(":8085", nil)
	if err != nil {
		println(err)
	}
}
