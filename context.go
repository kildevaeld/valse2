package valse2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"

	"github.com/kildevaeld/strong"

	"github.com/julienschmidt/httprouter"
)

type Context struct {
	res         http.ResponseWriter
	req         *http.Request
	status      int
	params      httprouter.Params
	headersSent bool
	u           map[string]interface{}
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
	if !c.headersSent {
		status := c.status
		if status == 0 {
			status = strong.StatusOK
		}
		c.res.WriteHeader(status)
		c.headersSent = true
	}
	return c.res.Write(bs)
}

func (c *Context) PostBody() (bs []byte, err error) {
	bs, err = ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return bs, err
	}
	c.Request().Body.Close()
	return bs, err
}

func (c *Context) PostJSONBody(v interface{}) error {
	bs, err := c.PostBody()
	if err != nil {
		return err
	}
	return json.Unmarshal(bs, v)
}

func (c *Context) SetUserValue(k string, v interface{}) *Context {
	c.u[k] = v
	return c
}

func (c *Context) UserValue(k string) interface{} {
	return c.u[k]
}

type Link struct {
	Last    int
	First   int
	Current int
	Path    string
}

var reg = regexp.MustCompile("https?:.*")

const loverheader = 7

func writelink(rel string, url *url.URL) []byte {
	buf := bytes.NewBuffer(nil)
	buf.WriteString("<")
	buf.WriteString(url.String())
	buf.WriteString(`>; rel="` + rel + `"`)

	return buf.Bytes()
}

func (c *Context) SetLinkHeader(l Link) *Context {

	url, err := url.Parse(c.Request().URL.String())
	if err != nil {
		panic(err)
	}

	if l.Path != "" {
		url.Path = l.Path
	}

	var links [][]byte
	var page = "page"
	args := c.Request().URL.Query()

	args.Set(page, fmt.Sprintf("%d", l.First))
	links = append(links, writelink("first", url))

	args.Set(page, fmt.Sprintf("%d", l.Current))
	links = append(links, writelink("current", url))

	if l.Last > l.Current {
		args.Set(page, fmt.Sprintf("%d", l.Current+1))

		links = append(links, writelink("next", url))
	}
	if l.Current > l.First {
		args.Set(page, fmt.Sprintf("%d", l.Current-1))

		links = append(links, writelink("prev", url))
	}
	args.Set(page, fmt.Sprintf("%d", l.Last))
	url.RawQuery = args.Encode()
	links = append(links, writelink("last", url))

	c.Header().Set("Link", string(bytes.Join(links, []byte(", "))))
	return c
}

func (c *Context) reset() *Context {
	c.res = nil
	c.req = nil
	c.params = nil
	c.status = strong.StatusOK
	c.u = map[string]interface{}{}
	return c
}
