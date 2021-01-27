package kuu

import "fmt"

func init() {
	DefaultCallback.Delete().Register("kuu:biz_before_delete", bizBeforeDeleteCallback)
	DefaultCallback.Delete().Register("kuu:biz_delete", bizDeleteCallback)
	DefaultCallback.Delete().Register("kuu:biz_after_delete", bizAfterDeleteCallback)
}

func bizBeforeDeleteCallback(scope *Scope) {
	if !scope.HasError() {
		scope.CallMethod("BizBeforeDelete")
		if scope.Meta != nil {
			if err := execBizHooks(fmt.Sprintf("%s:BizBeforeDelete", scope.Meta.Name), scope); err != nil {
				_ = scope.Err(err)
				return
			}
		}
	}
}

func bizDeleteCallback(scope *Scope) {
	if !scope.HasError() {
		scope.DB = scope.DB.Delete(scope.Value)
		if err := scope.DB.Error; err != nil {
			_ = scope.Err(err)
			return
		}
	}
}

func bizAfterDeleteCallback(scope *Scope) {
	if !scope.HasError() {
		scope.CallMethod("BizAfterDelete")
		if scope.Meta != nil {
			if err := execBizHooks(fmt.Sprintf("%s:BizAfterDelete", scope.Meta.Name), scope); err != nil {
				_ = scope.Err(err)
				return
			}
		}
	}
}
