package valse2

import (
	"fmt"
	"path/filepath"

	"github.com/kildevaeld/strong"
)

type Group struct {
	m      []MiddlewareHandler
	routes []Route
}

func (g *Group) Use(handlers ...interface{}) *Group {

	for _, handler := range handlers {
		switch h := handler.(type) {
		case func(*Context) error:
			g.m = append(g.m, mWrapper(h))
		case func(RequestHandler) RequestHandler:
			g.m = append(g.m, h)
		case func(ctx *Context, next RequestHandler) error:
			g.m = append(g.m, cWrapper(h))
		case MiddlewareHandler:
			g.m = append(g.m, h)
		default:
			panic(fmt.Sprintf("middleware is of wrong type %t", handler))
		}
	}

	return g
}

func (g *Group) Get(path string, handlers ...interface{}) *Group {
	return g.Route(strong.GET, path, handlers...)
}

func (g *Group) Post(path string, handlers ...interface{}) *Group {
	return g.Route(strong.POST, path, handlers...)
}

func (g *Group) Put(path string, handlers ...interface{}) *Group {
	return g.Route(strong.PUT, path, handlers...)
}

func (g *Group) Delete(path string, handlers ...interface{}) *Group {
	return g.Route(strong.DELETE, path, handlers...)
}

func (g *Group) Head(path string, handlers ...interface{}) *Group {
	return g.Route(strong.HEAD, path, handlers...)
}

func (g *Group) Options(path string, handlers ...interface{}) *Group {
	return g.Route(strong.OPTIONS, path, handlers...)
}

func (g *Group) Route(method, path string, handlers ...interface{}) *Group {
	if len(handlers) == 0 {
		return g
	}

	handler, err := compose(handlers)

	if err != nil {
		panic(err)
	}

	g.routes = append(g.routes, Route{
		Method:  method,
		Path:    path,
		Handler: handler,
	})

	return g
}

func (s *Group) Mount(path string, group *Group) *Group {
	for _, route := range group.routes {
		p := route.Path
		if path != "" {
			p = filepath.Join(path, route.Path)
		}

		s.Route(route.Method, p, cpy(group.m, route.Handler)...)
	}
	return s
}

func NewGroup() *Group {
	return &Group{}
}
