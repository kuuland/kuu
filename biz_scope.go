package kuu

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"reflect"
)

// Scope
type Scope struct {
	skipLeft     bool
	Value        interface{}
	Meta         *Metadata
	ReflectType  reflect.Type
	Context      *Context
	DB           *gorm.DB
	callbacks    *Callback
	QueryResult  *BizQueryResult
	UpdateCond   interface{}
	UpdateParams *BizUpdateParams
}

// BizUpdateParams
type BizUpdateParams struct {
	All   bool
	Multi bool
	Cond  map[string]interface{}
	Doc   map[string]interface{}
}

// BizQueryResult
type BizQueryResult struct {
	Cond         map[string]interface{} `json:"cond,omitempty"`
	Project      string                 `json:"project,omitempty"`
	Preload      string                 `json:"preload,omitempty"`
	Sort         string                 `json:"sort,omitempty"`
	Range        string                 `json:"range,omitempty"`
	Page         int                    `json:"page,omitempty"`
	Size         int                    `json:"size,omitempty"`
	TotalRecords int                    `json:"totalrecords,omitempty"`
	TotalPages   int                    `json:"totalpages,omitempty"`
	List         interface{}            `json:"list,omitempty"`
}

type BizPreloadInterface interface {
	BizPreloadHandlers() map[string]func(*gorm.DB) *gorm.DB
}

// NewBizScope
func NewBizScope(c *Context, value interface{}, db *gorm.DB) *Scope {
	reflectType := reflect.ValueOf(value).Type()
	for reflectType.Kind() == reflect.Slice || reflectType.Kind() == reflect.Ptr {
		reflectType = reflectType.Elem()
	}
	scope := &Scope{
		Context:     c,
		DB:          db,
		Value:       value,
		Meta:        Meta(value),
		callbacks:   DefaultCallback,
		ReflectType: reflectType,
	}
	return scope
}

// IndirectValue
func (scope *Scope) IndirectValue() reflect.Value {
	return indirectValue(scope.Value)
}

// SkipLeft skip remaining callbacks
func (scope *Scope) SkipLeft() {
	scope.skipLeft = true
}

// Err
func (scope *Scope) Err(err error) error {
	return scope.DB.AddError(err)
}

// HasError
func (scope *Scope) HasError() bool {
	return scope.DB.Error != nil
}

// CallMethod
func (scope *Scope) CallMethod(methodName string) {
	if scope.Value == nil {
		return
	}

	if indirectScopeValue := scope.IndirectValue(); indirectScopeValue.Kind() == reflect.Slice {
		for i := 0; i < indirectScopeValue.Len(); i++ {
			scope.callMethod(methodName, indirectScopeValue.Index(i))
		}
	} else {
		scope.callMethod(methodName, indirectScopeValue)
	}
}

func (scope *Scope) callCallbacks(kind string) *Scope {
	defer func() {
		if err := recover(); err != nil {
			if scope.DB != nil {
				scope.DB.Rollback()
			}
			panic(err)
		}
	}()

	var funcs []*func(s *Scope)
	switch kind {
	case BizCreateKind:
		funcs = scope.callbacks.creates
	case BizQueryKind:
		funcs = scope.callbacks.queries
	case BizUpdateKind:
		funcs = scope.callbacks.updates
	case BizDeleteKind:
		funcs = scope.callbacks.deletes
	}

	for _, f := range funcs {
		(*f)(scope)
		if scope.skipLeft {
			break
		}
	}

	return scope
}

func (scope *Scope) callMethod(methodName string, reflectValue reflect.Value) {
	// Only get address from non-pointer
	if reflectValue.CanAddr() && reflectValue.Kind() != reflect.Ptr {
		reflectValue = reflectValue.Addr()
	}
	if methodValue := reflectValue.MethodByName(methodName); methodValue.IsValid() {
		prisdesc, _ := scope.Context.Get(GLSPrisDescKey)
		scope.DB = scope.DB.Set(GLSPrisDescKey, prisdesc)
		switch method := methodValue.Interface().(type) {
		case func():
			method()
		case func(*Scope):
			method(scope)
		case func(*gorm.DB):
			method(scope.DB)
		case func(*gorm.DB) *gorm.DB:
			scope.DB = method(scope.DB)
		case func() error:
			_ = scope.Err(method())
		case func(*Scope) error:
			_ = scope.Err(method(scope))
		case func(*gorm.DB) error:
			_ = scope.Err(method(scope.DB))
		default:
			_ = scope.Err(fmt.Errorf("unsupported function %v", methodName))
		}
	}
}
