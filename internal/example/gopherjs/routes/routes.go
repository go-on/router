package routes

import "github.com/go-on/router/route"

type Mountpath string

func (mp Mountpath) MountPath() string {
	return string(mp)
}

var AdminMountPoint = "/admin"

var GetArticle = route.New("/articles/:id")

func init() {
	GetArticle.Router = Mountpath(AdminMountPoint)
}