package valse2

import (
	"path/filepath"

	"github.com/kildevaeld/valse2/httpcontext"

	"github.com/kildevaeld/strong"
)

type Group struct {
	m      []httpcontext.MiddlewareHandler
	routes []Route
}

func (g *Group) Use(handlers ...interface{}) *Group {

	for _, handler := range handlers {
		m, err := httpcontext.ToMiddlewareHandler(handler)
		if err != nil {
			panic(err)
		}
		g.m = append(g.m, m)
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

func (g *Group) Patch(path string, handlers ...interface{}) *Group {
	return g.Route(strong.PATCH, path, handlers...)
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

	handler, err := httpcontext.Compose(handlers)

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
