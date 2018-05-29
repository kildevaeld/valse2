package valse2

type RequestHandler func(*Context) error

type MiddlewareHandler func(next RequestHandler) RequestHandler



type ValseHTTPHandler interface {
	ServeHTTP(*Context) error
}

type Route struct {
	Method  string
	Path    string
	Handler RequestHandler
}
