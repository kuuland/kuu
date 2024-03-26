package kuu

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/iamdanielyin/strcase"
	"mime/multipart"
	"strings"
)

type Context struct {
	raw       *fiber.Ctx
	app       *App
	requestId string
}

func (c *Context) Context() context.Context {
	return c.raw.Context()
}

func (c *Context) App() *App {
	return c.app
}

func (c *Context) Raw() *fiber.Ctx {
	return c.raw
}

func (c *Context) RequestId() string {
	return c.requestId
}

func (c *Context) Next() error {
	return c.raw.Next()
}

func (c *Context) ClientIP() string {
	return c.raw.IP()
}

func (c *Context) FormFile(key string) (*multipart.FileHeader, error) {
	return c.raw.FormFile(key)
}

func (c *Context) SaveFile(fh *multipart.FileHeader, path string) error {
	return c.raw.SaveFile(fh, path)
}

func (c *Context) Queries() map[string]string {
	return c.raw.Queries()
}

func (c *Context) FormValue(key string, defaultValue ...string) string {
	return c.raw.FormValue(key, defaultValue...)
}

func (c *Context) Redirect(location string, status ...int) error {
	return c.raw.Redirect(location, status...)
}

func (c *Context) RawBody() []byte {
	return c.raw.Body()
}

func (c *Context) Token() *Token {
	if v := c.raw.Context().UserValue(ContextValueTokenKey); v != nil {
		if t, ok := v.(*Token); ok {
			return t
		}
	}
	var value string
	for _, key := range tokenKeys {
		if v := c.raw.Get(key); v != "" {
			value = v
			break
		}
	}
	if value == "" {
		for _, key := range tokenKeys {
			if v := c.raw.Query(key); v != "" {
				value = v
				break
			}
		}
	}
	var token *Token
	if value != "" {
		token = NewToken(value, c.app.cache)
		c.raw.Context().SetUserValue(ContextValueTokenKey, token)
	} else {
		Warningf("token not found: queries=%s, headers=%s\n", JSONStringify(c.Queries()), JSONStringify(c.Headers()))
	}
	return token
}

type Handler = func(*Context) *Reply

func (c *Context) Headers() map[string][]string {
	return c.raw.GetReqHeaders()
}

func (c *Context) OK(msgIdAndValues ...any) *Reply {
	var id string
	if len(msgIdAndValues) > 0 {
		id = strings.TrimSpace(msgIdAndValues[0].(string))
	}
	if id == "" {
		id = "OK"
	}
	id = strcase.ToScreamingSnake(id)
	var values Map
	if len(msgIdAndValues) > 1 {
		values = msgIdAndValues[1].(Map)
	}
	return &Reply{
		Code:          0,
		Success:       true,
		RequestId:     c.requestId,
		Message:       id,
		MessageType:   MessageTypeInfo,
		MessageValues: values,
	}
}

func (c *Context) Err(data any, msgId string, msgValues ...Map) *Reply {
	return c.ErrWithCode(-1, data, msgId, msgValues...)
}

func (c *Context) ErrWithCode(code int, data any, msgId string, msgValues ...Map) *Reply {
	var values Map
	if len(msgValues) > 0 {
		values = msgValues[0]
	}
	return &Reply{
		Code:          code,
		Success:       false,
		RequestId:     c.requestId,
		Message:       msgId,
		MessageType:   MessageTypeError,
		MessageValues: values,
		Data:          data,
	}
}
