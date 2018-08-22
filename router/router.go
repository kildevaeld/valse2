// Copyright 2013 Julien Schmidt. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

// Package httprouter is a trie based high performance HTTP request router.
//
// A trivial example is:
//
//  package main
//
//  import (
//      "fmt"
//      "github.com/julienschmidt/httprouter"
//      "net/http"
//      "log"
//  )
//
//  func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
//      fmt.Fprint(w, "Welcome!\n")
//  }
//
//  func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
//      fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
//  }
//
//  func main() {
//      router := httprouter.New()
//      router.GET("/", Index)
//      router.GET("/hello/:name", Hello)
//
//      log.Fatal(http.ListenAndServe(":8080", router))
//  }
//
// The router matches incoming requests by the request method and the path.
// If a handle is registered for this path and method, the router delegates the
// request to that function.
// For the methods GET, POST, PUT, PATCH and DELETE shortcut functions exist to
// register handles, for all other methods router.Handle can be used.
//
// The registered path, against which the router matches incoming requests, can
// contain two types of parameters:
//  Syntax    Type
//  :name     named parameter
//  *name     catch-all parameter
//
// Named parameters are dynamic path segments. They match anything until the
// next '/' or the path end:
//  Path: /blog/:category/:post
//
//  Requests:
//   /blog/go/request-routers            match: category="go", post="request-routers"
//   /blog/go/request-routers/           no match, but the router would redirect
//   /blog/go/                           no match
//   /blog/go/request-routers/comments   no match
//
// Catch-all parameters match anything until the path end, including the
// directory index (the '/' before the catch-all). Since they match anything
// until the end, catch-all parameters must always be the final path element.
//  Path: /files/*filepath
//
//  Requests:
//   /files/                             match: filepath="/"
//   /files/LICENSE                      match: filepath="/LICENSE"
//   /files/templates/article.html       match: filepath="/templates/article.html"
//   /files                              no match, but the router would redirect
//
// The value of parameters is saved as a slice of the Param struct, consisting
// each of a key and a value. The slice is passed to the Handle func as a third
// parameter.
// There are two ways to retrieve the value of a parameter:
//  // by the name of the parameter
//  user := ps.ByName("user") // defined by :user or *user
//
//  // by the index of the parameter. This way you can also get the name (key)
//  thirdKey   := ps[2].Key   // the name of the 3rd parameter
//  thirdValue := ps[2].Value // the value of the 3rd parameter
package router

import (
	"net/http"

	"github.com/kildevaeld/strong"

	"github.com/kildevaeld/valse2/httpcontext"
)

// // Handle is a function that can be registered to a route to handle HTTP
// // requests. Like http.HandlerFunc, but has a third parameter for the values of
// // wildcards (variables).
// type Handle func(ctx *httpcontext.Context) error

// // Param is a single URL parameter, consisting of a key and a value.
// type Param struct {
// 	Key   string
// 	Value string
// }

// // Params is a Param-slice, as returned by the router.
// // The slice is ordered, the first URL parameter is also the first slice value.
// // It is therefore safe to read values by the index.
// type Params []Param

// ByName returns the value of the first Param which key matches the given name.
// If no matching Param is found, an empty string is returned.
// func (ps Params) ByName(name string) string {
// 	for i := range ps {
// 		if ps[i].Key == name {
// 			return ps[i].Value
// 		}
// 	}
// 	return ""
// }

