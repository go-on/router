package router

// stolen from  https://raw.github.com/gocraft/web/master/tree.go and modified

import (
	"net/http"
	"strings"

	"github.com/go-on/method"
	"github.com/go-on/router/route"
)

// todo: base the node mangling on https://github.com/julienschmidt/httprouter

type pathNode struct {
	// Given the next segment s, if edges[s] exists, then we'll look there first.
	edges map[string]*pathNode

	// If set, failure to match on edges will match on wildcard
	wildcard *pathNode

	// If set, and we have nothing left to match, then we match on this node
	leaf *pathLeaf
}

// For the route /admin/forums/:forum_id/suggestions/:suggestion_id
// We'd have wildcards = ["forum_id", "suggestion_id"]
type pathLeaf struct {
	// names of wildcards that lead to this leaf. eg, ["category_id"] for the wildcard ":category_id"
	wildcards []string

	// Pointer back to the route
	*route.Route
}

func newPathNode() *pathNode {
	return &pathNode{edges: make(map[string]*pathNode)}
}

/*
func (pn *pathNode) inspect(indent int) string {
	var buf bytes.Buffer
	for p, edg := range pn.edges {
		fmt.Fprintf(&buf, "%s/%s\n%s\n", strings.Repeat("\t", indent), p, edg.inspect(indent+1))
	}
	if pn.wildcard != nil {
		fmt.Fprintf(&buf, "%s*\n%s", strings.Repeat("\t", indent), pn.wildcard.inspect(indent+1))
	}
	if pn.leaf != nil && pn.leaf.Route != nil {
		fmt.Fprintf(&buf, "%s\n%s", strings.Repeat("\t", indent), pn.leaf.Route.inspect(indent))
	}

	return buf.String()
}
*/

func (pn *pathNode) add(path string, v method.Method, handler http.Handler, router *Router) error {
	return pn.addInternal(path, splitPath(path), v, handler, nil, router)
}

func (pn *pathNode) addInternal(originalPath string, segments []string, v method.Method, handler http.Handler, wildcards []string, router *Router) error {
	if len(segments) == 0 {
		if pn.leaf == nil {
			path := "/" + strings.Join(segments, "/")
			rrt := route.NewRoute(path)
			rrt.Router = router
			rrt.OriginalPath = originalPath
			pn.leaf = &pathLeaf{Route: rrt, wildcards: wildcards}
		}

		switch v {
		case method.GET:
			pn.leaf.Route.GETHandler = handler
		case method.POST:
			pn.leaf.Route.POSTHandler = handler
		case method.PUT:
			pn.leaf.Route.PUTHandler = handler
		case method.DELETE:
			pn.leaf.Route.DELETEHandler = handler
		case method.PATCH:
			pn.leaf.Route.PATCHHandler = handler
		case method.HEAD:
			pn.leaf.Route.HEADHandler = handler
		}
		return nil
		//return pn.leaf.Route.AddHandler(handler, v)

	}
	seg := segments[0]
	wc, wcName := isWildcard(seg)
	if wc {
		if pn.wildcard == nil {
			pn.wildcard = newPathNode()
		}
		return pn.wildcard.addInternal(originalPath, segments[1:], v, handler, append(wildcards, wcName), router)
	}
	subPn, ok := pn.edges[seg]
	if !ok {
		subPn = newPathNode()
		pn.edges[seg] = subPn
	}
	return subPn.addInternal(originalPath, segments[1:], v, handler, wildcards, router)

}

func (pn *pathNode) Match(path string) (leaf *pathLeaf, wildcards []string) {
	// Bail on invalid paths.
	if len(path) == 0 || path[0] != '/' {
		return nil, nil
	}

	return pn.match(splitPath(path), nil)
	/*
		if len(wc) > 0 {
			wildcards := make(map[string]string, len(wc))
			for i, val := range wc {
				wildcards[l.wildcards[i]] = val
			}
		}
		leaf = l
	*/
	// return
}

// Segments is like ["admin", "users"] representing "/admin/users"
// wildcardValues are the actual values accumulated when we match on a wildcard.
func (pn *pathNode) match(segments []string, wildcardValues []string) (leaf *pathLeaf, wildcardMap []string) {
	// Handle leaf nodes:
	if len(segments) == 0 {
		leaf = pn.leaf
		if leaf == nil {
			return
		}

		if len(wildcardValues) != 0 && (len(pn.leaf.wildcards) == len(wildcardValues)) {
			wildcardMap = wildcardValues
		}

		return
		//return pn.leaf, makeWildcardMap(pn.leaf, wildcardValues)
	}

	var seg string
	seg, segments = segments[0], segments[1:]

	subPn, ok := pn.edges[seg]
	if ok {
		leaf, wildcardMap = subPn.match(segments, wildcardValues)
	}

	if leaf == nil && pn.wildcard != nil {
		leaf, wildcardMap = pn.wildcard.match(segments, append(wildcardValues, seg))
	}

	return leaf, wildcardMap
}

// key is a non-empty path segment like "admin" or ":category_id" or ":category_id:\d+"
// Returns true if it's a wildcard, and if it is, also returns it's name / regexp.
// Eg, (true, "category_id", "\d+")
func isWildcard(key string) (bool, string) {
	if key[0] == ':' {
		substrs := strings.SplitN(key[1:], ":", 2)
		return true, substrs[0]
	} else {
		return false, ""
	}
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

func makeWildcardMap(leaf *pathLeaf, wildcards []string) map[string]string {
	if leaf == nil {
		return nil
	}

	leafWildcards := leaf.wildcards

	if len(wildcards) == 0 || (len(leafWildcards) != len(wildcards)) {
		return nil
	}

	// At this point, we know that wildcards and leaf.wildcards match in length.
	assoc := make(map[string]string)
	for i, w := range wildcards {
		assoc[leafWildcards[i]] = w
	}

	return assoc
}
