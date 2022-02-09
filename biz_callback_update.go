package kuu

import "fmt"

func init() {
	DefaultCallback.Update().Register("kuu:biz_before_update", bizBeforeUpdateCallback)
	DefaultCallback.Update().Register("kuu:biz_update", bizUpdateCallback)
	DefaultCallback.Update().Register("kuu:biz_after_update", bizAfterUpdateCallback)
}

func bizBeforeUpdateCallback(scope *Scope) {
	if !scope.HasError() {
		scope.CallMethod("BizBeforeUpdate")
		if scope.Meta != nil {
			if err := execBizHooks(fmt.Sprintf("%s:BizBeforeUpdate", scope.Meta.Name), scope); err != nil {
				_ = scope.Err(err)
				return
			}
		}
	}
}

func bizUpdateCallback(scope *Scope) {
	if !scope.HasError() {
		scope.DB = scope.DB.Set("__KUU_UPDATE_BEFORE__", JSONStringify(scope.UpdateCond))
		scope.DB = scope.DB.Model(scope.UpdateCond).Updates(scope.Value)
		if err := scope.DB.Error; err != nil {
			_ = scope.Err(err)
			return
		}
	}
}

func bizAfterUpdateCallback(scope *Scope) {
	if !scope.HasError() {
		scope.CallMethod("BizAfterUpdate")
		if scope.Meta != nil {
			if err := execBizHooks(fmt.Sprintf("%s:BizAfterUpdate", scope.Meta.Name), scope); err != nil {
				_ = scope.Err(err)
				return
			}
		}
	}
}
