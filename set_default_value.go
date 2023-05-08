package kuu

import (
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
)

var _setDefaultValueMap = make(SetDefaultMap)

func init() {
	RegisterSetDefaultFn("__kuu_default_create", setDefaultCreateAudit)
}

type SetDefaultFunc func(ctx context.Context, entity any) error

type SetDefaultMap map[string][]SetDefaultFunc

func (set SetDefaultMap) Set(ctx context.Context, key string, entity any) error {
	if fns, has := set[key]; has {
		for _, fn := range fns {
			if err := fn(ctx, entity); err != nil {
				return err
			}
		}
	}
	return nil
}

func (set SetDefaultMap) Register(key string, fn SetDefaultFunc) {
	set[key] = append(set[key], fn)
}

func RegisterSetDefaultFn(key string, fn SetDefaultFunc) {
	_setDefaultValueMap.Register(key, fn)
}

func SetDefault(ctx context.Context, key string, entity any) error {
	return _setDefaultValueMap.Set(ctx, key, entity)
}

func setDefaultCreateAudit(ctx context.Context, entity any) error {
	d := ctx.Value("desc")
	desc, ok := d.(*PrivilegesDesc)
	if !ok {
		return nil
	}
	s := ctx.Value("scope")
	scope, ok := s.(*gorm.Scope)
	if !ok {
		return nil
	}
	if meta := getMetadata(scope); meta != nil {
		for _, fieldmeta := range meta.Fields {
			if key, has := fieldmeta.TagSetting["REF_META"]; has {
				switch key {
				case "org_id":
					if field, ok := scope.FieldByName(fieldmeta.Code); ok {
						if field.IsBlank && desc.ActOrgID != 0 {
							if err := scope.SetColumn(field.DBName, desc.ActOrgID); err != nil {
								_ = scope.Err(fmt.Errorf("自动设置组织ID失败：%s", err.Error()))
								return nil
							}
						}
					}
				case "org_code":
					if field, ok := scope.FieldByName(fieldmeta.Code); ok {
						if field.IsBlank && desc.ActOrgCode != "" {
							if err := scope.SetColumn(field.DBName, desc.ActOrgCode); err != nil {
								_ = scope.Err(fmt.Errorf("自动设置组织编码失败：%s", err.Error()))
								return nil
							}
						}
					}
				case "org_name":
					if field, ok := scope.FieldByName(fieldmeta.Code); ok {
						if field.IsBlank && desc.ActOrgName != "" {
							if err := scope.SetColumn(field.DBName, desc.ActOrgName); err != nil {
								_ = scope.Err(fmt.Errorf("自动设置组织编码失败：%s", err.Error()))
								return nil
							}
						}
					}
				case "user_id":
					if field, ok := scope.FieldByName(fieldmeta.Code); ok {
						if field.IsBlank {
							if err := scope.SetColumn(field.DBName, desc.UID); err != nil {
								_ = scope.Err(fmt.Errorf("自动设置创建人ID失败：%s", err.Error()))
								return nil
							}
						}
					}
				case "user_name":
					if field, ok := scope.FieldByName(fieldmeta.Code); ok {
						if field.IsBlank {
							if err := scope.SetColumn(field.DBName, desc.UserName); err != nil {
								_ = scope.Err(fmt.Errorf("自动设置创建人登陆名失败：%s", err.Error()))
								return nil
							}
						}
					}
				case "user_alias":
					if field, ok := scope.FieldByName(fieldmeta.Code); ok {
						if field.IsBlank {
							if err := scope.SetColumn(field.DBName, desc.UserAlias); err != nil {
								_ = scope.Err(fmt.Errorf("自动设置创建人登陆名失败：%s", err.Error()))
								return nil
							}
						}
					}
				}
			}
		}
	}
	return nil
}
