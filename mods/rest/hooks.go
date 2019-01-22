package rest

// IBeforeCreateRoute 新增路由前置钩子
type IBeforeCreateRoute interface {
	BeforeCreateRoute(*Scope) error
}

// IAfterCreateRoute 新增路由后置钩子
type IAfterCreateRoute interface {
	AfterCreateRoute(*Scope) error
}

// IBeforeIDRoute ID查询前置钩子
type IBeforeIDRoute interface {
	BeforeIDRoute(*Scope) error
}

// IAfterIDRoute ID查询后置钩子
type IAfterIDRoute interface {
	AfterIDRoute(*Scope) error
}

// IBeforeListRoute 列表查询前置钩子
type IBeforeListRoute interface {
	BeforeListRoute(*Scope) error
}

// IAfterListRoute 列表查询后置钩子
type IAfterListRoute interface {
	AfterListRoute(*Scope) error
}

// IBeforeRemoveRoute 删除路由前置钩子
type IBeforeRemoveRoute interface {
	BeforeRemoveRoute(*Scope) error
}

// IAfterRemoveRoute 删除路由后置钩子
type IAfterRemoveRoute interface {
	AfterRemoveRoute(*Scope) error
}

// IBeforeUpdateRoute 更新路由前置钩子
type IBeforeUpdateRoute interface {
	BeforeUpdateRoute(*Scope) error
}

// IAfterUpdateRoute 更新路由后置钩子
type IAfterUpdateRoute interface {
	AfterUpdateRoute(*Scope) error
}
