package valse2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/kildevaeld/strong"
)

type RequestBody struct {
	r           io.ReadCloser
	contentType string
	done        bool
}

func (r *RequestBody) Read(bs []byte) (int, error) {
	if r.done {
		return 0, io.EOF
	}
	read, err := r.r.Read(bs)
	if err == io.EOF {
		r.done = true
	}
	return read, err
}

func (r *RequestBody) Close() error {
	r.done = true
	return r.r.Close()
}

func (r *RequestBody) ReadAll() ([]byte, error) {
	return ioutil.ReadAll(r.r)
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

	switch r.contentType {
	case strong.MIMEApplicationJSON:
		return json.Unmarshal(bs, v)
	default:
		return fmt.Errorf("cannot decode content type '%s'", r.contentType)
	}

}

type Context2 struct {
	req    *http.Request
	parms  httprouter.Params
	header http.Header
	body   io.ReadCloser
	status int
	u      map[string]interface{}
}

// Response
func (c *Context2) SetContentType(contentType string) *Context2 {
	c.header.Set(strong.HeaderContentType, contentType)
	return c
}
func (c *Context2) SetBody(v io.ReadCloser) *Context2 {
	if c.body != nil {
		c.body.Close()
	}
	c.body = v
	return c
}

func (c *Context2) Body() io.Reader {
	return c.body
}

func (c *Context2) SetStatusCode(status int) *Context2 {
	c.status = status
	return c
}

func (c *Context2) StatusCode() int {
	return c.status
}

func (c *Context2) *RequestBody {
	

	return nil
} 

func (c *Context2) Text(str string) error {
	c.header.Set(strong.HeaderContentType, strong.MIMETextPlain)
	return c.bytes([]byte(str))
}

func (c *Context2) JSON(v interface{}) error {
	c.header.Set(strong.HeaderContentType, strong.MIMETextPlain)

	bs, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return c.bytes(bs)
}

func (c *Context2) HTML(str string) error {
	c.header.Set(strong.HeaderContentType, strong.MIMETextHTMLCharsetUTF8)
	return c.bytes([]byte(str))
}

func (c *Context2) SetUserValue(k string, v interface{}) *Context2 {
	c.u[k] = v
	return c
}

func (c *Context2) UserValue(k string) interface{} {
	return c.u[k]
}

func (c *Context2) bytes(bs []byte) error {
	if c.body != nil {
		c.body.Close()
	}
	c.body = ioutil.NopCloser(bytes.NewBuffer(bs))
	return nil
}
