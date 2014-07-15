package router

import (
	"github.com/go-on/router/route"
	"github.com/metakeule/menu"
)

type MenuParameter interface {
	Params(*route.Route) []map[string]string

	// Text returns the menu text for the given route with the given
	// parameters
	Text(rt *route.Route, params map[string]string) string
}

type MenuAdder interface {
	// Add adds the given item somewhere. Where might be decided
	// by looking at the given route
	Add(item menu.Leaf, rt *route.Route, params map[string]string)
}

// Menu creates a menu item for each route via solver
// and adds it via appender
func (r *Router) Menu(adder MenuAdder, solver MenuParameter) {
	fn := func(mountPoint string, rt *route.Route) {
		if HasParams(rt) {
			paramsArr := solver.Params(rt)
			for _, params := range paramsArr {
				adder.Add(
					menu.Item(solver.Text(rt, params), MustURLMap(rt, params)),
					rt,
					params,
				)
			}

		} else {
			adder.Add(
				menu.Item(solver.Text(rt, nil), MustURL(rt)),
				rt,
				nil,
			)
		}
	}
	r.EachGETRoute(fn)
}
