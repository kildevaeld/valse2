package httpcontext

type HandlerFunc func(ctx *Context) error

type Handler interface {
	ServeHTTPContext(ctx *Context) error
}

type MiddlewareHandler func(next HandlerFunc) HandlerFunc
