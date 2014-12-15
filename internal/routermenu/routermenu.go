package routermenu

import (
	"github.com/go-on/lib/internal/menu"
	"gopkg.in/go-on/method.v1"
	"gopkg.in/go-on/router.v2"
	"gopkg.in/go-on/router.v2/route"
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
func Menu(r *router.Router, adder MenuAdder, solver MenuParameter) {
	fn := func(mountPoint string, rt *route.Route) {
		if rt.HasMethod(method.GET) {
			if rt.HasParams() {
				paramsArr := solver.Params(rt)
				for _, params := range paramsArr {
					adder.Add(
						menu.Item(solver.Text(rt, params), rt.MustURLMap(params)),
						rt,
						params,
					)
				}

			} else {
				adder.Add(
					menu.Item(solver.Text(rt, nil), rt.MustURL()),
					rt,
					nil,
				)
			}
		}
	}

	r.EachRoute(fn)
}
