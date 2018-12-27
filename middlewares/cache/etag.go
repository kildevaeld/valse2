package cache

import (
	"github.com/kildevaeld/valse2/httpcontext"
)

type Etag struct {
	Weak bool
}

func NewEtag(options *Etag) httpcontext.MiddlewareHandler {
	if options == nil {
		options = &Etag{
			Weak: false,
		}
	}

	return func(next httpcontext.HandlerFunc) httpcontext.HandlerFunc {
		return func(ctx *httpcontext.Context) error {

			return nil
		}
	}
}
