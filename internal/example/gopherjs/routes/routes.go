package routes

import (
	"github.com/go-on/method"

	"github.com/go-on/router/route"
)

var (
	ADMIN   = "/admin"
	Id_     = "article_id"
	Article = route.New("/article/:"+Id_, method.GET)
)

func init() {
	route.Mount(ADMIN, Article)
}
