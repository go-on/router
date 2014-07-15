package router

import (
	"net/http"
	"strings"

	"github.com/go-on/method"
	"github.com/go-on/router/route"
)

const (
	dblslash = "//"
	slash    = "/"
)

var sep = '/'

func newPathNode() *pathNode {
	return &pathNode{edges: make(map[string]*pathNode)}
}

type pathNode struct {
	edges    map[string]*pathNode // the empty key is for the next wildcard node (the node after my wildcard)
	wildcard string
	route    *route.Route
}

type paramQuery struct {
	params    []byte
	request   *http.Request
	route     *route.Route
	handler   http.Handler
	meth      method.Method
	startPath int
	endPath   int
}

func (wc *paramQuery) SetFragment() {
	if len(wc.params) == 0 {
		wc.request.URL.Fragment = wc.route.OriginalPath + "//"
		return
	}
	wc.request.URL.Fragment = wc.route.OriginalPath + "//" + string(wc.params)
}

func (pn *pathNode) add(path string, rt *route.Route) {

	node := pn
	var start int = 1
	var end int
	var fin bool
	for {
		if start >= len(path) {
			break
			panic("unaccessible")
		}
		end = strings.Index(path[start:], "/")

		if end == 0 {
			start++
			continue
			panic("unaccessible")
		}

		if end == -1 {
			end = len(path)
			fin = true
		} else {
			end += start
		}

		p := path[start:end]
		if ok, wc := isWildcard(p); ok {
			node.wildcard = wc //append(node.wildcard, wc)

			subnode, exist := node.edges[""]
			if !exist {
				subnode = newPathNode()
				node.edges[""] = subnode
			}
			node = subnode
		} else {
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
}

func (n *pathNode) FindPlaceholders(wc *paramQuery) {
	n._FindPositions(wc.startPath+1, wc)
}

func (n *pathNode) findSlash(wc *paramQuery, start int) (pos int) {
	for i, r := range wc.request.URL.Path[start:wc.endPath] {
		if r == sep {
			return i
		}
	}
	return -1
}

func (n *pathNode) _FindEdge(start int, wc *paramQuery) (found bool) {
	if wc.endPath-start < 1 {
		wc.route = n.route
		return true
	}

	pos := n.findSlash(wc, start)
	end := start + pos

	if pos == -1 {
		end = wc.endPath
	}

	for k, val := range n.edges {
		if k == wc.request.URL.Path[start:end] {
			if len(val.edges) == 0 && val.wildcard == "" {
				wc.route = val.route
				return true
			}
			val._FindPositions(end+1, wc)
			return true
		}
	}
	return false
}

func (n *pathNode) _FindPositions(start int, wc *paramQuery) {
	if wc.endPath-start < 1 {
		wc.route = n.route
		return
	}

	pos := n.findSlash(wc, start)
	end := start + pos

	if pos == -1 {
		end = wc.endPath
	}

	if n._FindEdge(start, wc) {
		return
	}

	//	if len(n.wildcard) > 0 {
	if n.wildcard != "" {
		wc.params = append(wc.params, (n.wildcard + "/" + wc.request.URL.Path[start:end] + "//")...)
		if next, has := n.edges[""]; has {
			next._FindPositions(end+1, wc)
		}
	}
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
