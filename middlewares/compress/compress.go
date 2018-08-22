package gzip

import (
	"github.com/kildevaeld/valse2"
)

func GzipWithConfig() valse2.MiddlewareHandler {
	return func(next valse2.RequestHandler) valse2.RequestHandler {
		return func(c *valse2.Context) error {
			return nil
		}
	}
}
