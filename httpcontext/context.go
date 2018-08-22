package httpcontext

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/kildevaeld/strong"
)

var (
	requestPool sync.Pool
	contextPool sync.Pool
)

func init() {
	requestPool = sync.Pool{
		New: func() interface{} {
			return &RequestBody{}
		},
	}

	contextPool = sync.Pool{
		New: func() interface{} {
			return &Context{}
		},
	}
}

type RequestBody struct {
	reader      io.ReadCloser
	contentType string
	done        bool
}

func (r *RequestBody) Read(bs []byte) (int, error) {
	if r.done {
		return 0, io.EOF
	}
	read, err := r.reader.Read(bs)
	if err == io.EOF {
		r.done = true
	}
	return read, err
}

func (r *RequestBody) Close() error {
	r.done = true
	return r.reader.Close()
}

func (r *RequestBody) ReadAll() ([]byte, error) {
	return ioutil.ReadAll(r.reader)
}

func (r *RequestBody) Decode(v interface{}) error {
	if r.done {
		return io.EOF
	}

	bs, err := r.ReadAll()
	defer r.Close()
	if err != nil {
		return err
	}

	decoder := GetDecoder(r.contentType)
	if decoder == nil {
		return fmt.Errorf("cannot decode content type '%s'", r.contentType)
	}

	return decoder.Decode(bs, v)

	// switch r.contentType {
	// case strong.MIMEApplicationJSON:
	// 	return json.Unmarshal(bs, v)
	// default:
	// 	return fmt.Errorf("cannot decode content type '%s'", r.contentType)
	// }

}

func (r *RequestBody) reset() *RequestBody {
	r.done = false
	r.reader = nil
	return r
}

type Context struct {
	req      *http.Request
	req_body *RequestBody
	params   Params
	res      http.ResponseWriter

	body   io.ReadCloser
	status int
	u      map[string]interface{}
}

func (c *Context) Params() Params {
	return c.params
}

func (c *Context) SetParams(params Params) {
	c.params = params
}

func (c *Context) Request() *http.Request {
	return c.req
}

func (c *Context) Response() http.ResponseWriter {
	return c.res
}

// Response
func (c *Context) SetContentType(contentType string) *Context {
	c.res.Header().Set(strong.HeaderContentType, contentType)
	return c
}
func (c *Context) SetBody(v io.ReadCloser) *Context {
	if c.body != nil {
		c.body.Close()
	}
	c.body = v
	return c
}

func (c *Context) Body() io.Reader {
	return c.body
}

func (c *Context) SetStatusCode(status int) *Context {
	c.status = status
	return c
}

func (c *Context) StatusCode() int {
	return c.status
}

func (c *Context) RequestBody() *RequestBody {
	if c.req_body == nil {
		c.req_body = requestPool.Get().(*RequestBody)
		c.req_body.reader = c.req.Body
		c.req_body.contentType = c.req.Header.Get(strong.HeaderContentType)
	}
	return c.req_body
}

func (c *Context) Text(str string) error {
	c.res.Header().Set(strong.HeaderContentType, strong.MIMETextPlain)
	return c.bytes([]byte(str))
}

func (c *Context) JSON(v interface{}) error {
	c.res.Header().Set(strong.HeaderContentType, strong.MIMETextPlain)

	bs, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return c.bytes(bs)
}

func (c *Context) HTML(str string) error {
	c.res.Header().Set(strong.HeaderContentType, strong.MIMETextHTMLCharsetUTF8)
	return c.bytes([]byte(str))
}

func (c *Context) Error(status int, msg ...interface{}) error {
	return strong.NewHTTPError(status, msg)
}

func (c *Context) Redirect(status int, path string) error {
	return &RedirectError{status, path}
}

func (c *Context) SetUserValue(k string, v interface{}) *Context {
	if c.u == nil {
		c.u = make(map[string]interface{})
	}
	c.u[k] = v
	return c
}

func (c *Context) UserValue(k string) interface{} {
	if c.u == nil {
		return nil
	}
	return c.u[k]
}

func (c *Context) Header() http.Header {
	return c.res.Header()
}

func (c *Context) bytes(bs []byte) error {
	if c.body != nil {
		c.body.Close()
	}
	c.body = ioutil.NopCloser(bytes.NewBuffer(bs))
	return nil
}

func (c *Context) Websocket(upgrader *websocket.Upgrader) (*websocket.Conn, error) {
	if upgrader == nil {
		return websocket.Upgrade(c.res, c.req, c.Header(), 1024, 1024)
	}
	return upgrader.Upgrade(c.res, c.req, c.Header())
}

func (c *Context) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijack, ok := c.res.(http.Hijacker)
	if ok {
		return nil, nil, http.ErrNotSupported
	}

	return hijack.Hijack()
}

func (c *Context) reset() *Context {
	c.req = nil
	if c.req_body != nil {
		c.req_body.Close()
		requestPool.Put(c.req_body.reset())
	}
	c.req_body = nil
	c.res = nil
	c.params = nil
	c.u = nil
	return c
}

func Acquire(w http.ResponseWriter, r *http.Request) *Context {
	ctx := contextPool.Get().(*Context)
	ctx.res = w
	ctx.req = r
	return ctx
}

func Release(ctx *Context) {
	contextPool.Put(ctx.reset())
}
