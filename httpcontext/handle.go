package httpcontext

type HandlerFunc func(ctx *Context) error

type Handler interface {
	ServeHTTPContext(ctx *Context) error
}

type MiddlewareHandler func(next HandlerFunc) HandlerFunc

type handlerFuncWrap struct {
	fn HandlerFunc
}

func (h *handlerFuncWrap) ServeHTTPContext(ctx *Context) error {
	return h.fn(ctx)
}

func ToHandler(fn HandlerFunc) Handler {
	return &handlerFuncWrap{fn}
}
