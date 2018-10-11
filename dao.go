package kuu

// IDao 持久化接口
type IDao interface {
	Create(...interface{}) bool
	Delete() bool
	Update() bool
	List() []interface{}
	One() interface{}
	ID() interface{}
	Count() int
}
