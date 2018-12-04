package httpcontext

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

type writerWrapper struct {
	ctx    *Context
	writer *bytes.Buffer
}

func (w *writerWrapper) Write(bs []byte) (int, error) {
	return w.writer.Write(bs)
}

func (w *writerWrapper) Header() http.Header {
	return w.ctx.Header()
}

func (w *writerWrapper) WriteHeader(status int) {
	w.ctx.SetStatusCode(status)
}

func (w *writerWrapper) Close() error {
	w.ctx.SetBody(ioutil.NopCloser(w.writer))
	return nil
}

func newwriterWrapper(ctx *Context) *writerWrapper {
	return &writerWrapper{
		ctx, bytes.NewBuffer(nil),
	}
}
