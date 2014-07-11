package main

import (
	"fmt"
	"net/http"

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

func (p *Person) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	p.Load(req)

	switch req.URL.Fragment {
	case ArticlesRoute.OriginalPath:
		switch req.Method {
		case method.GET.String():
			(&Article{p, ""}).GETList(rw, req)
		case method.POST.String():
			(&Article{p, ""}).POST(rw, req)
		}
	case ArticleRoute.OriginalPath:
		(&Article{p, ""}).GET(rw, req)
	case CommentsRoute.OriginalPath:
		a := &Article{p, ""}
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
	personRouter     = router.New()
	personRouterFunc = router.RouterFunc(func() http.Handler { return &Person{} })
	CommentsRoute    = personRouter.GET("/:person_id/article/:article_id/comment", personRouterFunc)
	ArticleRoute     = personRouter.GET("/:person_id/article/:article_id", personRouterFunc)
	ArticlesRoute    = personRouter.MustHandleMethod("/:person_id/article", method.GET|method.POST, personRouterFunc)
	mainRouter       = router.New()
)

func main() {
	mainRouter.HandleMethod("/person", method.ALL, personRouter)
	router.MustMount("/", mainRouter)
	err := http.ListenAndServe(":8085", nil)
	if err != nil {
		println(err)
	}
}
