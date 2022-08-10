package web

import (
	"html/template"
	"net/http"
	"sync"
)

type Routes interface {
	Route(string, string, ...HandlerFunc) Routes
	Use(...HandlerFunc) Routes
	GET(string, ...HandlerFunc) Routes
	POST(string, ...HandlerFunc) Routes
	DELETE(string, ...HandlerFunc) Routes
	PATCH(string, ...HandlerFunc) Routes
	PUT(string, ...HandlerFunc) Routes
	OPTIONS(string, ...HandlerFunc) Routes
	HEAD(string, ...HandlerFunc) Routes
}

type HandlerFunc func(ctx *Context)

type HandlersChain []HandlerFunc

type Engine struct {
	*RouterGroup

	middleWares HandlersChain
	router      *router

	pool sync.Pool
	// html render
	htmlTemplate *template.Template
	funcMap      template.FuncMap
}

func (engine *Engine) Route(method string, fullPath string, handlers ...HandlerFunc) Routes {
	engine.addRoute(method, fullPath, handlers)
	return engine
}

func (engine *Engine) Use(handlers ...HandlerFunc) Routes {
	engine.use(handlers)
	return engine
}

func (engine *Engine) GET(fullPath string, handlers ...HandlerFunc) Routes {
	engine.addRoute(http.MethodGet, fullPath, handlers)
	return engine
}

func (engine *Engine) POST(fullPath string, handlers ...HandlerFunc) Routes {
	engine.addRoute(http.MethodPost, fullPath, handlers)
	return engine
}

func (engine *Engine) DELETE(fullPath string, handlers ...HandlerFunc) Routes {
	engine.addRoute(http.MethodDelete, fullPath, handlers)
	return engine
}

func (engine *Engine) PATCH(fullPath string, handlers ...HandlerFunc) Routes {
	engine.addRoute(http.MethodPatch, fullPath, handlers)
	return engine
}

func (engine *Engine) PUT(fullPath string, handlers ...HandlerFunc) Routes {
	engine.addRoute(http.MethodPut, fullPath, handlers)
	return engine
}

func (engine *Engine) OPTIONS(fullPath string, handlers ...HandlerFunc) Routes {
	engine.addRoute(http.MethodOptions, fullPath, handlers)
	return engine
}

func (engine *Engine) HEAD(fullPath string, handlers ...HandlerFunc) Routes {
	engine.addRoute(http.MethodHead, fullPath, handlers)
	return engine
}

func (engine *Engine) Group(prefix string, middleWares HandlersChain) *RouterGroup {
	engine.RouterGroup = &RouterGroup{
		engine:      engine,
		prefix:      prefix,
		middleWares: middleWares,
	}
	return engine.RouterGroup
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := engine.pool.Get().(*Context)
	ctx.reset()
	ctx.Init(w, req)
	engine.router.handle(ctx)
}

func (engine *Engine) addRoute(method string, fullPath string, handlers HandlersChain) {
	engine.router.addRoute(method, fullPath, handlers)
}

func (engine *Engine) combineHandlers(handlers HandlersChain) HandlersChain {
	finalSize := len(engine.middleWares) + len(handlers)
	mergedHandlers := make(HandlersChain, finalSize)
	copy(mergedHandlers, engine.middleWares)
	copy(mergedHandlers[len(engine.middleWares):], handlers)
	return mergedHandlers
}

func (engine *Engine) use(hanlers HandlersChain) {
	engine.middleWares = append(engine.middleWares, hanlers...)
}

func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

func (engine *Engine) LoadHTMLGlob(fullPath string) {
	engine.htmlTemplate = template.Must(template.New("").Funcs(engine.funcMap).Parse(fullPath))
}

var _ Routes = (*Engine)(nil)

func New() *Engine {
	return &Engine{
		router: newRouter(),
		pool: sync.Pool{
			New: func() interface{} {
				return newContext()
			},
		},
	}
}

func Default() *Engine {
	defaultEngine := New()
	defaultEngine.Use(Logger(), Recover())
	return defaultEngine
}
