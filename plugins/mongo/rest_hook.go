package mongo

import (
	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
	"github.com/kuuland/kuu"
)

// IPreRestCreate 新增前钩子
type IPreRestCreate interface {
	PreRestCreate(*gin.Context, *[]kuu.H)
}

// IPostRestCreate 新增后钩子
type IPostRestCreate interface {
	PostRestCreate(*gin.Context, *[]kuu.H)
}

// IPreRestID ID查询前钩子
type IPreRestID interface {
	PreRestID(*gin.Context, string, *Params) string
}

// IPostRestID ID查询后钩子
type IPostRestID interface {
	PostRestID(*gin.Context, *kuu.H)
}

// IPreRestList 列表查询前钩子
type IPreRestList interface {
	PreRestList(*gin.Context, *mgo.Query, *Params)
}

// IPostRestList 列表查询后钩子
type IPostRestList interface {
	PostRestList(*gin.Context, *kuu.H)
}

// IPreRestRemove 删除前钩子
type IPreRestRemove interface {
	PreRestRemove(*gin.Context, *kuu.H, bool)
}

// IPostRestRemove 删除后钩子
type IPostRestRemove interface {
	PostRestRemove(*gin.Context, interface{})
}

// IPreRestUpdate 修改后钩子
type IPreRestUpdate interface {
	PreRestUpdate(*gin.Context, *kuu.H, *kuu.H, bool)
}

// IPostRestUpdate 修改后钩子
type IPostRestUpdate interface {
	PostRestUpdate(*gin.Context, interface{})
}
