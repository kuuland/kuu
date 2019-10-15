package kuu

import (
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/jinzhu/gorm"
	"regexp"
	"strings"
	"time"
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
	// AfterSaveCallback
	AfterSaveCallback = afterSaveCallback
	// QueryCallback
	QueryCallback = queryCallback
	// ValidateCallback
	ValidateCallback = validateCallback
)

func registerCallbacks() {
	callback := DB().Callback()
	// 注册系统callback
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
	if callback.Create().Get("kuu:update_ts") == nil {
		callback.Create().Before("gorm:before_create").Register("kuu:update_ts", updateTsCallback)
	}
	if callback.Create().Get("kuu:create") == nil {
		callback.Create().Before("gorm:create").Register("kuu:create", CreateCallback)
	}
	if callback.Create().Get("kuu:after_save") == nil {
		callback.Create().After("gorm:after_create").Register("kuu:after_save", AfterSaveCallback)
	}
	if callback.Update().Get("kuu:update_ts") == nil {
		callback.Update().Before("gorm:assign_updating_attributes").Register("kuu:update_ts", updateTsCallback)
	}
	if callback.Update().Get("kuu:update") == nil {
		callback.Update().Before("gorm:update").Register("kuu:update", UpdateCallback)
	}
	if callback.Update().Get("kuu:after_save") == nil {
		callback.Update().After("gorm:after_update").Register("kuu:after_save", AfterSaveCallback)
	}
	if callback.Delete().Get("kuu:delete") == nil {
		callback.Delete().Replace("gorm:delete", DeleteCallback)
	}
	// 注册数据变更callback
	if callback.Create().Get("kuu:model_change") == nil {
		callback.Create().After("gorm:after_create").Register("kuu:model_change", modelChangeCallback)
	}
	if callback.Update().Get("kuu:model_change") == nil {
		callback.Update().After("gorm:after_update").Register("kuu:model_change", modelChangeCallback)
	}
	if callback.Delete().Get("kuu:model_change") == nil {
		callback.Delete().After("gorm:after_delete").Register("kuu:model_change", modelChangeCallback)
	}
	// 注册审计callback
	if C().DefaultGetBool("audit:callbacks", true) {
		registerAuditCallbacks(callback)
	}
}

func beforeQueryCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		scope.CallMethod("BeforeFind")
	}
}

func updateTsCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		// 注意：必须修改update_interface才能确保后续callback取到子表值
		now := time.Now()
		if v, ok := scope.InstanceGet("gorm:update_interface"); ok {
			newScope := scope.New(v)
			if field, ok := newScope.FieldByName("Ts"); ok {
				_ = field.Set(now)
				_ = newScope.SetColumn(field.DBName, now)
			}
			scope.InstanceSet("gorm:update_interface", newScope.Value)
		} else {
			if field, ok := scope.FieldByName("Ts"); ok {
				_ = field.Set(now)
				_ = scope.SetColumn(field.DBName, now)
			}
		}
	}
}

func createCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		if desc := GetRoutinePrivilegesDesc(); desc != nil {
			var (
				orgID       uint
				createdByID uint
			)
			if field, ok := scope.FieldByName("CreatedByID"); ok {
				if err := scope.SetColumn(field.DBName, desc.UID); err != nil {
					_ = scope.Err(fmt.Errorf("自动设置创建人ID失败：%s", err.Error()))
					return
				}
				createdByID = field.Field.Interface().(uint)
			}
			if field, ok := scope.FieldByName("UpdatedByID"); ok {
				if err := scope.SetColumn(field.DBName, desc.UID); err != nil {
					_ = scope.Err(fmt.Errorf("自动设置修改人ID失败：%s", err.Error()))
					return
				}
			}
			if field, ok := scope.FieldByName("OrgID"); ok {
				if field.IsBlank && desc.ActOrgID != 0 {
					if err := scope.SetColumn(field.DBName, desc.ActOrgID); err != nil {
						_ = scope.Err(fmt.Errorf("自动设置组织ID失败：%s", err.Error()))
						return
					}
				}
				orgID = field.Field.Interface().(uint)
			}

			// 有忽略标记时
			if caches := GetRoutineCaches(); caches != nil {
				if _, ignoreAuth := caches[GLSIgnoreAuthKey]; ignoreAuth {
					return
				}
			}
			auth := GetAuthProcessorDesc(scope, desc)
			auth.OrgID = orgID
			auth.CreatedByID = createdByID
			if err := ActiveAuthProcessor.AllowCreate(auth); err != nil {
				_ = scope.Err(err)
				return
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
		var desc *PrivilegesDesc
		if desc = GetRoutinePrivilegesDesc(); desc.IsValid() {
			auth := GetAuthProcessorDesc(scope, desc)
			if err := ActiveAuthProcessor.AddWritableWheres(auth); err != nil {
				_ = scope.Err(err)
				return
			}
		}

		if !scope.Search.Unscoped && hasDeletedAtField {
			var sql string
			if desc != nil {
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
		if scope.DB().RowsAffected < 1 && desc.IsValid() {
			_ = scope.Err(ErrAffectedDeleteToken)
			return
		}
	}
}

func modelChangeCallback(scope *gorm.Scope) {
	if !scope.HasError() && scope.Value != nil {
		meta := Meta(scope.Value)
		if meta != nil {
			NotifyModelChange(meta.Name)
		}
	}
}

func updateCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		if desc := GetRoutinePrivilegesDesc(); desc.IsValid() {
			// 添加可写权限控制
			auth := GetAuthProcessorDesc(scope, desc)
			if err := ActiveAuthProcessor.AddWritableWheres(auth); err != nil {
				_ = scope.Err(err)
				return
			}
			if err := scope.SetColumn("UpdatedByID", desc.UID); err != nil {
				ERROR("自动设置修改人ID失败：%s", err.Error())
			}
		}
	}
}

func afterSaveCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		// 处理子表更新
		for _, field := range scope.Fields() {
			checkCreateOrUpdateField(scope, field)
		}
		// 判断是否写入
		desc := GetRoutinePrivilegesDesc()
		if scope.DB().RowsAffected < 1 && desc.IsValid() {
			_ = scope.Err(ErrAffectedSaveToken)
			return
		}
	}
}

func queryCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		if desc := GetRoutinePrivilegesDesc(); desc.IsValid() {
			auth := GetAuthProcessorDesc(scope, desc)
			if err := ActiveAuthProcessor.AddReadableWheres(auth); err != nil {
				_ = scope.Err(err)
				return
			}
		}
	}
}

func validateCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		if _, ok := scope.Get("gorm:update_column"); !ok {
			result, ok := scope.DB().Get(skipValidations)
			if !(ok && result.(bool)) {
				scope.CallMethod("Validate")
				if scope.Value == nil {
					return
				}
				resource := scope.IndirectValue().Interface()
				_, validatorErrors := govalidator.ValidateStruct(resource)
				if validatorErrors != nil {
					if errs, ok := validatorErrors.(govalidator.Errors); ok {
						for _, err := range FlatValidatorErrors(errs) {
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
		if errs, ok := validatorError.(govalidator.Errors); ok {
			for _, e := range errs {
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

func createOrUpdateItem(scope *gorm.Scope, item interface{}) {
	if IsNil(item) {
		return
	}

	tx := scope.DB()
	if tx.NewRecord(item) {
		if err := tx.Create(item).Error; err != nil {
			_ = scope.Err(err)
			return
		}
	} else {
		itemScope := DB().NewScope(item)
		if field, ok := itemScope.FieldByName("DeletedAt"); ok && !field.IsBlank {
			if err := tx.Delete(item).Error; err != nil {
				_ = scope.Err(err)
				return
			}
		} else {
			if err := tx.Model(item).Updates(item).Error; err != nil {
				_ = scope.Err(err)
				return
			}
		}
	}
}

func checkCreateOrUpdateField(scope *gorm.Scope, field *gorm.Field) {
	// 只需要处理has_many和has_one
	// belongs_to和many_to_many不允许直接创建或更新关联档案
	if field.Relationship != nil && !field.IsBlank {
		switch field.Relationship.Kind {
		case "has_many":
			for i := 0; i < field.Field.Len(); i++ {
				item := addrValue(field.Field.Index(i)).Interface()
				createOrUpdateItem(scope, item)
			}
		case "has_one":
			item := addrValue(field.Field).Interface()
			createOrUpdateItem(scope, item)
		}
	}
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