// Router is a http.Handler which can be used to dispatch requests to different
// handler functions via configurable routes
type Router struct {
	trees map[string]*node

	// Enables automatic redirection if the current route can't be matched but a
	// handler for the path with (without) the trailing slash exists.
	// For example if /foo/ is requested but a route only exists for /foo, the
	// client is redirected to /foo with http status code 301 for GET requests
	// and 307 for all other request methods.
	RedirectTrailingSlash bool

	// If enabled, the router tries to fix the current request path, if no
	// handle is registered for it.
	// First superfluous path elements like ../ or // are removed.
	// Afterwards the router does a case-insensitive lookup of the cleaned path.
	// If a handle can be found for this route, the router makes a redirection
	// to the corrected path with status code 301 for GET requests and 307 for
	// all other request methods.
	// For example /FOO and /..//Foo could be redirected to /foo.
	// RedirectTrailingSlash is independent of this option.
	RedirectFixedPath bool

	// If enabled, the router checks if another method is allowed for the
	// current route, if the current request can not be routed.
	// If this is the case, the request is answered with 'Method Not Allowed'
	// and HTTP status code 405.
	// If no other Method is allowed, the request is delegated to the NotFound
	// handler.
	HandleMethodNotAllowed bool

	// If enabled, the router automatically replies to OPTIONS requests.
	// Custom OPTIONS handlers take priority over automatic replies.
	HandleOPTIONS bool

	// Configurable http.Handler which is called when no matching route is
	// found. If it is not set, http.NotFound is used.
	NotFound http.Handler

	// Configurable http.Handler which is called when a request
	// cannot be routed and HandleMethodNotAllowed is true.
	// If it is not set, http.Error with http.StatusMethodNotAllowed is used.
	// The "Allow" header with allowed request methods is set before the handler
	// is called.
	MethodNotAllowed http.Handler

	// Function to handle panics recovered from http handlers.
	// It should be used to generate a error page and return the http error code
	// 500 (Internal Server Error).
	// The handler can be used to keep your server from crashing because of
	// unrecovered panics.
	PanicHandler func(*httpcontext.Context, interface{})
}

// Make sure the Router conforms with the http.Handler interface
var _ http.Handler = New()

// New returns a new initialized Router.
// Path auto-correction, including trailing slashes, is enabled by default.
func New() *Router {
	return &Router{
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      true,
		HandleMethodNotAllowed: true,
		HandleOPTIONS:          true,
	}
}

// GET is a shortcut for router.Handle("GET", path, handle)
func (r *Router) GET(path string, handle httpcontext.HandlerFunc) {
	r.Handle("GET", path, handle)
}

// HEAD is a shortcut for router.Handle("HEAD", path, handle)
func (r *Router) HEAD(path string, handle httpcontext.HandlerFunc) {
	r.Handle("HEAD", path, handle)
}

// OPTIONS is a shortcut for router.Handle("OPTIONS", path, handle)
func (r *Router) OPTIONS(path string, handle httpcontext.HandlerFunc) {
	r.Handle("OPTIONS", path, handle)
}

// POST is a shortcut for router.Handle("POST", path, handle)
func (r *Router) POST(path string, handle httpcontext.HandlerFunc) {
	r.Handle("POST", path, handle)
}

// PUT is a shortcut for router.Handle("PUT", path, handle)
func (r *Router) PUT(path string, handle httpcontext.HandlerFunc) {
	r.Handle("PUT", path, handle)
}

// PATCH is a shortcut for router.Handle("PATCH", path, handle)
func (r *Router) PATCH(path string, handle httpcontext.HandlerFunc) {
	r.Handle("PATCH", path, handle)
}

// DELETE is a shortcut for router.Handle("DELETE", path, handle)
func (r *Router) DELETE(path string, handle httpcontext.HandlerFunc) {
	r.Handle("DELETE", path, handle)
}

// Handle registers a new request handle with the given path and method.
//
// For GET, POST, PUT, PATCH and DELETE requests the respective shortcut
// functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
func (r *Router) Handle(method, path string, handle httpcontext.HandlerFunc) {
	if path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}

	if r.trees == nil {
		r.trees = make(map[string]*node)
	}

	root := r.trees[method]
	if root == nil {
		root = new(node)
		r.trees[method] = root
	}

	root.addRoute(path, handle)
}

// Handler is an adapter which allows the usage of an http.Handler as a
// request handle.
// func (r *Router) Handler(method, path string, handler http.Handler) {
// 	r.Handle(method, path,
// 		func(ctx *httpcontext.Context) {
// 			handler.ServeHTTP(w, req)
// 		},
// 	)
// }

