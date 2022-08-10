package web

import "strings"

type node struct {
	path     string
	fullPath string

	handlers  HandlersChain
	children  []*node
	wildchild bool
}

type methodTree struct {
	method string
	root   *node
}

// 第一个匹配成功的节点，用于插入
func (n *node) matchChild(path string) *node {
	for _, child := range n.children {
		if child.path == path || child.wildchild {
			return child
		}
	}
	return nil
}

// 所有匹配成功的节点，用于查找
func (n *node) matchChildren(path string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.path == path || child.wildchild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

func (n *node) insert(fullPath string, paths []string, handlers HandlersChain) {
	if len(paths) == 0 {
		n.fullPath = fullPath
		n.handlers = handlers
		return
	}
	path := paths[0]
	child := n.matchChild(path)
	if child == nil {
		child = &node{
			path:      path,
			wildchild: path[0] == ':' || path[0] == '*',
		}
		n.children = append(n.children, child)
		if path[0] == '*' && len(paths) > 1 {
			panic("'*' is only permitted at the last of full path")
		}
	}
	child.insert(fullPath, paths[1:], handlers)
}

func (n *node) search(paths []string) *node {
	if len(paths) == 0 || strings.HasPrefix(n.path, "*") {
		if n.fullPath == "" {
			return nil
		}
		return n
	}
	path := paths[0]
	children := n.matchChildren(path)
	for _, child := range children {
		if result := child.search(paths[1:]); result != nil {
			return result
		}
	}
	return nil
}

type methodTrees []*methodTree

func (trees methodTrees) Get(method string) *node {
	for _, tree := range trees {
		if tree.method == method {
			return tree.root
		}
	}
	return nil
}
