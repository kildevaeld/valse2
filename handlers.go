package valse2

import "github.com/kildevaeld/valse2/httpcontext"

// type RequestHandler func(*httpcontext.Context) error

// type MiddlewareHandler func(next httpcontext.Context) httpcontext.Context

// type ValseHTTPHandler interface {
// 	ServeHTTPContext(*httpcontext.Context) error
// }

type Route struct {
	Method  string
	Path    string
	Handler httpcontext.HandlerFunc
}