// // HandlerFunc is an adapter which allows the usage of an http.HandlerFunc as a
// // request handle.
// func (r *Router) HandlerFunc(method, path string, handler http.HandlerFunc) {
// 	r.Handler(method, path, handler)
// }

// ServeFiles serves files from the given file system root.
// The path must end with "/*filepath", files are then served from the local
// path /defined/root/dir/*filepath.
// For example if root is "/etc" and *filepath is "passwd", the local file
// "/etc/passwd" would be served.
// Internally a http.FileServer is used, therefore http.NotFound is used instead
// of the Router's NotFound handler.
// To use the operating system's file system implementation,
// use http.Dir:
//     router.ServeFiles("/src/*filepath", http.Dir("/var/www"))
func (r *Router) ServeFiles(path string, root http.FileSystem, middlewares ...interface{}) error {
	if len(path) < 10 || path[len(path)-10:] != "/*filepath" {
		panic("path must end with /*filepath in path '" + path + "'")
	}

	fileServer := http.FileServer(root)

	handler := func(ctx *httpcontext.Context) error {
		ctx.Request().URL.Path = ctx.Params().ByName("filepath")
		fileServer.ServeHTTP(ctx.Response(), ctx.Request())
		return httpcontext.ErrHandled
	}

	middlewares = append(middlewares, handler)

	handler, err := httpcontext.Compose(middlewares)

	if err != nil {
		return err
	}

	r.GET(path, handler)

	return nil
}

func (r *Router) recv(ctx *httpcontext.Context) {
	if rcv := recover(); rcv != nil {
		r.PanicHandler(ctx, rcv)
	}
}

// Lookup allows the manual lookup of a method + path combo.
// This is e.g. useful to build a framework around this router.
// If the path was found, it returns the handle function and the path parameter
// values. Otherwise the third return value indicates whether a redirection to
// the same path with an extra / without the trailing slash should be performed.
func (r *Router) Lookup(method, path string) (httpcontext.HandlerFunc, httpcontext.Params, bool) {
	if root := r.trees[method]; root != nil {
		return root.getValue(path)
	}
	return nil, nil, false
}

func (r *Router) allowed(path, reqMethod string) (allow string) {
	if path == "*" { // server-wide
		for method := range r.trees {
			if method == "OPTIONS" {
				continue
			}

			// add request method to list of allowed methods
			if len(allow) == 0 {
				allow = method
			} else {
				allow += ", " + method
			}
		}
	} else { // specific path
		for method := range r.trees {
			// Skip the requested method - we already tried this one
			if method == reqMethod || method == "OPTIONS" {
				continue
			}

			handle, _, _ := r.trees[method].getValue(path)
			if handle != nil {
				// add request method to list of allowed methods
				if len(allow) == 0 {
					allow = method
				} else {
					allow += ", " + method
				}
			}
		}
	}
	if len(allow) > 0 {
		allow += ", OPTIONS"
	}
	return
}

