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
		dbScope := scope.DB.NewScope(scope.Value)
		for _, field := range dbScope.Fields() {
			checkCreateOrUpdateField(scope, field)
		}
		err := scope.DB.
			Set("gorm:association_autoupdate", false).
			Create(scope.Value).Error
		if err != nil {
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
