package kuu

func init() {
	DefaultCallback.Update().Register("kuu:biz_before_update", bizBeforeUpdateCallback)
	DefaultCallback.Update().Register("kuu:biz_update", bizUpdateCallback)
	DefaultCallback.Update().Register("kuu:biz_after_update", bizAfterUpdateCallback)
}

func bizBeforeUpdateCallback(scope *Scope) {
	if !scope.HasError() {
		scope.CallMethod("BizBeforeUpdate")
	}
}

func bizUpdateCallback(scope *Scope) {
	if !scope.HasError() {
		if scope.IsAutoSave {
			scope.DB.Save(scope.Value)
		} else {
			scope.DB = scope.DB.Model(scope.UpdateCond).Updates(scope.Value)
		}
	}
}

func bizAfterUpdateCallback(scope *Scope) {
	if !scope.HasError() {
		scope.CallMethod("BizAfterUpdate")
	}
}
