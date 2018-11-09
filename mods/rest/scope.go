package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
)

const (
	BeforeCreateEnum = iota
	BeforeUpdateEnum
	AfterUpdateEnum
	AfterCreateEnum
	BeforeRemoveEnum
	AfterRemoveEnum
	BeforeListEnum
	AfterListEnum
	BeforeIDEnum
	AfterIDEnum
)

type Scope struct {
	Context      *gin.Context
	Cache        kuu.H
	Model        kuu.IModel
	Params       *Params
	CreateData   *[]kuu.H
	ResponseData *kuu.H
	RemoveCond   *kuu.H
	RemoveDoc    *kuu.H
	RemoveAll    bool
	UpdateCond   *kuu.H
	UpdateDoc    *kuu.H
	UpdateAll    bool
}

func (scope *Scope) CallMethod(action int, schema *kuu.Schema) (err error) {
	switch action {
	case BeforeCreateEnum:
		if s, ok := schema.Origin.(IBeforeCreate); ok {
			err = s.BeforeCreate(scope)
		}
	case AfterCreateEnum:
		if s, ok := schema.Origin.(IAfterCreate); ok {
			err = s.AfterCreate(scope)
		}
	case BeforeUpdateEnum:
		if s, ok := schema.Origin.(IBeforeUpdate); ok {
			err = s.BeforeUpdate(scope)
		}
	case AfterUpdateEnum:
		if s, ok := schema.Origin.(IAfterUpdate); ok {
			err = s.AfterUpdate(scope)
		}
	case BeforeRemoveEnum:
		if s, ok := schema.Origin.(IBeforeRemove); ok {
			err = s.BeforeRemove(scope)
		}
	case AfterRemoveEnum:
		if s, ok := schema.Origin.(IAfterRemove); ok {
			err = s.AfterRemove(scope)
		}
	case BeforeListEnum:
		if s, ok := schema.Origin.(IBeforeList); ok {
			err = s.BeforeList(scope)
		}
	case AfterListEnum:
		if s, ok := schema.Origin.(IAfterList); ok {
			err = s.AfterList(scope)
		}
	case BeforeIDEnum:
		if s, ok := schema.Origin.(IBeforeID); ok {
			err = s.BeforeID(scope)
		}
	case AfterIDEnum:
		if s, ok := schema.Origin.(IAfterID); ok {
			err = s.AfterID(scope)
		}
	}
	return err
}
