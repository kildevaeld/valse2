package valse2

import (
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func rWrapper(handler http.HandlerFunc) RequestHandler {
	return func(ctx *Context) error {
		handler(ctx.res, ctx.req)
		return nil
	}
}

func routerWrapper(handler httprouter.Handle) RequestHandler {
	return func(ctx *Context) error {
		handler(ctx.res, ctx.req, ctx.Params())
		return nil
	}
}

func mWrapper(r RequestHandler) MiddlewareHandler {
	return func(next RequestHandler) RequestHandler {
		return r
	}
}

func cWrapper(fn func(ctx *Context, next RequestHandler) error) MiddlewareHandler {
	return func(next RequestHandler) RequestHandler {
		return func(ctx *Context) error {
			return fn(ctx, next)
		}
	}
}

func toMiddlewareHandler(handler interface{}) (MiddlewareHandler, error) {
	switch h := handler.(type) {
	case func(*Context) error:
		return mWrapper(h), nil
	//case func(*fasthttp.RequestCtx):
	//	return mWrapper(rWrapper(h)), nil
	case func(RequestHandler) RequestHandler:
		return h, nil
	case func(ctx *Context, next RequestHandler) error:
		return cWrapper(h), nil
	case MiddlewareHandler:
		return h, nil

	default:
		return nil, errors.New("middleware is of wrong type")
	}
}

func compose(handlers []interface{}) (RequestHandler, error) {
	last := handlers[len(handlers)-1]

	var routeHandler func(ctx *Context) error
	if fn, ok := last.(func(ctx *Context) error); ok {
		routeHandler = fn
	} else if fn, ok := last.(http.Handler); ok {
		routeHandler = rWrapper(fn.ServeHTTP)
	} else if fn, ok := last.(http.HandlerFunc); ok {
		routeHandler = rWrapper(fn)
	} else if fn, ok := last.(httprouter.Handle); ok {
		routeHandler = routerWrapper(fn)
	} else if fn, ok := last.(ValseHTTPHandler); ok {
		routeHandler = fn.ServeHTTP
	} else if fn, ok := last.(RequestHandler); ok {
		routeHandler = fn
	} else {
		return nil, errors.New("The last handle must be a RequestHandler or a fasthttp Handler")
	}

	var mHandlers []MiddlewareHandler
	for _, h := range handlers[:len(handlers)-1] {
		hh, err := toMiddlewareHandler(h)
		if err != nil {
			return nil, err
		}
		mHandlers = append(mHandlers, hh)
	}

	for i := len(mHandlers) - 1; i >= 0; i-- {
		routeHandler = mHandlers[i](routeHandler)
	}

	// Now compose

	return routeHandler, nil
}
