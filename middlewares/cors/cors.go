package cors

import (
	"strconv"
	"strings"

	"github.com/kildevaeld/strong"
	"github.com/kildevaeld/valse2"
)

// Shamefully stolen from the echo framework https://github.com/labstack/echo

type (
	// CORSConfig defines the config for CORS middleware.
	CORSConfig struct {
		// Skipper defines a function to skip middleware.
		//Skipper Skipper

		// AllowOrigin defines a list of origins that may access the resource.
		// Optional. If request header `Origin` is set, value is []string{"<Origin>"}
		// else []string{"*"}.
		AllowOrigins []string `json:"allow_origins"`

		// AllowMethods defines a list methods allowed when accessing the resource.
		// This is used in response to a preflight request.
		// Optional. Default value DefaultCORSConfig.AllowMethods.
		AllowMethods []string `json:"allow_methods"`

		// AllowHeaders defines a list of request headers that can be used when
		// making the actual request. This in response to a preflight request.
		// Optional. Default value []string{}.
		AllowHeaders []string `json:"allow_headers"`

		// AllowCredentials indicates whether or not the response to the request
		// can be exposed when the credentials flag is true. When used as part of
		// a response to a preflight request, this indicates whether or not the
		// actual request can be made using credentials.
		// Optional. Default value false.
		AllowCredentials bool `json:"allow_credentials"`

		// ExposeHeaders defines a whitelist headers that clients are allowed to
		// access.
		// Optional. Default value []string{}.
		ExposeHeaders []string `json:"expose_headers"`

		// MaxAge indicates how long (in seconds) the results of a preflight request
		// can be cached.
		// Optional. Default value 0.
		MaxAge int `json:"max_age"`
	}
)

var (
	// DefaultCORSConfig is the default CORS middleware config.
	DefaultCORSConfig = CORSConfig{
		//Skipper:      defaultSkipper,
		AllowMethods: []string{strong.GET, strong.HEAD, strong.PUT, strong.PATCH, strong.POST, strong.DELETE},
	}
)

// CORS returns a Cross-Origin Resource Sharing (CORS) middleware.
// See: https://developer.mozilla.org/en/docs/Web/HTTP/Access_control_CORS
func CORS() valse2.MiddlewareHandler {
	return CORSWithConfig(DefaultCORSConfig)
}

// CORSWithConfig returns a CORS middleware with config.
// See: `CORS()`.
func CORSWithConfig(config CORSConfig) valse2.MiddlewareHandler {
	// Defaults
	/*if config.Skipper == nil {
		config.Skipper = DefaultCORSConfig.Skipper
	}*/

	if len(config.AllowMethods) == 0 {
		config.AllowMethods = DefaultCORSConfig.AllowMethods
	}

	allowedOrigins := strings.Join(config.AllowOrigins, ",")
	allowMethods := strings.Join(config.AllowMethods, ",")
	allowHeaders := strings.Join(config.AllowHeaders, ",")
	exposeHeaders := strings.Join(config.ExposeHeaders, ",")
	maxAge := strconv.Itoa(config.MaxAge)

	return func(next valse2.RequestHandler) valse2.RequestHandler {
		return func(c *valse2.Context) error {

			//req := c.Request()
			//res := c.Response()
			origin := c.Request().Header.Get(strong.HeaderOrigin)

			if allowedOrigins == "" {
				if origin != "" {
					allowedOrigins = origin
				} else {
					if !config.AllowCredentials {
						allowedOrigins = "*"
					}
				}
			}

			// Simple request
			if c.Request().Method != strong.OPTIONS {
				c.Header().Add(strong.HeaderVary, strong.HeaderOrigin)
				c.Header().Set(strong.HeaderAccessControlAllowOrigin, allowedOrigins)
				if config.AllowCredentials {
					c.Header().Set(strong.HeaderAccessControlAllowCredentials, "true")
				}
				if exposeHeaders != "" {
					c.Header().Set(strong.HeaderAccessControlExposeHeaders, exposeHeaders)
				}
				return next(c)
			}

			// Preflight request
			c.Header().Add(strong.HeaderVary, strong.HeaderOrigin)
			c.Header().Add(strong.HeaderVary, strong.HeaderAccessControlRequestMethod)
			c.Header().Add(strong.HeaderVary, strong.HeaderAccessControlRequestHeaders)
			c.Header().Set(strong.HeaderAccessControlAllowOrigin, allowedOrigins)
			c.Header().Set(strong.HeaderAccessControlAllowMethods, allowMethods)
			if config.AllowCredentials {
				c.Header().Set(strong.HeaderAccessControlAllowCredentials, "true")
			}
			if allowHeaders != "" {
				c.Header().Set(strong.HeaderAccessControlAllowHeaders, allowHeaders)
			} else {
				h := c.Request().Header.Get(strong.HeaderAccessControlRequestHeaders)
				if string(h) != "" {
					c.Header().Set(strong.HeaderAccessControlAllowHeaders, string(h))
				}
			}
			if config.MaxAge > 0 {
				c.Header().Set(strong.HeaderAccessControlMaxAge, maxAge)
			}

			return nil
		}
	}
}
