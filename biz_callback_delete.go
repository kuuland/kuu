package kuu

func init() {
	DefaultCallback.Delete().Register("kuu:biz_before_delete", bizBeforeDeleteCallback)
	DefaultCallback.Delete().Register("kuu:biz_delete", bizDeleteCallback)
	DefaultCallback.Delete().Register("kuu:biz_after_delete", bizAfterDeleteCallback)
}

func bizBeforeDeleteCallback(scope *Scope) {
	if !scope.HasError() {
		scope.CallMethod("BizBeforeDelete")
	}
}

func bizDeleteCallback(scope *Scope) {
	if !scope.HasError() {
		if err := scope.DB.Delete(scope.Value).Error; err != nil {
			_ = scope.Err(err)
		}
	}
}

func bizAfterDeleteCallback(scope *Scope) {
	if !scope.HasError() {
		scope.CallMethod("BizAfterDelete")
	}
}
