package main

import (
	"fmt"

	"gopkg.in/go-on/router.v2/internal/routermenu"
	// "gopkg.in/go-on/lib.v2/html"
	"os"

	"gopkg.in/go-on/lib.v2/internal/menu"
	"gopkg.in/go-on/lib.v2/internal/menu/menuhtml"
	"gopkg.in/go-on/lib.v2/types"
	"gopkg.in/go-on/router.v2/internal/example/static/site"

	"gopkg.in/go-on/router.v2/route"
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
		panic("unhandled route for text: " + rt.DefinitionPath)
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
		panic("unhandled route: " + rt.DefinitionPath)
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
	site.Router.Mount("/", nil)

	root := &menu.Node{}
	solver := &resolver{
		root: root,
		subs: map[string]*menu.Node{},
	}

	routermenu.Menu(site.Router, solver, solver)
	// site.Router.Menu(solver, solver)

	menuhtml.NewUL(
		types.Class("menu-open"),
		types.Class("menu-active"),
		types.Class("menu-sub"),
	).WriterTo(root, 4, "/d/a0/x/b0/d0.html").WriteTo(os.Stdout)
}
