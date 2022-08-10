package web

import (
	"net/http"
	"strings"
)

type router struct {
	trees methodTrees
}

func newRouter() *router {
	return &router{
		trees: make(methodTrees, 0, 9),
	}
}

func (r *router) addRoute(method, fullpath string, handlers HandlersChain) {
	root := r.trees.Get(method)
	if root == nil {
		root = new(node)
		root.path = "/"
		root.fullPath = "/"
		r.trees = append(r.trees, &methodTree{method: method, root: root})
	}
	root.insert(fullpath, parseFullPath(fullpath), handlers)
}

func (r *router) getRoute(method, fullPath string) (*node, map[string]string) {
	root := r.trees.Get(method)
	if root == nil {
		return nil, nil
	}
	params := make(map[string]string)
	searchPaths := parseFullPath(fullPath)
	n := root.search(searchPaths)
	if n != nil {
		paths := parseFullPath(n.fullPath)
		for i, path := range paths {
			if path[0] == ':' {
				params[path[1:]] = searchPaths[i]
			}
			if path[0] == '*' && len(path) > 0 {
				params[path[1:]] = strings.Join(searchPaths[i:], "/")
				break
			}
		}
	}
	return n, params
}

func (r *router) handle(c *Context) {
	n, param := r.getRoute(c.Method, c.FullPath)
	if n == nil {
		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.FullPath)
	}
	c.Param = param
	if c.handlers != nil {
		c.handlers = n.handlers
		c.Next()
	}
}

func parseFullPath(fullPath string) []string {
	strs := strings.Split(fullPath, "/")
	paths := make([]string, 0)
	paths = append(paths, "/")
	for _, str := range strs {
		if str != "" {
			paths = append(paths, str)
			if str[0] == '*' {
				break
			}
		}
	}
	return paths
}
