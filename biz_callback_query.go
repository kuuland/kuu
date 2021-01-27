package kuu

import (
	"fmt"
	"math"
	"reflect"
)

func init() {
	DefaultCallback.Query().Register("kuu:biz_before_query", bizBeforeQueryCallback)
	DefaultCallback.Query().Register("kuu:biz_query", bizQueryCallback)
	DefaultCallback.Query().Register("kuu:biz_after_query", bizAfterQueryCallback)
}

func bizBeforeQueryCallback(scope *Scope) {
	if !scope.HasError() {
		scope.CallMethod("BizBeforeFind")
		if scope.Meta != nil {
			if err := execBizHooks(fmt.Sprintf("%s:BizBeforeFind", scope.Meta.Name), scope); err != nil {
				_ = scope.Err(err)
				return
			}
		}
	}
}
func bizQueryCallback(scope *Scope) {
	if !scope.HasError() {
		if err := scope.DB.Find(scope.QueryResult.List).Error; err != nil {
			_ = scope.Err(err)
			return
		}
		scope.QueryResult.List = Meta(reflect.New(scope.ReflectType).Interface()).OmitPassword(scope.QueryResult.List)
		scope.QueryResult.List = ProjectFields(scope.QueryResult.List, scope.QueryResult.Project)

		// 处理totalrecords、totalpages
		var totalRecords int
		if err := scope.DB.Offset(-1).Limit(-1).Count(&totalRecords).Error; err != nil {
			_ = scope.Err(err)
			return
		}
		scope.QueryResult.TotalRecords = totalRecords
		if scope.QueryResult.Range == "PAGE" {
			scope.QueryResult.TotalPages = int(math.Ceil(float64(totalRecords) / float64(scope.QueryResult.Size)))
		}
	}
}

func bizAfterQueryCallback(scope *Scope) {
	if !scope.HasError() {
		scope.CallMethod("BizAfterFind")
		if scope.Meta != nil {
			if err := execBizHooks(fmt.Sprintf("%s:BizAfterFind", scope.Meta.Name), scope); err != nil {
				_ = scope.Err(err)
				return
			}
		}
	}
}
