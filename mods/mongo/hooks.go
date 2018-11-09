package mongo

type IBeforeSave interface {
	BeforeSave(*Scope) error
}

type IBeforeCreate interface {
	BeforeCreate(*Scope) error
}

type IAfterCreate interface {
	AfterCreate(*Scope) error
}

type IBeforeUpdate interface {
	BeforeUpdate(*Scope) error
}

type IAfterUpdate interface {
	AfterUpdate(*Scope) error
}

type IAfterSave interface {
	AfterSave(*Scope) error
}

type IBeforeRemove interface {
	BeforeRemove(*Scope) error
}

type IAfterRemove interface {
	AfterRemove(*Scope) error
}

type IBeforePhyRemove interface {
	BeforePhyRemove(*Scope) error
}

type IAfterPhyRemove interface {
	AfterPhyRemove(*Scope) error
}

type IBeforeFind interface {
	BeforeFind(*Scope) error
}

type IAfterFind interface {
	AfterFind(*Scope) error
}
