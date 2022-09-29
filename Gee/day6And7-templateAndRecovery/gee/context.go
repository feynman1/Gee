package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type H map[string]interface{}

type Context struct {
	//origin method
	Writer http.ResponseWriter
	Req    *http.Request
	//request info
	Path   string
	Method string
	//response info
	StatusCode int
	Params     map[string]string
	//middlewares
	index       int
	middlewares []HandlerFunc
	//engine
	engine *Engine
}

func newContext(writer http.ResponseWriter, req *http.Request, engine *Engine) *Context {
	return &Context{
		Writer: writer, Req: req,
		Path: req.URL.Path, Method: req.Method,
		index:  -1,
		engine: engine,
	}
}

func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

func (c *Context) Status(code int) {
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

func (c *Context) Next() {
	c.index++
	s := len(c.middlewares)
	for ; c.index < s; c.index++ {
		c.middlewares[c.index](c)
	}
}

func (c *Context) Fail(code int, err string) {
	c.index = len(c.middlewares)
	c.JSON(code, H{"message:": err})
}

// -----------------------Create String/Json/Data/Html Response-----------------------
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Context-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values)))
}

func (c *Context) JSON(code int, obj interface{}) {
	c.Status(code)
	c.SetHeader("Context-Type", "application/json")
	encoder := json.NewEncoder(c.Writer)
	if error := encoder.Encode(obj); error != nil {
		http.Error(c.Writer, error.Error(), code)
	}
}

func (c *Context) DATA(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

func (c *Context) HTML(code int, html string, data interface{}) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	//c.Writer.Write([]byte(html))
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, html, data); err != nil {
		c.Fail(500, err.Error())
	}
}
