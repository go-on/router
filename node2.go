package router

import (
	"fmt"
	"strings"

	"github.com/go-on/router/route"
)

/*
/abc/:x => "" => edge["abc"] => ":x"$
/:y/abc => ":y" => edge["abc"] => ""$

GET /abc/hu => {abc} =>


*/

var dbg bool = false

func debug(s string) {
	if dbg {
		fmt.Println(s)
	}
}

func debugf(s string, v ...interface{}) {
	debug(fmt.Sprintf(s, v...))
}

type pos [3]int // first int: -1 => regular mountpath, otherwise: no of placeholder
// second int: startpos third int endpos

var sep = '/'

func newPathNode() *pathNode {
	return &pathNode{edges: make(map[string]*pathNode)}
}

type pathNode struct {
	edges    map[string]*pathNode // the empty key is for the next wildcard node (the node after my wildcard)
	wildcard string
	route    *route.Route
	// route    http.Handler
}

type paramQuery struct {
	wildcards []string //[]byte
	found     [][3]int
	route     *route.Route
	path      string
}

func (wc *paramQuery) ToMap() (params map[string]string) {
	params = map[string]string{}
	for _, f := range wc.found {
		if f[0] != -1 {
			params[wc.wildcards[f[0]]] = wc.path[f[1]:f[2]]
		}
	}
	return
}

func (wc *paramQuery) ParamStr() (params string) {
	// res += _k + "/" + _v + "//"
	for _, f := range wc.found {
		if f[0] != -1 {
			params += wc.wildcards[f[0]] + "/" + wc.path[f[1]:f[2]] + "//"
			// params[wc.wildcards[f[0]]] = wc.path[f[1]:f[2]]
		}
	}
	return
}

func (pn *pathNode) add(path string, rt *route.Route) {
	//pathArr := splitPath(path)

	node := pn
	var start int = 1
	var end int
	var fin bool
	// debug("register " + path)
	// var pa string = path
	for {
		if start >= len(path) {
			break
			panic("unaccessible")
		}
		end = strings.Index(path[start:], "/")
		// debugf("start: %d end: %d\n", start, end)

		if end == 0 {
			start++
			continue
			panic("unaccessible")
			// end = strings.Index(path[start:], "/")
		}

		if end == -1 {
			end = len(path)
			fin = true
		} else {
			end += start
		}
		// debugf("revised start: %d end: %d\n", start, end)
		p := path[start:end]
		if ok, wc := isWildcard(p); ok {
			// debugf("wildcard: %#v (%#v)\n", wc, p)
			node.wildcard = wc //append(node.wildcard, wc)

			subnode, exist := node.edges[""]
			if !exist {
				subnode = newPathNode()
				node.edges[""] = subnode
			}
			node = subnode
		} else {
			// debugf("subnode: %#v\n", p)
			subnode, exist := node.edges[p]
			if !exist {
				subnode = newPathNode()
				node.edges[p] = subnode
			}
			node = subnode
		}

		if fin {
			break
			panic("unaccessible")
		}

		start = end + 1
	}

	node.route = rt

	// fmt.Printf("node: %#v\n", pn)
	//	pn._add(splitPath(path), nil, rt)
}

func (n *pathNode) FindPlaceholders(wc *paramQuery) {
	// wc.path = wc.path[1:]
	// debugf("searching for: %#v\n", wc.path)
	// return
	n._FindPositions(1, wc)
}

func (n *pathNode) findSlash(wc *paramQuery, start int) (pos int) {
	// fmt.Printf("searching from pos: %d of %d\n", start, len(wc.path)-1)
	for i, r := range wc.path[start:] {
		if r == sep {
			return i
		}
	}
	return -1
}

func (n *pathNode) _FindEdge(start int, wc *paramQuery) (found bool) {
	if len(wc.path)-start < 1 {
		wc.route = n.route
		// debugf("foundroute: %#v\n", n.route)
		return true
	}

	pos := n.findSlash(wc, start)
	end := start + pos

	if pos == -1 {
		end = len(wc.path)
	}

	for k, val := range n.edges {
		if k == wc.path[start:end] {
			// wc.found = append(wc.found, [3]int{-1, start, end})
			// fmt.Printf("foundedge: %#v %T\n", k, wc.route)
			if len(val.edges) == 0 && val.wildcard == "" {
				wc.route = val.route
				return true
			}
			val._FindPositions(end+1, wc)
			return true
		}
	}
	//fmt.Printf("foundroute: %#v\n", n.route)

	//wc.route = n.route
	return false
}

/*
func (n *pathNode) _FindWildcards(wcCounter int, start int, wc *wildcards) (end int) {
	if len(wc.path)-start < 1 {
		return start
	}

	if wcCounter >= len(n.wildcard)-1 {
		return start
	}

	wc.found = append(wc.found, [3]int{len(wc.wildcards), start, end})
	wc.wildcards = append(wc.wildcards, n.wildcard)
}
*/

func (n *pathNode) _FindPositions(start int, wc *paramQuery) {
	if len(wc.path)-start < 1 {
		wc.route = n.route
		return
	}

	pos := n.findSlash(wc, start)
	end := start + pos

	if pos == -1 {
		end = len(wc.path)
	}
	if n._FindEdge(start, wc) {
		return
	}

	// return
	if len(n.wildcard) > 0 {
		wc.found = append(wc.found, [3]int{len(wc.wildcards), start, end})
		wc.wildcards = append(wc.wildcards, n.wildcard)
		next, has := n.edges[""]
		if has {
			next._FindPositions(end+1, wc)
		}
	}

	return
}

// key is a non-empty path segment like "admin" or ":category_id" or ":category_id:\d+"
// Returns true if it's a wildcard, and if it is, also returns it's name / regexp.
// Eg, (true, "category_id", "\d+")
func isWildcard(key string) (is bool, wc string) {
	if key[0] == ':' {
		substrs := strings.SplitN(key[1:], ":", 2)
		is, wc = true, substrs[0]
	}
	return
}

// "/" -> []
// "/admin" -> ["admin"]
// "/admin/" -> ["admin"]
// "/admin/users" -> ["admin", "users"]
func splitPath(key string) []string {
	elements := strings.Split(key, "/")
	if elements[0] == "" {
		elements = elements[1:]
	}
	if elements[len(elements)-1] == "" {
		elements = elements[:len(elements)-1]
	}
	return elements
}
