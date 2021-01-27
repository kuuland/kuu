package kuu

import "fmt"

func init() {
	DefaultCallback.Create().Register("kuu:biz_before_create", bizBeforeCreateCallback)
	DefaultCallback.Create().Register("kuu:biz_create", bizCreateCallback)
	DefaultCallback.Create().Register("kuu:biz_after_create", bizAfterCreateCallback)
}

func bizBeforeCreateCallback(scope *Scope) {
	if !scope.HasError() {
		scope.CallMethod("BizBeforeCreate")
		if scope.Meta != nil {
			if err := execBizHooks(fmt.Sprintf("%s:BizBeforeCreate", scope.Meta.Name), scope); err != nil {
				_ = scope.Err(err)
				return
			}
		}
	}
}

func bizCreateCallback(scope *Scope) {
	if !scope.HasError() {
		if err := scope.DB.Create(scope.Value).Error; err != nil {
			_ = scope.Err(err)
			return
		}
	}
}

func bizAfterCreateCallback(scope *Scope) {
	if !scope.HasError() {
		scope.CallMethod("BizAfterCreate")
		if scope.Meta != nil {
			if err := execBizHooks(fmt.Sprintf("%s:BizAfterCreate", scope.Meta.Name), scope); err != nil {
				_ = scope.Err(err)
				return
			}
		}
	}
}
