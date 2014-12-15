package routes

import (
	"gopkg.in/go-on/method.v1"

	"gopkg.in/go-on/router.v2/route"
)

var (
	ADMIN   = "/admin"
	Id_     = "article_id"
	Article = route.New("/article/:"+Id_, method.GET)
)

func init() {
	route.Mount(ADMIN, Article)
}
