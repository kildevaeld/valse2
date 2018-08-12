package valse2

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/kildevaeld/strong"
	"github.com/kildevaeld/valse2/httpcontext"
	"github.com/kildevaeld/valse2/router"
	"go.uber.org/zap"
)

type Options struct {
	Debug bool
}

type Valse struct {
	noCopy
	router    *router.Router
	listening bool

	m []httpcontext.MiddlewareHandler
	s *http.Server

	h httpcontext.HandlerFunc
	o *Options
	//links LinksFactory
}

func New() *Valse {
	return NewWithOptions(nil)
}

func NewWithOptions(o *Options) *Valse {
	if o == nil {
		o = &Options{}
	}
	v := &Valse{
		s:      &http.Server{},
		router: router.New(),
		o:      o,
	}

	v.s.Handler = v

	return v
}

func (v *Valse) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if v.h == nil {
		http.NotFound(w, r)
		return
	}

	if err := httpcontext.Run(w, r, v.h); err != nil {
		v.handleError(nil, w, r, err)
		return
	}

	// ctx := httpcontext.Acquire(w, r)
	// defer httpcontext.Release(ctx)

	// if err := v.h(ctx); err != nil {
	// 	v.handleError(ctx, w, r, err)
	// 	return
	// }

	// status := ctx.StatusCode()
	// hasBody := ctx.Body() != nil

	// if !hasBody && status <= 0 {
	// 	http.NotFound(w, r)
	// 	return
	// } else if hasBody && status <= 0 {
	// 	status = strong.StatusOK
	// }

	// w.WriteHeader(status)
	// if hasBody {
	// 	_, err := io.Copy(w, ctx.Body())
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }

}

func (v *Valse) Listen(addr string) error {

	if v.listening {
		return errors.New("Already running")
	}
	v.listening = true
	v.s.Addr = addr

	var err error
	if v.h, err = v.compose(); err != nil {
		return err
	}

	return v.s.ListenAndServe()

}

func (v *Valse) compose() (httpcontext.HandlerFunc, error) {
	l := len(v.m)
	handlers := make([]interface{}, l+1)
	for i, h := range v.m {
		handlers[i] = h
	}
	handlers[l] = v.router.ServeHTTPContext

	return httpcontext.Compose(handlers)
}

func (v *Valse) Close() error {
	return v.s.Close()
}

func (v *Valse) Use(handlers ...interface{}) *Valse {
	if v.listening {
		panic("cannot add middleware when running.")
	}

	for _, handler := range handlers {
		m, err := httpcontext.ToMiddlewareHandler(handler)
		if err != nil {
			panic(err)
		}
		v.m = append(v.m, m)
	}

	return v
}

func cpy(hns []httpcontext.MiddlewareHandler, r httpcontext.HandlerFunc) []interface{} {
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

func (v *Valse) Patch(path string, handlers ...interface{}) *Valse {
	return v.Route(strong.PATCH, path, handlers...)
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

	handler, err := httpcontext.Compose(handlers)

	if err != nil {
		panic(err)
	}

	if v.o.Debug {
		zap.L().Debug("register route", zap.String("method", method), zap.String("path", path))
	}
	v.router.Handle(method, path, handler)

	return v
}

func (v *Valse) handleError(ctx *httpcontext.Context, w http.ResponseWriter, r *http.Request, err error) {

	if httperr, ok := err.(*strong.HttpError); ok {
		http.Error(w, httperr.Error(), httperr.StatusCode())
	} else {
		http.Error(w, strong.StatusText(strong.StatusInternalServerError), strong.StatusInternalServerError)
	}
	fmt.Printf("handle error %#v\n", err)

}
