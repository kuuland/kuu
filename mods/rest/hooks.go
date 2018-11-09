package rest

type IBeforeCreateRoute interface {
	BeforeCreateRoute(*Scope) error
}

type IAfterCreateRoute interface {
	AfterCreateRoute(*Scope) error
}

type IBeforeIDRoute interface {
	BeforeIDRoute(*Scope) error
}

type IAfterIDRoute interface {
	AfterIDRoute(*Scope) error
}

type IBeforeListRoute interface {
	BeforeListRoute(*Scope) error
}

type IAfterListRoute interface {
	AfterListRoute(*Scope) error
}

type IBeforeRemoveRoute interface {
	BeforeRemoveRoute(*Scope) error
}

type IAfterRemoveRoute interface {
	AfterRemoveRoute(*Scope) error
}

type IBeforeUpdateRoute interface {
	BeforeUpdateRoute(*Scope) error
}

type IAfterUpdateRoute interface {
	AfterUpdateRoute(*Scope) error
}
