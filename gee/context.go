package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

//map[string] interface是值 key是string value是interface{}，即是value可以是任意值
//这里把map[string] interface取了个别名H
type H map[string]interface{}

type Context struct {
	//原始对象
	Writer http.ResponseWriter
	Req    *http.Request
	//对path和method提供直接访问
	Path   string
	Method string
	//链接里传的参数
	Params map[string]string

	StatusCode int
	// 中间件
	handlers []HandlerFunc
	index    int
}

func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
		index:  -1,
	}
}

//之所以要加这个for循环是因为，不是每个中间件都会显示的调用next()方法
//next()方法只在需要对执行之前和之后进行处理的时候使用
//加上for循环就可以兼容不写next()的写法
func (c *Context) Next() {
	c.index++
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

//获得链接里面的参数
func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

//返回Post里存放的信息
func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

//返回get里面存放的信息
func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

func (c *Context) Status(code int) {
	c.StatusCode = code
	// WriteHeader函数就是用来写code的
	// The provided code must be a valid HTTP 1xx-5xx status code.
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

func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

func (c *Context) Html(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	c.Writer.Write([]byte(html))
}
