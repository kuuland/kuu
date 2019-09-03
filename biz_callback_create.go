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
		// 先创建主表、再创建/更新子表
		err := scope.DB.
			Set("gorm:association_autoupdate", false).
			Set("gorm:association_autocreate", false).
			Create(scope.Value).Error
		if err != nil {
			_ = scope.Err(err)
			return
		}
		dbScope := scope.DB.NewScope(scope.Value)
		for _, field := range dbScope.Fields() {
			checkCreateOrUpdateField(scope, field)
		}
	}
}

func bizAfterCreateCallback(scope *Scope) {
	if !scope.HasError() {
		scope.CallMethod("BizAfterCreate")
	}
}
