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
		dbScope := scope.DB.NewScope(scope.Value)
		for key, _ := range scope.UpdateParams.Doc {
			if field, ok := dbScope.FieldByName(key); ok {
				checkCreateOrUpdateField(scope, field)
			}
		}
		scope.DB = scope.DB.Model(scope.UpdateCond).
			Set("gorm:association_autoupdate", false).
			Set("gorm:association_autocreate", false).
			Updates(scope.Value)
		if err := scope.DB.Error; err != nil {
			_ = scope.Err(err)
		} else if scope.DB.RowsAffected < 1 {
			_ = scope.Err(ErrAffectedSaveToken)
			return
		}
	}
}

func bizAfterUpdateCallback(scope *Scope) {
	if !scope.HasError() {
		scope.CallMethod("BizAfterUpdate")
	}
}
