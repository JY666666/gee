package gee

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
)

//这个函数是给用户定义的处理函数
type HandlerFunc func(c *Context)

type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc //支持中间件
	parent      *RouterGroup  //支持分组嵌套
	engine      *Engine
}

//Engine实现了ServeHTTP接口
type Engine struct {
	*RouterGroup
	router        *router
	groups        []*RouterGroup
	htmlTemplates *template.Template //html渲染
	funcMap       template.FuncMap   //html渲染
}

func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine

}

func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

//向这个group里面添加中间件
func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}

func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}

func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	//绝对路径是group前缀加上相对路径
	absolutePath := path.Join(group.prefix, relativePath)
	//这一行是把前面的absolutePath转换成，在上一个函数中的root
	//以root为根路径，访问文件
	//假设absolute里面的路径是/asserts
	//root里面路径是./static(这里是相对路径，也可以是绝对路径/usr/jy/blog/static
	//这样来一个请求 127.0.0.1:9999/asserts/css/jy.css 就会转换成本地的文件路径 ./static/css/jy.css
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		//首先判断文件是否存在，不存在直接退出
		//这里的fs就已经可以代表文件在本地的相对路径了
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

//对外暴露的接口，作用是把relativePath 转换成 root文件目录
func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	//注册GET handler，把这个
	group.GET(urlPattern, handler)
}

func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

func (engine *Engine) LoadHtmlGlob(pattern string) {
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}

//其实是Engine实现了Handler接口，ServeHTTP是Handler的函数
//把w和req封装进Context中
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []HandlerFunc
	//每来一个请求，就比较该请求和所有group里面的前缀
	//要是有这个前缀， 就把该group里面的中间件加到这次请求的Context里面
	for _, group := range engine.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}

	c := newContext(w, req)
	//把中间件放进Context中
	c.handlers = middlewares
	c.engine = engine
	engine.router.handle(c)
}
