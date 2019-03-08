package mongo

import (
	"github.com/globalsign/mgo"
	"github.com/kuuland/kuu"
)

type Scope struct {
	Operation  string
	Session    *mgo.Session
	Collection *mgo.Collection
	Params     *Params
	ListData   kuu.H
	Cache      *kuu.H
	CreateData *[]interface{}
	UpdateCond *kuu.H
	UpdateDoc  *kuu.H
	Schema     *kuu.Schema
}

const (
	BeforeSaveEnum = iota
	BeforeCreateEnum
	BeforeUpdateEnum
	AfterUpdateEnum
	AfterCreateEnum
	AfterSaveEnum
	BeforeRemoveEnum
	AfterRemoveEnum
	BeforePhyRemoveEnum
	AfterPhyRemoveEnum
	BeforeFindEnum
	AfterFindEnum
)

var globalHooks map[int32][]func(*Scope) error

// AddGlobalHook 添加全局持久化钩子
func AddGlobalHook(action int32, handler func(*Scope) error) {
	if handler == nil {
		return
	}
	if globalHooks == nil {
		globalHooks = make(map[int32][]func(*Scope) error, 0)
	}
	hooks := globalHooks[action]
	if hooks == nil {
		hooks = make([]func(*Scope) error, 0)
	}
	hooks = append(hooks, handler)
	globalHooks[action] = hooks
}

func callGlobalHooks(action int32, scope *Scope) error {
	if globalHooks == nil {
		return nil
	}
	for key, value := range globalHooks {
		if key != action {
			continue
		}
		for _, handler := range value {
			if err := handler(scope); err != nil {
				return err
			}
		}
	}
	return nil
}

// CallMethod 调用钩子
func (scope *Scope) CallMethod(action int32, schema *kuu.Schema) (err error) {
	scope.Schema = schema
	// 调用全局钩子
	if err = callGlobalHooks(action, scope); err != nil {
		return
	}
	// 调用模型钩子
	switch action {
	case BeforeSaveEnum:
		if s, ok := schema.Origin.(IBeforeSave); ok {
			err = s.BeforeSave(scope)
		}
	case BeforeCreateEnum:
		if s, ok := schema.Origin.(IBeforeCreate); ok {
			err = s.BeforeCreate(scope)
		}
	case BeforeUpdateEnum:
		if s, ok := schema.Origin.(IBeforeUpdate); ok {
			err = s.BeforeUpdate(scope)
		}
	case AfterUpdateEnum:
		if s, ok := schema.Origin.(IAfterUpdate); ok {
			err = s.AfterUpdate(scope)
		}
	case AfterCreateEnum:
		if s, ok := schema.Origin.(IAfterCreate); ok {
			err = s.AfterCreate(scope)
		}
	case AfterSaveEnum:
		if s, ok := schema.Origin.(IAfterSave); ok {
			err = s.AfterSave(scope)
		}
	case BeforeRemoveEnum:
		if s, ok := schema.Origin.(IBeforeRemove); ok {
			err = s.BeforeRemove(scope)
		}
	case AfterRemoveEnum:
		if s, ok := schema.Origin.(IAfterRemove); ok {
			err = s.AfterRemove(scope)
		}
	case BeforePhyRemoveEnum:
		if s, ok := schema.Origin.(IBeforePhyRemove); ok {
			err = s.BeforePhyRemove(scope)
		}
	case AfterPhyRemoveEnum:
		if s, ok := schema.Origin.(IAfterPhyRemove); ok {
			err = s.AfterPhyRemove(scope)
		}
	case BeforeFindEnum:
		if s, ok := schema.Origin.(IBeforeFind); ok {
			err = s.BeforeFind(scope)
		}
	case AfterFindEnum:
		if s, ok := schema.Origin.(IAfterFind); ok {
			err = s.AfterFind(scope)
		}
	}
	scope.Schema = nil
	return err
}
