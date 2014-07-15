package router

import (
	"net/http"
	"strings"

	"github.com/go-on/router/route"
)

const (
	dblslash = "//"
	slash    = "/"
)

var sep = '/'

func newNode() *node {
	return &node{edges: make(map[string]*node)}
}

type node struct {
	edges    map[string]*node // the empty key is for the next wildcard node (the node after my wildcard)
	wildcard []byte           //string
	route    *route.Route
	sub      *node
}

func (pn *node) add(path string, rt *route.Route) {

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
			node.wildcard = []byte(wc)

			if node.sub == nil {
				node.sub = newNode()
			}

			node = node.sub
		} else {
			subnode, exist := node.edges[p]
			if !exist {
				subnode = newNode()
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

//func (n *node) FindPlaceholders(startPath int, endPath int, req *http.Request, params *[]byte) (parms *[]byte, rt *route.Route) {
func (n *node) FindPlaceholders(startPath int, endPath int, req *http.Request) (parms *[]byte, rt *route.Route) {
	return n.findPositions(startPath+1, endPath, req, nil)
}

func (n *node) findSlash(req *http.Request, start int, endPath int) (pos int) {
	for i, r := range req.URL.Path[start:endPath] {
		if r == sep {
			return i
		}
	}
	return -1
}

func (n *node) findEdge(start int, endPath int, req *http.Request, params *[]byte) (*[]byte, *route.Route) {
	pos := n.findSlash(req, start, endPath)
	end := start + pos

	if pos == -1 {
		end = endPath
	}

	for k, val := range n.edges {
		if k == req.URL.Path[start:end] {
			if len(val.edges) == 0 && val.wildcard == nil {
				return params, val.route
			}
			return val.findPositions(end+1, endPath, req, params)
		}
	}
	return params, nil
}

func (n *node) findPositions(start int, endPath int, req *http.Request, params *[]byte) (*[]byte, *route.Route) {
	if endPath-start < 1 {
		return params, n.route
	}

	pos := n.findSlash(req, start, endPath)
	end := start + pos

	if pos == -1 {
		end = endPath
	}

	var edgeRoute *route.Route
	params, edgeRoute = n.findEdge(start, endPath, req, params)
	if edgeRoute != nil {
		return params, edgeRoute
	}

	if n.wildcard != nil {
		if params == nil {
			pArr := make([]byte, 0, len(n.wildcard)+len(req.URL.Path[start:end])+2)
			params = &pArr
		}
		*params = append(*params, n.wildcard...)
		*params = append(*params, ("/" + req.URL.Path[start:end] + "/")...)
		if n.sub != nil {
			return n.sub.findPositions(end+1, endPath, req, params)
		}
	}

	return params, nil
}

// stolen from  https://raw.github.com/gocraft/web/master/tree.go and modified

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
