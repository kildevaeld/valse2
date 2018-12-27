package valse2

import (
	"fmt"

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

type RestCallback func(ctx *httpcontext.Context, id string) error

func normalizeHandlers(handlers []interface{}, name string) (httpcontext.HandlerFunc, string, error) {
	param := name + "_id"
	lastIndex := len(handlers) - 1
	if v, ok := handlers[lastIndex].(func(ctx *httpcontext.Context, id string) error); ok {
		handlers[lastIndex] = func(ctx *httpcontext.Context) error {
			id := ctx.Params().ByName(param)
			if id == "" {
				return strong.NewHTTPError(strong.StatusBadRequest)
			}
			return v(ctx, id)
		}
	}

	handler, err := httpcontext.Compose(handlers)
	if err != nil {
		return nil, "", err
	}

	return handler, param, nil
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
	if len(handlers) == 0 {
		return r
	}
	handler, param, err := normalizeHandlers(handlers, r.name)
	if err != nil {
		panic(err)
	}

	return r.method(restCreate, Route{
		Path:    fmt.Sprintf("/:%s", param),
		Method:  strong.PUT,
		Handler: handler,
	})
}

func (r *Rest) Patch(handlers ...interface{}) *Rest {

	if len(handlers) == 0 {
		return r
	}
	handler, param, err := normalizeHandlers(handlers, r.name)
	if err != nil {
		panic(err)
	}

	return r.method(restCreate, Route{
		Path:    fmt.Sprintf("/:%s", param),
		Method:  strong.PATCH,
		Handler: handler,
	})
}
func (r *Rest) Delete(handlers ...interface{}) *Rest {

	if len(handlers) == 0 {
		return r
	}
	handler, param, err := normalizeHandlers(handlers, r.name)
	if err != nil {
		panic(err)
	}

	return r.method(restCreate, Route{
		Path:    fmt.Sprintf("/:%s", param),
		Method:  strong.DELETE,
		Handler: handler,
	})
}
func (r *Rest) Get(handlers ...interface{}) *Rest {

	if len(handlers) == 0 {
		return r
	}
	handler, param, err := normalizeHandlers(handlers, r.name)
	if err != nil {
		panic(err)
	}

	return r.method(restCreate, Route{
		Path:    fmt.Sprintf("/:%s", param),
		Method:  strong.GET,
		Handler: handler,
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
