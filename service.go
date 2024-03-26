package kuu

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/iamdanielyin/strcase"
	"os"
)

type Service[I, O any] struct {
	Name            string
	Method          string
	Description     string
	Errors          map[int]string
	IgnoreToken     bool
	IgnoreBodyParse bool
	ReturnDataOnly  bool
	Handler         func(ctx *Context, inputs *I, outputs *O) *Reply
}

func Register[I, O any](app *App, s Service[I, O]) {
	s.Name = strcase.ToSnake(s.Name)
	if s.Method == "" {
		s.Method = fiber.MethodPost
	}
	app.raw.Add(s.Method, os.Getenv(ConfigAPIPrefix)+"/"+s.Name, func(fc *fiber.Ctx) error {
		// 获取上下文
		c := fc.Context().UserValue(ContextValueKey).(*Context)
		// 白名单限制
		t := c.Token()
		if !s.IgnoreToken {
			if t == nil || !t.Valid() {
				return fc.JSON(c.ErrWithCode(401, errors.New("401 ERR_UNAUTHORIZED"), ErrUnauthorized))
			}
		}
		// 获取请求参数
		var inputs I
		if !s.IgnoreBodyParse {
			if err := fc.BodyParser(&inputs); err != nil {
				Errorln(err)
				return fc.JSON(c.Err(err, ErrParsingFailed))
			}
			if err := validator.New().Struct(&inputs); err != nil {
				Errorln(err)
				return fc.JSON(c.Err(err, ErrValidationFailed))
			}
		}
		// 处理响应参数
		var outputs O
		reply := s.Handler(c, &inputs, &outputs)
		if reply == nil {
			reply = c.OK()
		}
		if reply.Data == nil {
			reply.Data = outputs
		}
		if v, ok := reply.Data.(error); ok {
			Errorln(v)
			reply.Data = nil
		}
		if s.ReturnDataOnly {
			return fc.JSON(reply.Data)
		}
		return fc.JSON(reply)
	})
}
