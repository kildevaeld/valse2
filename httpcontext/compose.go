package httpcontext

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kildevaeld/strong"
)

func handlerToMiddleware(r HandlerFunc) MiddlewareHandler {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {

			if err := r(ctx); err != nil {
				return err
			}

			if next != nil {
				return next(ctx)
			}

			return nil

		}
	}
}

func cWrapper(fn func(ctx *Context, next HandlerFunc) error) MiddlewareHandler {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			return fn(ctx, next)
		}
	}
}

func httpHandlerToHandler(fn http.HandlerFunc) HandlerFunc {

	return func(ctx *Context) error {

		writer := newwriterWrapper(ctx)
		defer writer.Close()

		fn(writer, ctx.Request())

		return nil

	}
}

func ToMiddlewareHandler(handler interface{}) (MiddlewareHandler, error) {
	switch h := handler.(type) {
	case func(*Context) error:
		return handlerToMiddleware(h), nil
	case MiddlewareHandler:
		return h, nil
	case func(HandlerFunc) HandlerFunc:
		return h, nil
	case func(ctx *Context, next HandlerFunc) error:
		return cWrapper(h), nil
	case func(http.ResponseWriter, *http.Request):
		return handlerToMiddleware(httpHandlerToHandler(h)), nil
	case http.HandlerFunc:
		return handlerToMiddleware(httpHandlerToHandler(h)), nil
	}

	return nil, fmt.Errorf("middleware is of wrong type '%T'", handler)
}

func ToHandler(handler interface{}) (HandlerFunc, error) {

	switch h := handler.(type) {
	case HandlerFunc:
	case func(*Context) error:
		return h, nil
	case Handler:
		return h.ServeHTTPContext, nil
	case func(http.ResponseWriter, *http.Request):
		return httpHandlerToHandler(h), nil
	case http.HandlerFunc:
		return httpHandlerToHandler(h), nil
	case http.Handler:
		return httpHandlerToHandler(h.ServeHTTP), nil
	default:
		return nil, fmt.Errorf("handler is of wrong type '%T'", handler)
	}
	return nil, nil
}

func Compose(handlers []interface{}) (HandlerFunc, error) {

	last := handlers[len(handlers)-1]

	routeHandler, err := ToHandler(last)
	if err != nil {
		return nil, err
	}

	var middleware MiddlewareHandler

	if len(handlers) > 1 {
		for i := len(handlers) - 1; i >= 0; i-- {
			if middleware, err = ToMiddlewareHandler(handlers[i]); err != nil {
				return nil, err
			}
			routeHandler = middleware(routeHandler)
		}
	}

	return routeHandler, nil
}

func ComposeRun(w http.ResponseWriter, req *http.Request, handlers []interface{}) error {

	handler, err := Compose(handlers)

	if err != nil {
		return err
	}

	return Run(w, req, handler)

}

func Run(w http.ResponseWriter, r *http.Request, handler HandlerFunc) error {

	ctx := Acquire(w, r)
	defer Release(ctx)

	err := handler(ctx)

	if err != nil {
		if err == ErrHandled {
			return nil
		}
		return err
	}

	status := ctx.StatusCode()
	hasBody := ctx.Body() != nil

	if !hasBody && status <= 0 {
		http.NotFound(w, r)
		return nil
	} else if hasBody && status <= 0 {
		status = strong.StatusOK
	}

	w.WriteHeader(status)
	if hasBody {
		_, err := io.Copy(w, ctx.Body())
		if err != nil {
			return err
		}
	}

	return nil
}
