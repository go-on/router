package router

import (
	"net/http"
	"strings"

	"github.com/go-on/router/route"
)

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
		}
		end = strings.Index(path[start:], "/")

		if end == 0 {
			start++
			continue
		}

		if end == -1 {
			end = len(path)
			fin = true
		} else {
			end += start
		}

		if path[start] == route.WILDCARD_SEPARATOR {
			node.wildcard = []byte(path[start+1 : end])

			if node.sub == nil {
				node.sub = newNode()
			}

			node = node.sub
		} else {
			p := path[start:end]
			subnode := node.edges[p]
			if subnode == nil {
				subnode = newNode()
				node.edges[p] = subnode
			}
			node = subnode
		}

		if fin {
			break
		}

		start = end + 1
	}
	node.route = rt
}

func (n *node) FindPlaceholders(start int, end int, req *http.Request) (parms *[]byte, rt *route.Route) {
	return n.findPositions(start+1, end, req, nil)
}

func (n *node) findSlash(req *http.Request, start int, end int) (pos int) {
	for i, r := range req.URL.Path[start:end] {
		if r == '/' {
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
			pArr := make([]byte, 0, (len(n.wildcard)+len(req.URL.Path[start:end])+2)*2)
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
