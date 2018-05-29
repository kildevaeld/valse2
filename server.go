package valse2

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/julienschmidt/httprouter"
	"github.com/kildevaeld/strong"
)

type Valse struct {
	noCopy
	router    *httprouter.Router
	listening bool

	m []MiddlewareHandler
	p sync.Pool
	s *http.Server

	h RequestHandler
	//links LinksFactory
}

func New() *Valse {
	v := &Valse{
		s:      &http.Server{},
		router: httprouter.New(),
		p: sync.Pool{
			New: func() interface{} {
				return &Context{
					u: make(map[string]interface{}),
				}
			},
		},
	}

	v.s.Handler = v

	return v
}

func (v *Valse) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if v.h == nil {
		return
	}
	ctx := v.p.Get().(*Context)
	ctx.req = r
	ctx.res = w

	defer v.cleanup_context(ctx)
	if err := v.h(ctx); err != nil {
		notFoundOrErr(ctx, err)
	}
}

func (v *Valse) Listen(addr string) error {

	if v.listening {
		return errors.New("Already running")
	}
	v.listening = true
	v.s.Addr = addr
	handlers := rWrapper(v.router.ServeHTTP)
	for i := len(v.m) - 1; i >= 0; i-- {
		handlers = v.m[i](handlers)
	}

	v.h = handlers

	return v.s.ListenAndServe()

}

func (v *Valse) Close() error {
	return v.s.Close()
}

func (v *Valse) Use(handlers ...interface{}) *Valse {
	if v.listening {
		panic("cannot add middleware when running.")
	}

	for _, handler := range handlers {
		switch h := handler.(type) {
		case func(*Context) error:
			v.m = append(v.m, mWrapper(h))
		case func(RequestHandler) RequestHandler:
			v.m = append(v.m, h)
		case func(ctx *Context, next RequestHandler) error:
			v.m = append(v.m, cWrapper(h))
		case MiddlewareHandler:
			v.m = append(v.m, h)
		default:
			panic(fmt.Sprintf("middleware is of wrong type %t", handler))
		}
	}

	return v
}

func cpy(hns []MiddlewareHandler, r RequestHandler) []interface{} {
	out := make([]interface{}, len(hns)+1)
	for i, rr := range hns {
		out[i] = rr
	}
	out[len(hns)] = r

	return out
}

func (v *Valse) Mount(path string, group *Group) *Valse {
	for _, route := range group.routes {
		p := route.Path
		if path != "" {
			p = filepath.Join(path, route.Path)
		}

		v.Route(route.Method, p, cpy(group.m, route.Handler)...)
	}
	return v
}

func (v *Valse) Get(path string, handlers ...interface{}) *Valse {
	return v.Route(strong.GET, path, handlers...)
}

func (v *Valse) Post(path string, handlers ...interface{}) *Valse {
	return v.Route(strong.POST, path, handlers...)
}

func (v *Valse) Put(path string, handlers ...interface{}) *Valse {
	return v.Route(strong.PUT, path, handlers...)
}

func (v *Valse) Delete(path string, handlers ...interface{}) *Valse {
	return v.Route(strong.DELETE, path, handlers...)
}

func (v *Valse) Head(path string, handlers ...interface{}) *Valse {
	return v.Route(strong.HEAD, path, handlers...)
}

func (v *Valse) Options(path string, handlers ...interface{}) *Valse {
	return v.Route(strong.OPTIONS, path, handlers...)
}

func (v *Valse) Route(method, path string, handlers ...interface{}) *Valse {
	if len(handlers) == 0 {
		return v
	}

	handler, err :=  v.compose(handlers)

	if err != nil {
		panic(err)
	}
	
	v.router.Handle(method, path, v.handleRequest(handler))

	return v
}

func (v *Valse) compose(handlers []interface{}) (RequestHandler, error) {
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
		hh, err := v.toMiddlewareHandler(h)
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

func (v *Valse) toMiddlewareHandler(handler interface{}) (MiddlewareHandler, error) {
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

func (v *Valse) cleanup_context(ctx *Context) {
	v.p.Put(ctx.reset())
}

func notFoundOrErr(ctx *Context, err error) error {
	/*if ctx.Response().Status() == http.StatusNotFound || err == nil {
		return nil
	}*/

	status := http.StatusInternalServerError
	if e, ok := err.(*strong.HttpError); ok {
		ctx.Error(e.Message(), e.Code())
		return nil
	}

	ctx.Error(err.Error(), status)

	return nil
}

func (v *Valse) handleRequest(handler RequestHandler) httprouter.Handle {

	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

		ctx := v.p.Get().(*Context)
		ctx.req = r
		ctx.res = w
		ctx.params = ps

		defer v.cleanup_context(ctx)
		if err := handler(ctx); err != nil {
			notFoundOrErr(ctx, err)
		}

	}
}
