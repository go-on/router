package main

import (
	"fmt"
	"github.com/go-on/lib/html"
	"github.com/go-on/lib/internal/menu"
	"github.com/go-on/lib/internal/menu/menuhtml"
	"github.com/go-on/router/example/static/site"
	"os"

	"github.com/go-on/router"
	"github.com/go-on/router/route"
)

type resolver struct {
	subs map[string]*menu.Node
	root *menu.Node
}

func (rs *resolver) Text(rt *route.Route, params map[string]string) string {
	switch rt {
	case site.DRoute:
		return fmt.Sprintf("A: %s B: %s ", params["a"], params["b"])
	case site.HomeRoute:
		return "Home"
	case site.ARoute:
		return "A"
	default:
		panic("unhandled route for text: " + rt.OriginalPath)
	}
}

func (rs *resolver) Params(rt *route.Route) []map[string]string {
	switch rt {
	case site.DRoute:
		return []map[string]string{
			map[string]string{"a": "a0", "b": "b0", "d": "d0.html"},
			map[string]string{"a": "a01", "b": "b0", "d": "d01.html"},
			map[string]string{"a": "a1", "b": "b1", "d": "d1.html"},
			map[string]string{"a": "a2", "b": "b2", "d": "d2.html"},
		}
	default:
		panic("unhandled route: " + rt.OriginalPath)
	}
}

func (rs *resolver) Add(l menu.Leaf, rt *route.Route, params map[string]string) {
	switch rt {
	case site.DRoute:
		b := params["b"]
		sn, has := rs.subs[b]
		if !has {
			sn = &menu.Node{Leaf: menu.Item("category "+b, "")}
			rs.root.Edges = append(rs.root.Edges, sn)
			rs.subs[b] = sn
		}
		sn.Edges = append(sn.Edges, &menu.Node{Leaf: l})
	default:
		rs.root.Edges = append(rs.root.Edges, &menu.Node{Leaf: l})
	}
}

func main() {
	router.Mount("/", site.Router)

	root := &menu.Node{}
	solver := &resolver{
		root: root,
		subs: map[string]*menu.Node{},
	}

	site.Router.Menu(solver, solver)

	menuhtml.NewUL(
		html.Class("menu-open"),
		html.Class("menu-active"),
		html.Class("menu-sub"),
	).WriterTo(root, 4, "/d/a0/x/b0/d0.html").WriteTo(os.Stdout)
}
