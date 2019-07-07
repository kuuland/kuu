package kuu

import (
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/jinzhu/gorm"
	"regexp"
	"strings"
)

var (
	skipValidations = "validations:skip_validations"
	// CreateCallback
	CreateCallback = createCallback
	// BeforeQueryCallback
	BeforeQueryCallback = beforeQueryCallback
	// DeleteCallback
	DeleteCallback = deleteCallback
	// UpdateCallback
	UpdateCallback = updateCallback
	// QueryCallback
	QueryCallback = queryCallback
	// ValidateCallback
	ValidateCallback = validateCallback
)

func registerCallbacks() {
	callback := DB().Callback()
	if callback.Create().Get("validations:validate") == nil {
		callback.Create().Before("gorm:before_create").Register("validations:validate", ValidateCallback)
	}
	if callback.Update().Get("validations:validate") == nil {
		callback.Update().Before("gorm:before_update").Register("validations:validate", ValidateCallback)
	}
	if callback.Query().Get("kuu:query") == nil {
		callback.Query().Before("gorm:query").Register("kuu:query", QueryCallback)
	}
	if callback.Query().Get("kuu:before_query") == nil {
		callback.Query().Before("gorm:query").Register("kuu:before_query", BeforeQueryCallback)
	}
	if callback.Update().Get("kuu:update") == nil {
		callback.Update().Before("gorm:update").Register("kuu:update", UpdateCallback)
	}
	if callback.Delete().Get("kuu:delete") == nil {
		callback.Delete().Replace("gorm:delete", DeleteCallback)
	}
	if callback.Create().Get("kuu:create") == nil {
		callback.Create().Before("gorm:create").Register("kuu:create", CreateCallback)
	}
	if C().DefaultGetBool("audit:callbacks", true) {
		registerAuditCallbacks(callback)
	}
}

func beforeQueryCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		scope.CallMethod("BeforeFind")
	}
}

func createCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		if desc := GetRoutinePrivilegesDesc(); desc != nil {
			if orgIDField, ok := scope.FieldByName("OrgID"); ok {
				if orgIDField.IsBlank {
					if err := orgIDField.Set(desc.SignOrgID); err != nil {
						ERROR("自动设置组织ID失败：%s", err.Error())
					}
				}
			}
			if createdByField, ok := scope.FieldByName("CreatedByID"); ok {
				if err := createdByField.Set(desc.UID); err != nil {
					ERROR("自动设置创建人ID失败：%s", err.Error())
				}
			}

			if updatedByField, ok := scope.FieldByName("UpdatedByID"); ok {
				if err := updatedByField.Set(desc.UID); err != nil {
					ERROR("自动设置修改人ID失败：%s", err.Error())
				}
			}
		}
	}
}

func deleteCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		var extraOption string
		if str, ok := scope.Get("gorm:delete_option"); ok {
			extraOption = fmt.Sprint(str)
		}

		deletedAtField, hasDeletedAtField := scope.FieldByName("DeletedAt")

		if !scope.Search.Unscoped && hasDeletedAtField {
			var sql string
			if desc := GetRoutinePrivilegesDesc(); desc != nil {
				deletedByField, hasDeletedByField := scope.FieldByName("DeletedByID")
				if !scope.Search.Unscoped && hasDeletedByField {
					sql = fmt.Sprintf(
						"UPDATE %v SET %v=%v,%v=%v%v%v",
						scope.QuotedTableName(),
						scope.Quote(deletedByField.DBName),
						scope.AddToVars(desc.UID),
						scope.Quote(deletedAtField.DBName),
						scope.AddToVars(gorm.NowFunc()),
						AddExtraSpaceIfExist(scope.CombinedConditionSql()),
						AddExtraSpaceIfExist(extraOption),
					)
				}
			}
			if sql == "" {
				sql = fmt.Sprintf(
					"UPDATE %v SET %v=%v%v%v",
					scope.QuotedTableName(),
					scope.Quote(deletedAtField.DBName),
					scope.AddToVars(gorm.NowFunc()),
					AddExtraSpaceIfExist(scope.CombinedConditionSql()),
					AddExtraSpaceIfExist(extraOption),
				)
			}
			scope.Raw(sql).Exec()
		} else {
			scope.Raw(fmt.Sprintf(
				"DELETE FROM %v%v%v",
				scope.QuotedTableName(),
				AddExtraSpaceIfExist(scope.CombinedConditionSql()),
				AddExtraSpaceIfExist(extraOption),
			)).Exec()
		}
	}
}

func updateCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		desc := GetRoutinePrivilegesDesc()
		if desc != nil {
			if err := scope.SetColumn("UpdatedByID", desc.UID); err != nil {
				ERROR("自动设置修改人ID失败：%s", err.Error())
			}
		}
	}
}

func queryCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		desc := GetRoutinePrivilegesDesc()
		values := GetRoutineValues()

		if desc == nil {
			// 无登录登录态时
			return
		}

		if values != nil {
			// 有忽略标记时
			if _, ignoreAuth := values[IgnoreAuthKey]; ignoreAuth {
				return
			}
			// 查询用户菜单时
			if _, queryUserMenus := values[UserMenusKey]; queryUserMenus {
				if desc.NotRootUser() {
					_, hasCodeField := scope.FieldByName("Code")
					_, hasCreatedByIDField := scope.FieldByName("CreatedByID")
					if hasCodeField && hasCreatedByIDField {
						// 菜单数据权限控制与组织无关，且只有两种情况：
						// 1.自己创建的，一定看得到
						// 2.别人创建的，必须通过分配操作权限才能看到
						scope.Search.Where("(code in (?)) OR (created_by_id = ?)", desc.Codes, desc.UID)
					}
				}
				return
			}
		}

		// 有登录态时的常规情况
		if desc.NotRootUser() {
			_, hasOrgIDField := scope.FieldByName("OrgID")
			_, hasCreatedByIDField := scope.FieldByName("CreatedByID")
			if hasOrgIDField && hasCreatedByIDField {
				scope.Search.Where("(org_id IS NULL) OR (org_id = 0) OR (org_id in (?)) OR (created_by_id = ?)", desc.ReadableOrgIDs, desc.UID)
			}
		}
	}
}

func validateCallback(scope *gorm.Scope) {
	if _, ok := scope.Get("gorm:update_column"); !ok {
		if result, ok := scope.DB().Get(skipValidations); !(ok && result.(bool)) {
			if !scope.HasError() {
				scope.CallMethod("Validate")
				if scope.Value != nil {
					resource := scope.IndirectValue().Interface()
					_, validatorErrors := govalidator.ValidateStruct(resource)
					if validatorErrors != nil {
						if errors, ok := validatorErrors.(govalidator.Errors); ok {
							for _, err := range FlatValidatorErrors(errors) {
								if err := scope.DB().AddError(formattedValidError(err, resource)); err != nil {
									ERROR("添加验证错误信息失败：%s", err.Error())
								}

							}
						} else {
							if err := scope.DB().AddError(validatorErrors); err != nil {
								ERROR("添加验证错误信息失败：%s", err.Error())
							}
						}
					}
				}
			}
		}
	}
}

// AddExtraSpaceIfExist
func AddExtraSpaceIfExist(str string) string {
	if str != "" {
		return " " + str
	}
	return ""
}

// FlatValidatorErrors
func FlatValidatorErrors(validatorErrors govalidator.Errors) []govalidator.Error {
	resultErrors := make([]govalidator.Error, 0)
	for _, validatorError := range validatorErrors.Errors() {
		if errors, ok := validatorError.(govalidator.Errors); ok {
			for _, e := range errors {
				resultErrors = append(resultErrors, e.(govalidator.Error))
			}
		}
		if e, ok := validatorError.(govalidator.Error); ok {
			resultErrors = append(resultErrors, e)
		}
	}
	return resultErrors
}

func formattedValidError(err govalidator.Error, resource interface{}) error {
	message := err.Error()
	attrName := err.Name
	if strings.Index(message, "non zero value required") >= 0 {
		message = fmt.Sprintf("%v can't be blank", attrName)
	} else if strings.Index(message, "as length") >= 0 {
		reg, _ := regexp.Compile(`\(([0-9]+)\|([0-9]+)\)`)
		submatch := reg.FindSubmatch([]byte(err.Error()))
		message = fmt.Sprintf("%v is the wrong length (should be %v~%v characters)", attrName, string(submatch[1]), string(submatch[2]))
	} else if strings.Index(message, "as numeric") >= 0 {
		message = fmt.Sprintf("%v is not a number", attrName)
	} else if strings.Index(message, "as email") >= 0 {
		message = fmt.Sprintf("%v is not a valid email address", attrName)
	}
	return NewValidError(resource, attrName, message)

}

// NewValidError generate a new error for a model's field
func NewValidError(resource interface{}, column, err string) error {
	return &ValidError{Resource: resource, Column: column, Message: err}
}

// ValidError is a validation error struct that hold model, column and error message
type ValidError struct {
	Resource interface{}
	Column   string
	Message  string
}

// ValidError is a label including model type, primary key and column name
func (err ValidError) Label() string {
	scope := gorm.Scope{Value: err.Resource}
	return fmt.Sprintf("%v_%v_%v", scope.GetModelStruct().ModelType.Name(), scope.PrimaryKeyValue(), err.Column)
}

// ValidError show error message
func (err ValidError) Error() string {
	return fmt.Sprintf("%v", err.Message)
}
