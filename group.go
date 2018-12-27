package valse2

import (
	"path/filepath"

	"github.com/kildevaeld/valse2/router"
	"go.uber.org/zap"

	"github.com/kildevaeld/valse2/httpcontext"

	"github.com/kildevaeld/strong"
)

type Namespace interface {
	Use(handlers ...interface{}) Namespace
	Get(path string, handlers ...interface{}) Namespace
	Post(path string, handlers ...interface{}) Namespace
	Put(path string, handlers ...interface{}) Namespace
	Patch(path string, handlers ...interface{}) Namespace
	Delete(path string, handlers ...interface{}) Namespace
	Head(path string, handlers ...interface{}) Namespace
	Options(path string, handlers ...interface{}) Namespace
	Route(method, path string, handlers ...interface{}) Namespace
	Mount(path string, group Mountable, midlewares ...interface{}) Namespace
}

type Mountable interface {
	Compose(root string) (httpcontext.HandlerFunc, error)
}

type groupMount struct {
	path  string
	group Mountable
	m     []httpcontext.MiddlewareHandler
}

type Group struct {
	m      []httpcontext.MiddlewareHandler
	routes []Route
	groups []groupMount
}

func (g *Group) Use(handlers ...interface{}) Namespace {

	for _, handler := range handlers {
		m, err := httpcontext.ToMiddlewareHandler(handler)
		if err != nil {
			panic(err)
		}
		g.m = append(g.m, m)
	}

	return g
}

func (g *Group) Get(path string, handlers ...interface{}) Namespace {
	return g.Route(strong.GET, path, handlers...)
}

func (g *Group) Post(path string, handlers ...interface{}) Namespace {
	return g.Route(strong.POST, path, handlers...)
}

func (g *Group) Put(path string, handlers ...interface{}) Namespace {
	return g.Route(strong.PUT, path, handlers...)
}

func (g *Group) Patch(path string, handlers ...interface{}) Namespace {
	return g.Route(strong.PATCH, path, handlers...)
}

func (g *Group) Delete(path string, handlers ...interface{}) Namespace {
	return g.Route(strong.DELETE, path, handlers...)
}

func (g *Group) Head(path string, handlers ...interface{}) Namespace {
	return g.Route(strong.HEAD, path, handlers...)
}

func (g *Group) Options(path string, handlers ...interface{}) Namespace {
	return g.Route(strong.OPTIONS, path, handlers...)
}

func (g *Group) Route(method, path string, handlers ...interface{}) Namespace {
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

func (s *Group) Mount(path string, group Mountable, middleware ...interface{}) Namespace {

	var m []httpcontext.MiddlewareHandler
	for _, mi := range middleware {
		h, e := httpcontext.ToMiddlewareHandler(mi)
		if e != nil {
			panic(e)
		}
		m = append(m, h)
	}

	s.groups = append(s.groups, groupMount{
		path:  path,
		group: group,
		m:     m,
	})

	return s
}

func (g *Group) Compose(root string) (httpcontext.HandlerFunc, error) {
	router := router.New()
	for _, route := range g.routes {
		p := route.Path
		if root != "" {
			p = filepath.Join(root, route.Path)
		}

		handler, err := httpcontext.Compose(cpy(g.m, route.Handler))

		if err != nil {
			return nil, err
		}

		zap.L().Debug("register route", zap.String("method", route.Method), zap.String("path", p))

		router.Handle(route.Method, p, handler)
	}

	routes := []httpcontext.HandlerFunc{router.ServeHTTPContext}

	for _, subgroup := range g.groups {
		p := subgroup.path
		if root != "" {
			p = filepath.Join(root, subgroup.path)
		}

		handler, err := subgroup.group.Compose(p)
		if err != nil {
			return nil, err
		}

		if handler, err = httpcontext.Compose(cpy(g.m, handler)); err != nil {
			return nil, err
		}

		if len(subgroup.m) > 0 {
			if handler, err = httpcontext.Compose(cpy(subgroup.m, handler)); err != nil {
				return nil, err
			}
		}

		zap.L().Debug("register subgroup", zap.String("path", p))

		routes = append(routes, handler)

	}

	return func(ctx *httpcontext.Context) error {
		var err error
		for _, route := range routes {
			if err = route(ctx); err != nil {
				if err != strong.ErrNotFound {
					break
				}
			}

			hasBody := ctx.Body() != nil
			status := ctx.StatusCode()

			if hasBody || status > 0 {
				break
			}
		}

		return err
	}, nil

}

func NewGroup() *Group {
	return &Group{}
}
