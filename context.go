package web

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Context struct {
	// origin objects
	Writer http.ResponseWriter
	Req    *http.Request
	// request info
	FullPath string
	Method   string
	Param    map[string]string
	// response info
	handlers   HandlersChain
	StatusCode int
	index      int8
	engine     *Engine
}

func newContext() *Context {
	return &Context{}
}

func (c *Context) reset() {
	c.Writer = nil
	c.Req = nil
	c.FullPath = ""
	c.Method = ""
	c.Param = nil
	c.handlers = nil
	c.StatusCode = 0
}

func (c *Context) Init(w http.ResponseWriter, req *http.Request) {
	c.Writer = w
	c.Req = req
	c.FullPath = req.URL.Path
	c.Method = req.Method
	c.index = -1
}

func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

func (c *Context) PostQuery(key string) string {
	return c.Req.URL.Query().Get(key)
}

func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) Json(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

func (c *Context) HTML(code int, name string, data interface{}) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	if err := c.engine.htmlTemplate.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.Fail(500, err.Error())
	}
}

func (c *Context) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)
		c.index++
	}
}

func (c *Context) Fail(code int, msg string) {
	c.Status(code)
	c.Writer.Write([]byte(msg))
}
