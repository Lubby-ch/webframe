package web

import (
	"net/http"
	"path"
)

type Router interface {
	Routes
	Group(string, HandlersChain) *RouterGroup
}

var _ Router = (*RouterGroup)(nil)

type RouterGroup struct {
	prefix      string
	middleWares HandlersChain
	engine      *Engine
	parent      *RouterGroup
}

func (g *RouterGroup) Group(prefix string, middleWares HandlersChain) *RouterGroup {
	engine := g.engine
	newGroup := &RouterGroup{
		prefix:      g.prefix + prefix,
		middleWares: middleWares,
		engine:      engine,
		parent:      g,
	}
	return newGroup
}

func (g *RouterGroup) Route(method string, fullPath string, handlers ...HandlerFunc) Routes {
	g.engine.addRoute(method, g.prefix+fullPath, g.combineHandlers(handlers))
	return g
}

func (g *RouterGroup) Use(handlers ...HandlerFunc) Routes {
	g.use(handlers)
	return g
}

func (g *RouterGroup) GET(fullPath string, handlers ...HandlerFunc) Routes {
	g.engine.addRoute(http.MethodGet, g.prefix+fullPath, g.combineHandlers(handlers))
	return g
}

func (g *RouterGroup) POST(fullPath string, handlers ...HandlerFunc) Routes {
	g.engine.addRoute(http.MethodPost, g.prefix+fullPath, g.combineHandlers(handlers))
	return g
}

func (g *RouterGroup) DELETE(fullPath string, handlers ...HandlerFunc) Routes {
	g.engine.addRoute(http.MethodDelete, g.prefix+fullPath, g.combineHandlers(handlers))
	return g
}

func (g *RouterGroup) PATCH(fullPath string, handlers ...HandlerFunc) Routes {
	g.engine.addRoute(http.MethodPatch, g.prefix+fullPath, g.combineHandlers(handlers))
	return g
}

func (g *RouterGroup) PUT(fullPath string, handlers ...HandlerFunc) Routes {
	g.engine.addRoute(http.MethodPut, g.prefix+fullPath, g.combineHandlers(handlers))
	return g
}

func (g *RouterGroup) OPTIONS(fullPath string, handlers ...HandlerFunc) Routes {
	g.engine.addRoute(http.MethodOptions, g.prefix+fullPath, g.combineHandlers(handlers))
	return g
}

func (g *RouterGroup) HEAD(fullPath string, handlers ...HandlerFunc) Routes {
	g.engine.addRoute(http.MethodHead, g.prefix+fullPath, g.combineHandlers(handlers))
	return g
}

func (g *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(g.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param["filepath"]
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

func (g *RouterGroup) Static(relativePath, root string) {
	handler := g.createStaticHandler(relativePath, http.Dir(root))
	fullPath := path.Join(relativePath, "/*filepath")
	g.GET(fullPath, handler)
}

func (g *RouterGroup) combineHandlers(handlers HandlersChain) HandlersChain {
	finalSize := len(g.middleWares) + len(handlers)
	mergedHandlers := make(HandlersChain, finalSize)
	copy(mergedHandlers, g.middleWares)
	copy(mergedHandlers[len(g.middleWares):], handlers)
	return mergedHandlers
}

func (g *RouterGroup) use(handlers HandlersChain) {
	g.middleWares = append(g.middleWares, handlers...)
}
