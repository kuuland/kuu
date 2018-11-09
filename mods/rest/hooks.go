package rest

type IBeforeCreate interface {
	BeforeCreate(*Scope) error
}

type IAfterCreate interface {
	AfterCreate(*Scope) error
}

type IBeforeID interface {
	//BeforeID(*gin.Context, string, *Params) string
	BeforeID(*Scope) error
}

type IAfterID interface {
	//AfterID(*gin.Context, *kuu.H)
	AfterID(*Scope) error
}

type IBeforeList interface {
	//BeforeList(*gin.Context, interface{}, *Params)
	BeforeList(*Scope) error
}

type IAfterList interface {
	//AfterList(*gin.Context, *kuu.H)
	AfterList(*Scope) error
}

type IBeforeRemove interface {
	//BeforeRemove(*gin.Context, *kuu.H, bool)
	BeforeRemove(*Scope) error
}

type IAfterRemove interface {
	//AfterRemove(*gin.Context, interface{})
	AfterRemove(*Scope) error
}

type IBeforeUpdate interface {
	//BeforeUpdate(*gin.Context, *kuu.H, *kuu.H, bool)
	BeforeUpdate(*Scope) error
}

type IAfterUpdate interface {
	//AfterUpdate(*gin.Context, interface{})
	AfterUpdate(*Scope) error
}
