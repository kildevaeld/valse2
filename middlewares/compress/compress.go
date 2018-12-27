package gzip

import (
	"github.com/kildevaeld/valse2/httpcontext"
)

func GzipWithConfig() httpcontext.MiddlewareHandler {
	return func(next httpcontext.HandlerFunc) httpcontext.HandlerFunc {
		return func(c *httpcontext.Context) error {
			return nil
		}
	}
}
