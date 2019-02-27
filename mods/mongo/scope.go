package mongo

import (
	"github.com/globalsign/mgo"
	"github.com/kuuland/kuu"
)

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

type Scope struct {
	Operation  string
	Session    *mgo.Session
	Collection *mgo.Collection
	Query      *mgo.Query
	Cache      kuu.H
	CreateData *[]interface{}
}

func (scope *Scope) CallMethod(action int, schema *kuu.Schema) (err error) {
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
	return err
}
