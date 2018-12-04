package router

import (
	"github.com/kildevaeld/strong"
)

var (
	ErrNotFound = strong.NewHTTPError(strong.StatusNotFound)
)
