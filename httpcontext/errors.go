package httpcontext

import (
	"errors"
)

type RedirectError struct {
	status int
	url    string
}

func (r *RedirectError) Error() string {
	return r.url
}

func newRedirect(status int, err string) error {
	return &RedirectError{status, err}
}

var (
	ErrHandled = errors.New("already handled")
)
