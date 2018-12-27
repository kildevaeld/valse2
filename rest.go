package valse2

import (
	"fmt"
	"strings"

	"github.com/kildevaeld/strong"
	"github.com/kildevaeld/valse2/httpcontext"
)

type restType int

const (
	restCreate restType = iota
	restUpdate
	restPatch
	restDelete
	restGet
	restList
)

type Rest struct {
	name    string
	methods map[restType]Route
}

func (r *Rest) Create(handlers ...interface{}) *Rest {

	handler, err := httpcontext.Compose(handlers)
	if err != nil {
		panic(err)
	}

	return r.method(restCreate, Route{
		Path:    "/",
		Method:  strong.POST,
		Handler: handler,
	})
}

func (r *Rest) Update(handlers ...interface{}) *Rest {

	handler, err := httpcontext.Compose(handlers)
	if err != nil {
		panic(err)
	}

	return r.method(restCreate, Route{
		Path:    fmt.Sprintf("/:%s_id", strings.ToLower(r.name)),
		Method:  strong.PUT,
		Handler: handler,
	})
}

func (r *Rest) Patch(handlers ...interface{}) *Rest {

	handler, err := httpcontext.Compose(handlers)
	if err != nil {
		panic(err)
	}

	return r.method(restCreate, Route{
		Path:   fmt.Sprintf("/:%s_id", strings.ToLower(r.name)),
		Method: strong.PATCH,
		Handler: func(ctx *httpcontext.Context) error {
			return handler(ctx)
		},
	})
}
func (r *Rest) Delete(handlers ...interface{}) *Rest {

	handler, err := httpcontext.Compose(handlers)
	if err != nil {
		panic(err)
	}
	return r.method(restCreate, Route{
		Path:   fmt.Sprintf("/:%s_id", strings.ToLower(r.name)),
		Method: strong.DELETE,
		Handler: func(ctx *httpcontext.Context) error {
			return handler(ctx)
		},
	})
}
func (r *Rest) Get(handlers ...interface{}) *Rest {

	handler, err := httpcontext.Compose(handlers)
	if err != nil {
		panic(err)
	}
	return r.method(restCreate, Route{
		Path:   fmt.Sprintf("/:%s_id", strings.ToLower(r.name)),
		Method: strong.GET,
		Handler: func(ctx *httpcontext.Context) error {
			return handler(ctx)
		},
	})
}

func (r *Rest) List(handlers ...interface{}) *Rest {
	handler, err := httpcontext.Compose(handlers)
	if err != nil {
		panic(err)
	}
	return r.method(restCreate, Route{
		Path:   "/",
		Method: strong.GET,
		Handler: func(ctx *httpcontext.Context) error {
			return handler(ctx)
		},
	})
}

func (r *Rest) method(t restType, route Route) *Rest {
	r.methods[t] = route
	return r
}

func (r *Rest) Compose(root string) (httpcontext.HandlerFunc, error) {
	group := NewGroup()
	for _, route := range r.methods {
		group.Route(route.Method, route.Path, route.Handler)
	}
	return group.Compose(root)
}

func NewRest(name string) *Rest {
	return &Rest{
		name:    name,
		methods: make(map[restType]Route),
	}
}
