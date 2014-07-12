package routes

import (
	"github.com/go-on/method"
	"github.com/go-on/router/route"
)

type Mountpath string

func (mp Mountpath) MountPath() string {
	return string(mp)
}

var AdminMountPoint = "/admin"

var GetArticle = route.NewRoute("/articles/:id").AddMethod(method.GET)

func init() {
	GetArticle.Router = Mountpath(AdminMountPoint)
}