func (r *Router) ServeHTTPContext(ctx *httpcontext.Context) error {
	if r.PanicHandler != nil {
		defer r.recv(ctx)
	}

	req := ctx.Request()

	path := req.URL.Path

	if root := r.trees[req.Method]; root != nil {
		if handle, ps, tsr := root.getValue(path); handle != nil {
			ctx.SetParams(ps)
			return handle(ctx)

		} else if req.Method != "CONNECT" && path != "/" {
			code := 301 // Permanent redirect, request with GET method
			if req.Method != "GET" {
				// Temporary redirect, request with same method
				// As of Go 1.3, Go does not support status code 308.
				code = 307
			}

			if tsr && r.RedirectTrailingSlash {
				if len(path) > 1 && path[len(path)-1] == '/' {
					req.URL.Path = path[:len(path)-1]
				} else {
					req.URL.Path = path + "/"
				}
				//http.Redirect(w, req, req.URL.String(), code)
				return ctx.Redirect(code, req.URL.String())
			}

			// Try to fix the request path
			if r.RedirectFixedPath {
				fixedPath, found := root.findCaseInsensitivePath(
					CleanPath(path),
					r.RedirectTrailingSlash,
				)
				if found {
					req.URL.Path = string(fixedPath)
					//http.Redirect(w, req, req.URL.String(), code)
					return ctx.Redirect(code, req.URL.String())
				}
			}
		}
	}

	if req.Method == "OPTIONS" {
		// Handle OPTIONS requests
		if r.HandleOPTIONS {
			if allow := r.allowed(path, req.Method); len(allow) > 0 {
				ctx.Header().Set("Allow", allow)
				return nil
			}
		}
	} else {
		// Handle 405
		if r.HandleMethodNotAllowed {
			if allow := r.allowed(path, req.Method); len(allow) > 0 {
				ctx.Header().Set("Allow", allow)
				if r.MethodNotAllowed != nil {
					//r.MethodNotAllowed.ServeHTTP(w, req)
				} else {
					// http.Error(w,
					// 	http.StatusText(http.StatusMethodNotAllowed),
					// 	http.StatusMethodNotAllowed,
					// )
					return ctx.Error(strong.StatusMethodNotAllowed)
				}
				return nil
			}
		}
	}

	// Handle 404
	// if r.NotFound != nil {
	// 	r.NotFound.ServeHTTP(w, req)
	// } else {
	// 	http.NotFound(w, req)
	// }
	return strong.ErrNotFound
}

// ServeHTTP makes the router implement the http.Handler interface.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	ctx := httpcontext.Acquire(w, req)
	defer httpcontext.Release(ctx)

	if err := r.ServeHTTPContext(ctx); err != nil {
		panic(err)
	}

	// if r.PanicHandler != nil {
	// 	defer r.recv(w, req)
	// }

	// path := req.URL.Path

	// if root := r.trees[req.Method]; root != nil {
	// 	if handle, ps, tsr := root.getValue(path); handle != nil {
	// 		handle(w, req, ps)
	// 		return
	// 	} else if req.Method != "CONNECT" && path != "/" {
	// 		code := 301 // Permanent redirect, request with GET method
	// 		if req.Method != "GET" {
	// 			// Temporary redirect, request with same method
	// 			// As of Go 1.3, Go does not support status code 308.
	// 			code = 307
	// 		}

	// 		if tsr && r.RedirectTrailingSlash {
	// 			if len(path) > 1 && path[len(path)-1] == '/' {
	// 				req.URL.Path = path[:len(path)-1]
	// 			} else {
	// 				req.URL.Path = path + "/"
	// 			}
	// 			http.Redirect(w, req, req.URL.String(), code)
	// 			return
	// 		}

	// 		// Try to fix the request path
	// 		if r.RedirectFixedPath {
	// 			fixedPath, found := root.findCaseInsensitivePath(
	// 				CleanPath(path),
	// 				r.RedirectTrailingSlash,
	// 			)
	// 			if found {
	// 				req.URL.Path = string(fixedPath)
	// 				http.Redirect(w, req, req.URL.String(), code)
	// 				return
	// 			}
	// 		}
	// 	}
	// }

	// if req.Method == "OPTIONS" {
	// 	// Handle OPTIONS requests
	// 	if r.HandleOPTIONS {
	// 		if allow := r.allowed(path, req.Method); len(allow) > 0 {
	// 			w.Header().Set("Allow", allow)
	// 			return
	// 		}
	// 	}
	// } else {
	// 	// Handle 405
	// 	if r.HandleMethodNotAllowed {
	// 		if allow := r.allowed(path, req.Method); len(allow) > 0 {
	// 			w.Header().Set("Allow", allow)
	// 			if r.MethodNotAllowed != nil {
	// 				r.MethodNotAllowed.ServeHTTP(w, req)
	// 			} else {
	// 				http.Error(w,
	// 					http.StatusText(http.StatusMethodNotAllowed),
	// 					http.StatusMethodNotAllowed,
	// 				)
	// 			}
	// 			return
	// 		}
	// 	}
	// }

	// // Handle 404
	// if r.NotFound != nil {
	// 	r.NotFound.ServeHTTP(w, req)
	// } else {
	// 	http.NotFound(w, req)
	// }
}
