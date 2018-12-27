package jwt

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"

	jwt "github.com/dgrijalva/jwt-go"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/kildevaeld/strong"
	"github.com/kildevaeld/valse2/httpcontext"
)

// Shamefully stolen from the echo framework https://github.com/labstack/echo

type (
	TokenLookup struct {
		Query  string `json:"query"`
		Header string `json:"header"`
		Cookie string `json:"cookie"`
	}

	// JWTConfig defines the config for JWT middleware.
	JWTConfig struct {
		// Skipper defines a function to skip middleware.
		//Skipper middleware.Skipper

		// Signing key to validate token.
		// Required.
		SigningKey interface{} `json:"signing_key"`

		// Signing method, used to check token signing method.
		// Optional. Default value HS256.
		SigningMethod string `json:"signing_method"`

		// Context key to store user information from the token into context.
		// Optional. Default value "user".
		ContextKey string `json:"context_key"`

		// Claims are extendable claims data defining token content.
		// Optional. Default value jwt.MapClaims
		Claims jwt.Claims

		// TokenLookup is a string in the form of "<source>:<name>" that is used
		// to extract token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		// - "cookie:<name>"
		TokenLookup *TokenLookup `json:"token_lookup"`

		keyFunc jwt.Keyfunc
	}

	jwtExtractor func(*httpcontext.Context) (string, error)
)

const (
	bearer = "Bearer"
)

// Algorithims
const (
	AlgorithmHS256 = "HS256"
)

var (
	// DefaultJWTConfig is the default JWT auth middleware config.
	DefaultJWTConfig = JWTConfig{
		/*Skipper: func(c *fasthttp.RequestCtx) bool {
			return false
		},*/
		SigningMethod: AlgorithmHS256,
		ContextKey:    "user",
		TokenLookup: &TokenLookup{
			Header: strong.HeaderAuthorization,
		},
		Claims: jwt.MapClaims{},
	}
)

// JWT returns a JSON Web Token (JWT) auth middleware.
//
// For valid token, it sets the user in context and calls next handler.
// For invalid token, it returns "401 - Unauthorized" error.
// For empty token, it returns "400 - Bad Request" error.
//
// See: https://jwt.io/introduction
// See `JWTConfig.TokenLookup`
func JWT(key []byte) httpcontext.MiddlewareHandler {
	c := DefaultJWTConfig
	c.SigningKey = key
	return JWTWithConfig(c)
}

// JWTWithConfig returns a JWT auth middleware with config.
// See: `JWT()`.
func JWTWithConfig(config JWTConfig) httpcontext.MiddlewareHandler {
	// Defaults
	/*if config.Skipper == nil {
		config.Skipper = DefaultJWTConfig.Skipper
	}*/
	if config.SigningKey == nil {
		panic("jwt middleware requires signing key")
	}
	if config.SigningMethod == "" {
		config.SigningMethod = DefaultJWTConfig.SigningMethod
	}
	if config.ContextKey == "" {
		config.ContextKey = DefaultJWTConfig.ContextKey
	}
	if config.Claims == nil {
		config.Claims = DefaultJWTConfig.Claims
	}
	if config.TokenLookup == nil {
		config.TokenLookup = DefaultJWTConfig.TokenLookup
	}
	config.keyFunc = func(t *jwt.Token) (interface{}, error) {
		// Check the signing method
		if t.Method.Alg() != config.SigningMethod {
			return nil, fmt.Errorf("unexpected jwt signing method=%v", t.Header["alg"])
		}
		return config.SigningKey, nil
	}

	// Initialize

	var extractors jwtExtractors

	if config.TokenLookup.Header != "" {
		extractors = append(extractors, jwtFromHeader(config.TokenLookup.Header))
	}
	if config.TokenLookup.Query != "" {
		extractors = append(extractors, jwtFromQuery(config.TokenLookup.Query))
	}
	if config.TokenLookup.Cookie != "" {
		extractors = append(extractors, jwtFromCookie(config.TokenLookup.Cookie))
	}

	return func(next httpcontext.HandlerFunc) httpcontext.HandlerFunc {
		return func(c *httpcontext.Context) error {
			/*if config.Skipper(c) {
				return next(c)
			}*/

			auth, err := extractors.fromContext(c)
			if err != nil {
				return strong.NewHTTPError(http.StatusBadRequest, "Invalid Auhtorization header")
			}
			token := new(jwt.Token)
			// Issue #647, #656
			if _, ok := config.Claims.(jwt.MapClaims); ok {
				token, err = jwt.Parse(auth, config.keyFunc)
			} else {
				claims := reflect.ValueOf(config.Claims).Interface().(jwt.Claims)
				token, err = jwt.ParseWithClaims(auth, claims, config.keyFunc)
			}
			if err == nil && token.Valid {
				// Store user information from token into context.
				c.SetUserValue(config.ContextKey, token)
				return next(c)
			}

			return strong.ErrUnauthorized
		}
	}
}

type jwtExtractors []jwtExtractor

func (jwt *jwtExtractors) fromContext(ctx *httpcontext.Context) (string, error) {

	var result error
	for _, extractor := range *jwt {
		jsw, err := extractor(ctx)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}
		return jsw, nil
	}

	return "", result
}

// jwtFromHeader returns a `jwtExtractor` that extracts token from request header.
func jwtFromHeader(header string) jwtExtractor {
	return func(c *httpcontext.Context) (string, error) {
		auth := c.Request().Header.Get(header)
		l := len(bearer)
		if len(auth) > l+1 && string(auth[:l]) == bearer {
			return string(auth[l+1:]), nil
		}
		return "", errors.New("empty or invalid jwt in request header")
	}
}

// jwtFromQuery returns a `jwtExtractor` that extracts token from query string.
func jwtFromQuery(param string) jwtExtractor {
	return func(c *httpcontext.Context) (string, error) {

		token := c.Request().URL.Query().Get(param)
		var err error
		if len(token) == 0 {
			return "", errors.New("empty jwt in query string")
		}
		return string(token), err
	}
}

/// jwtFromCookie returns a `jwtExtractor` that extracts token from named cookie.
func jwtFromCookie(name string) jwtExtractor {
	return func(c *httpcontext.Context) (string, error) {

		cookie, err := c.Request().Cookie(name)
		if err != nil {
			return "", errors.New("empty jwt in cookie")
		}

		return cookie.Value, nil
	}
}
