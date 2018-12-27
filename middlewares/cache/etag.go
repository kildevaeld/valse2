package cache

import (
	"crypto/md5"
	"fmt"
	"io"
	"time"

	"github.com/kildevaeld/strong"

	"github.com/kildevaeld/valse2/httpcontext"
)

type CacheSetOptions struct {
	MaxAge time.Duration
}

type Cache interface {
	Set(key string, value []byte)
	Get(key string) []byte
	Has(key string) bool
}

type Etag struct {
	Weak      bool
	Cache     Cache
	KeyPrefix string
}

func calcEtag(reader io.ReadCloser, options *Etag) (string, error) {

	hash := md5.New()
	defer reader.Close()

	if _, err := io.Copy(hash, reader); err != nil {
		return "", err
	}

	return fmt.Sprintf("\"%x\"", hash.Sum(nil)), nil
}

func readResponse(ctx *httpcontext.Context, options Etag) (string, error) {
	key := ctx.Request().RequestURI
	if options.KeyPrefix != "" {
		key = options.KeyPrefix + key
	}

	etag := options.Cache.Get(key)
	if etag != nil {

	}

	ctx.Header().Set(strong.HeaderETag, string(etag))

	return "", nil
}

func NewEtag(options *Etag) httpcontext.MiddlewareHandler {
	if options == nil {
		options = &Etag{
			Weak: false,
		}
	}

	return func(next httpcontext.HandlerFunc) httpcontext.HandlerFunc {
		return func(ctx *httpcontext.Context) error {
			key := ctx.Request().RequestURI
			if options.KeyPrefix != "" {
				key = options.KeyPrefix + key
			}

			etag := options.Cache.Get(key)
			if etag != nil {
				if err := next(ctx); err != nil {
					return err
				}

				body := ctx.Body()
				if body == nil {
					return nil
				}

			}

			ctx.Header().Set(strong.HeaderETag, string(etag))

			// if !options.Cache.Has(url) {
			// 	etag, err := calcEtag(ctx)
			// 	if err != nil {
			// 		return err
			// 	}
			// } else {
			// 	found := ctx.Request().Header.Get(strong.HeaderIfNoneMatch)
			// 	etag := options.Cache.Get(url)
			// 	if found != "" && options.Cache.Has(url) && string(etag) == found {
			// 		ctx.SetStatusCode(strong.StatusNotModified)
			// 		return nil
			// 	}
			// }

			// if err := next(ctx); err != nil {
			// 	return err
			// }

			return nil
		}
	}
}
