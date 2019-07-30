package kuu

import (
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
		scope.CallMethod("BizBeforeQuery")
	}
}
func bizQueryCallback(scope *Scope) {
	if !scope.HasError() {
		if err := scope.DB.Find(scope.QueryResult.List).Error; err != nil {
			_ = scope.Err(err)
			return
		}
		scope.QueryResult.List = Meta(reflect.New(scope.ReflectType).Interface()).OmitPassword(scope.QueryResult.List)
		// 处理totalrecords、totalpages
		var totalRecords int
		scope.DB = CountWheres(scope.Value, scope.DB)
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
		scope.CallMethod("BizAfterQuery")
	}
}
