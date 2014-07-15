package router

// stolen from  https://raw.github.com/gocraft/web/master/tree.go and modified
/*
import (
	"strings"

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

func (pn *pathNode) add(path string, rt *route.Route) {
	pn._add(splitPath(path), nil, rt)
}

func (pn *pathNode) _add(segments []string, wildcards []string, rt *route.Route) {
	if len(segments) == 0 {
		if pn.leaf == nil {
			pn.leaf = &pathLeaf{Route: rt, wildcards: wildcards}
		}
		return
	}
	seg := segments[0]
	wc, wcName := isWildcard(seg)
	if wc {
		if pn.wildcard == nil {
			pn.wildcard = newPathNode()
		}
		pn.wildcard._add(segments[1:], append(wildcards, wcName), rt)
		return
	}
	subPn, ok := pn.edges[seg]
	if !ok {
		subPn = newPathNode()
		pn.edges[seg] = subPn
	}
	subPn._add(segments[1:], wildcards, rt)
}

func (pn *pathNode) Match(path string) (leaf *pathLeaf, wildcards []string) {
	// Bail on invalid paths.
	if len(path) == 0 || path[0] != '/' {
		return nil, nil
	}

	return pn.match(splitPath(path), nil)
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
*/
