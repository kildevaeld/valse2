package valse2

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/kildevaeld/strong"

	"github.com/julienschmidt/httprouter"
)

type Context struct {
	res          http.ResponseWriter
	req          *http.Request
	status       int
	params       httprouter.Params
	headers_sent bool
	u            map[string]interface{}
}

func (c *Context) Request() *http.Request {
	return c.req
}

func (c *Context) Header() http.Header {
	return c.res.Header()
}

func (c *Context) Params() httprouter.Params {
	return c.params
}

func (c *Context) Error(errs string, status int) error {
	c.status = status
	_, err := c.Write([]byte(errs))
	return err
}

func (c *Context) SetStatusCode(status int) {
	c.status = status
}

func (c *Context) StatusCode() int {
	return c.status
}

func (c *Context) Text(str string) error {
	c.res.Header().Add(strong.HeaderContentType, strong.MIMETextPlain)
	_, err := c.Write([]byte(str))
	return err
}

func (c *Context) JSON(v interface{}) error {
	c.res.Header().Add(strong.HeaderContentType, strong.MIMETextPlain)

	bs, err := json.Marshal(v)
	if err != nil {
		return err
	}

	_, err = c.Write(bs)
	return nil
}

func (c *Context) HTML(str string) error {
	c.res.Header().Add(strong.HeaderContentType, strong.MIMETextHTMLCharsetUTF8)
	_, err := c.Write([]byte(str))
	return err
}

func (c *Context) Write(bs []byte) (int, error) {
	if !c.headers_sent {
		status := c.status
		if status == 0 {
			status = strong.StatusOK
		}
		c.res.WriteHeader(status)
		c.headers_sent = true
	}
	return c.res.Write(bs)
}

func (c *Context) PostBody() (bs []byte, err error) {
	bs, err =  ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return bs, err
	}
	c.Request().Body.Close();
	return bs, err
}

func (c *Context) SetUserValue(k string, v interface{}) *Context {
	c.u[k] = v
	return c
}

func (c *Context) UserValue(k string) interface{} {
	return c.u[k]
}

func (c *Context) reset() *Context {
	c.res = nil
	c.req = nil
	c.params = nil
	c.status = strong.StatusOK
	c.u = map[string]interface{}{}
	return c
}
