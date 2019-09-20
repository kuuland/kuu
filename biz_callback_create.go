package kuu

func init() {
	DefaultCallback.Create().Register("kuu:biz_before_create", bizBeforeCreateCallback)
	DefaultCallback.Create().Register("kuu:biz_create", bizCreateCallback)
	DefaultCallback.Create().Register("kuu:biz_after_create", bizAfterCreateCallback)
}

func bizBeforeCreateCallback(scope *Scope) {
	if !scope.HasError() {
		scope.CallMethod("BizBeforeCreate")
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
	}
}
