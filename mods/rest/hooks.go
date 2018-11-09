package rest

type IBeforeCreate interface {
	BeforeCreate(*Scope) error
}

type IAfterCreate interface {
	AfterCreate(*Scope) error
}

type IBeforeID interface {
	BeforeID(*Scope) error
}

type IAfterID interface {
	AfterID(*Scope) error
}

type IBeforeList interface {
	BeforeList(*Scope) error
}

type IAfterList interface {
	AfterList(*Scope) error
}

type IBeforeRemove interface {
	BeforeRemove(*Scope) error
}

type IAfterRemove interface {
	AfterRemove(*Scope) error
}

type IBeforeUpdate interface {
	BeforeUpdate(*Scope) error
}

type IAfterUpdate interface {
	AfterUpdate(*Scope) error
}
