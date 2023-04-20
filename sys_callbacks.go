package kuu

import (
	"bytes"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/guregu/null.v3"
	"reflect"
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
	if callback.Create().Get("kuu:uuid_create") == nil {
		callback.Create().Before("gorm:begin_transaction").Register("kuu:uuid_create", uuidCreateCallback)
	}
	if callback.Create().Get("validations:validate") == nil {
		callback.Create().Before("gorm:before_create").Register("validations:validate", ValidateCallback)
	}
	if callback.Update().Get("validations:validate") == nil {
		callback.Update().Before("gorm:before_update").Register("validations:validate", ValidateCallback)
	}
	if callback.Query().Get("kuu:before_query") == nil {
		callback.Query().Before("gorm:query").Register("kuu:before_query", BeforeQueryCallback)
	}
	if callback.Query().Get("kuu:query") == nil {
		callback.Query().Before("gorm:query").Register("kuu:query", QueryCallback)
	}
	if callback.RowQuery().Get("kuu:row_query") == nil {
		callback.RowQuery().Before("gorm:row_query").Register("kuu:row_query", QueryCallback)
	}
	if callback.Create().Get("kuu:update_ts") == nil {
		callback.Create().Before("gorm:create").Register("kuu:update_ts", updateTsForCreateCallback)
	}
	if callback.Create().Get("kuu:create") == nil {
		callback.Create().Before("gorm:create").Register("kuu:create", CreateCallback)
	}
	if callback.Create().Get("kuu:after_save") == nil {
		callback.Create().After("gorm:after_create").Register("kuu:after_save", AfterSaveCallback)
	}
	if callback.Update().Get("kuu:update_ts") == nil {
		callback.Update().Before("gorm:assign_updating_attributes").Register("kuu:update_ts", updateTsForUpdateCallback)
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
	// 注册持久层Hooks
	if callback.Create().Get("kuu:exec_before_create_hooks") == nil {
		callback.Create().After("gorm:before_create").Register("kuu:exec_before_create_hooks", func(scope *gorm.Scope) {
			execHooksCallback("BeforeSave", scope)
			execHooksCallback("BeforeCreate", scope)
		})
	}
	if callback.Create().Get("kuu:exec_after_create_hooks") == nil {
		callback.Create().After("gorm:after_create").Register("kuu:exec_after_create_hooks", func(scope *gorm.Scope) {
			execHooksCallback("AfterCreate", scope)
			execHooksCallback("AfterSave", scope)
		})
	}
	if callback.Delete().Get("kuu:exec_before_delete_hooks") == nil {
		callback.Delete().After("gorm:before_delete").Register("kuu:exec_before_delete_hooks", func(scope *gorm.Scope) {
			execHooksCallback("BeforeDelete", scope)
		})
	}
	if callback.Delete().Get("kuu:exec_after_delete_hooks") == nil {
		callback.Delete().After("gorm:after_delete").Register("kuu:exec_after_delete_hooks", func(scope *gorm.Scope) {
			execHooksCallback("AfterDelete", scope)
		})
	}
	if callback.Update().Get("kuu:exec_before_update_hooks") == nil {
		callback.Update().After("gorm:before_update").Register("kuu:exec_before_update_hooks", func(scope *gorm.Scope) {
			execHooksCallback("BeforeSave", scope)
			execHooksCallback("BeforeUpdate", scope)
		})
	}
	if callback.Update().Get("kuu:exec_after_update_hooks") == nil {
		callback.Update().After("gorm:after_update").Register("kuu:exec_after_update_hooks", func(scope *gorm.Scope) {
			execHooksCallback("AfterUpdate", scope)
			execHooksCallback("AfterSave", scope)
		})
	}
	if callback.Query().Get("kuu:exec_before_update_hooks") == nil {
		callback.Query().After("kuu:before_query").Register("kuu:exec_before_update_hooks", func(scope *gorm.Scope) {
			execHooksCallback("BeforeFind", scope)
		})
	}
	if callback.Query().Get("kuu:exec_after_query_hooks") == nil {
		callback.Query().After("gorm:after_query").Register("kuu:exec_after_query_hooks", func(scope *gorm.Scope) {
			execHooksCallback("AfterFind", scope)
		})
	}
}

func uuidCreateCallback(scope *gorm.Scope) {
	// 注意：由于gorm的scope.Fields函数内部缓存了主键的IsBlank状态，所以只能手动设置
	meta := Meta(scope.Value)
	if v, exists := meta.TagSettings["UUID"]; exists && v != "" {
		reflectValue := indirectValue(scope.Value)
		fieldValue := reflect.Indirect(reflectValue).FieldByName(v)
		fieldValue.SetString(strings.ReplaceAll(uuid.NewV4().String(), "-", ""))
	}
}

func beforeQueryCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		scope.CallMethod("BeforeFind")
	}
}

func updateTsForCreateCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		now := time.Now()

		if tsField, ok := scope.FieldByName("Ts"); ok {
			if tsField.IsBlank {
				tsField.Set(now)
			}
		}
	}
}

func updateTsForUpdateCallback(scope *gorm.Scope) {
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

func setMetadata(scope *gorm.Scope) *Metadata {
	if !scope.HasError() {
		if scope.Value == nil {
			return nil
		}
		meta := Meta(scope.Value)
		if meta != nil {
			scope.InstanceSet("Metadata", meta)
		}
		return meta
	}
	return nil
}

func getMetadata(scope *gorm.Scope) *Metadata {
	var meta *Metadata
	if v, has := scope.InstanceGet("Metadata"); has {
		meta = v.(*Metadata)
	} else if scope.Value != nil {
		meta = setMetadata(scope)
	}
	return meta
}

func execHooksCallback(partName string, scope *gorm.Scope) {
	if !scope.HasError() {
		if meta := getMetadata(scope); meta != nil {
			if err := execGormHooks(fmt.Sprintf("%s:%s", meta.Name, partName), scope); err != nil {
				_ = scope.Err(err)
				return
			}
		}
	}
}

func getDescFromDBScope(scope *gorm.Scope) *PrivilegesDesc {
	if idesc, has := scope.Get(GLSPrisDescKey); has && idesc != nil {
		desc := idesc.(*PrivilegesDesc)
		return desc
	}
	return nil
}

func createCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		if desc := getDescFromDBScope(scope); desc != nil {
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
				if v, ok := field.Field.Interface().(uint); ok {
					orgID = v
				} else if v, ok := field.Field.Interface().(int); ok {
					orgID = uint(v)
				} else if v, ok := field.Field.Interface().(int64); ok {
					orgID = uint(v)
				} else if v, ok := field.Field.Interface().(null.Int); ok {
					orgID = uint(v.Int64)
				}
			}

			// 有忽略标记时
			if getAuthIgnoreFromDBScope(scope) {
				return
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
		if desc = getDescFromDBScope(scope); desc != nil && desc.IsValid() {
			auth := GetAuthProcessorDesc(scope, desc)
			if err := ActiveAuthProcessor.AddWritableWheres(auth); err != nil {
				_ = scope.Err(err)
				return
			}
		}

		if !scope.Search.Unscoped && hasDeletedAtField {
			var (
				sqlbuf bytes.Buffer
				attrs  []interface{}
			)
			sqlbuf.WriteString("UPDATE %v SET ")
			attrs = append(attrs, scope.QuotedTableName())
			// 更新DeletedByID
			if desc != nil {
				if f, has := scope.FieldByName("DeletedByID"); has {
					sqlbuf.WriteString("%v=%v,")
					attrs = append(attrs,
						scope.Quote(f.DBName),
						scope.AddToVars(desc.UID),
					)
				}
			}
			// 更新Dr
			if f, has := scope.FieldByName("Dr"); has {
				sqlbuf.WriteString("%v=%v,")
				attrs = append(attrs,
					scope.Quote(f.DBName),
					scope.AddToVars(time.Now().Unix()),
				)
			}
			sqlbuf.WriteString("%v=%v%v%v")
			attrs = append(attrs,
				scope.Quote(deletedAtField.DBName),
				scope.AddToVars(gorm.NowFunc()),
				AddExtraSpaceIfExist(scope.CombinedConditionSql()),
				AddExtraSpaceIfExist(extraOption),
			)
			sql := fmt.Sprintf(sqlbuf.String(), attrs...)
			scope.Raw(sql).Exec()
		} else {
			scope.Raw(fmt.Sprintf(
				"DELETE FROM %v%v%v",
				scope.QuotedTableName(),
				AddExtraSpaceIfExist(scope.CombinedConditionSql()),
				AddExtraSpaceIfExist(extraOption),
			)).Exec()
		}
		if scope.DB().RowsAffected < 1 {
			WARN("未删除任何记录，请检查更新条件或数据权限")
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
		if desc := getDescFromDBScope(scope); desc != nil && desc.IsValid() {
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
		if scope.DB().RowsAffected < 1 {
			WARN("未新增或修改任何记录，请检查更新条件或数据权限")
			return
		}
	}
}

func queryCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		if desc := getDescFromDBScope(scope); desc != nil && desc.IsValid() {
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

	tx := scope.NewDB()
	if tx.NewRecord(item) {
		if err := tx.Create(item).Error; err != nil {
			_ = scope.Err(err)
			return
		}
	} else {
		var (
			itemScope          = DB().NewScope(item)
			meta               = Meta(item)
			deletedAtFieldName = "DeletedAt"
		)
		if v, exists := meta.TagSettings["DELETEDAT"]; exists && v != "" {
			deletedAtFieldName = v
		}
		if field, ok := itemScope.FieldByName(deletedAtFieldName); ok && !field.IsBlank {
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
