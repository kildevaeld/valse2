package cache

import (
	"fmt"
	"strings"
	"time"

	"github.com/kildevaeld/strong"
	"github.com/kildevaeld/valse2/httpcontext"
)

type CacheControl struct {
	MaxAge  int
	Private bool
	Debug   bool
}

func NewCacheControl(options *CacheControl) httpcontext.MiddlewareHandler {
	if options == nil {
		options = &CacheControl{
			MaxAge:  7 * 24 * 60 * 60,
			Private: false,
			Debug:   false,
		}
	}
	maxAge := options.MaxAge
	if options.Debug {
		maxAge = 1
	}

	return func(next httpcontext.HandlerFunc) httpcontext.HandlerFunc {
		return func(ctx *httpcontext.Context) error {

			if err := next(ctx); err != nil {
				return err
			} else if !(strong.IsSuccess(ctx.StatusCode()) || ctx.StatusCode() == 0) {
				return nil
			}

			scope := "public"
			if options.Private {
				scope = "private"
			}

			cacheCtrl := ctx.Request().Header.Get(strong.HeaderCacheControl)
			if cacheCtrl != "" && strings.ToLower(cacheCtrl) == "no-cache" {
				return nil
			}

			ctx.Header().Set(strong.HeaderCacheControl, fmt.Sprintf(scope+", max-age=%d", maxAge))

			now := time.Now().Add(time.Duration(maxAge) * time.Second)

			ctx.Header().Set("expires", now.Format(time.RFC1123))

			return nil
		}
	}
}
