package valse2

import "github.com/kildevaeld/valse2/httpcontext"

type Route struct {
	Method  string
	Path    string
	Handler httpcontext.HandlerFunc
}
