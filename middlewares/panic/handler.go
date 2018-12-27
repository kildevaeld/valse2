package panic

import (
	"fmt"

	"github.com/kildevaeld/valse2/httpcontext"
)

func New() httpcontext.MiddlewareHandler {
	return func(next httpcontext.HandlerFunc) httpcontext.HandlerFunc {
		return func(ctx *httpcontext.Context) (err error) {
			defer func() {
				if e := recover(); e != nil {
					if errerr, ok := e.(error); ok {
						err = errerr
					} else {
						err = fmt.Errorf("%s", e)
					}
				}
			}()
			err = next(ctx)
			return err
		}
	}
}
