package handler

import "net/http"

// Route 定义了单条路由的配置声明
type Route struct {
	Method      string // 例如 "GET", "POST"
	Path        string
	Handler     http.HandlerFunc
	Middlewares []func(http.Handler) http.Handler // 允许为单条路由绑定独立中间件
}

// Controller 接口，所有业务结构体需要实现此方法以暴露路由配置
type Controller interface {
	InitRoutes() []Route
}

// Router 实现了 http.Handler 接口，负责将请求分发给关联的结构体方法
type Router struct {
	routes map[string]http.HandlerFunc
}

// NewRouter 创建并初始化路由映射字典
func NewRouter(c Controller) *Router {
	r := &Router{
		routes: make(map[string]http.HandlerFunc),
	}
	for _, route := range c.InitRoutes() {
		// 生成内部精确查找的 key (Method + Path)
		key := route.Path
		if route.Method != "" {
			key = route.Method + " " + route.Path
		}
		r.routes[key] = route.Handler
	}
	return r
}

// ServeHTTP 使得嵌入 Router 的结构体自动成为标准 http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// 优先尝试精确匹配 Method + Path
	key := req.Method + " " + req.URL.Path
	if h, ok := r.routes[key]; ok {
		h(w, req)
		return
	}
	// 回退到仅匹配 Path (兼容未配置 Method 的通用路由)
	if h, ok := r.routes[req.URL.Path]; ok {
		h(w, req)
		return
	}
	http.NotFound(w, req)
}
