package sys

import (
	"fmt"
	"github.com/globalsign/mgo/bson"
	"github.com/kuuland/kuu"
	"github.com/kuuland/kuu/mods/mongo"
	"github.com/kuuland/kuu/mods/sys/models"
)

func hooksInit() {
	mongo.AddGlobalHook(mongo.BeforeCreateEnum, func(scope *mongo.Scope) error {
		routineCache := kuu.GetGoroutineCache()
		loginOrgID := routineCache["LoginOrgID"]
		if v, ok := loginOrgID.(string); ok && v != "" {
			orgID := bson.ObjectIdHex(v)
			createData := *scope.CreateData
			for index, item := range createData {
				doc := item.(kuu.H)
				doc["Org"] = orgID
				createData[index] = doc
			}
		}
		return nil
	})
	mongo.AddGlobalHook(mongo.BeforeFindEnum, func(scope *mongo.Scope) error {
		scopeCache := *scope.Cache
		routineCache := kuu.GetGoroutineCache()
		loginUID := routineCache["LoginUID"]
		if skipAuth(loginUID, scope) {
			return nil
		}
		uid := loginUID.(string)
		scopeCache["LoginUID"] = uid
		loginOrgID := routineCache["LoginOrgID"]
		if loginOrgID != nil {
			orgID := loginOrgID.(string)
			var rule models.AuthRule
			AuthRule := kuu.Model("AuthRule")
			AuthRule.One(kuu.H{
				"Cond": kuu.H{
					"UID":        uid,
					"OrgID":      orgID,
					"ObjectName": scope.Schema.Name,
				},
			}, &rule)
			orgIDs := []bson.ObjectId{}
			if rule.ReadableOrgIDs != nil {
				for _, item := range rule.ReadableOrgIDs {
					orgIDs = append(orgIDs, bson.ObjectIdHex(item))
				}
			}
			cond := scope.Params.Cond
			scopeCache["rawCond"] = cond
			if cond["$and"] == nil {
				cond = kuu.H{
					"$and": []interface{}{cond},
				}
			}
			var and []kuu.H
			kuu.JSONConvert(cond["$and"], &and)
			and = append(and, kuu.H{
				"$or": []kuu.H{
					{"Org": kuu.H{"$exists": false}},
					{"Org": kuu.H{"$in": orgIDs}},
					{"CreatedBy": bson.ObjectIdHex(uid)},
				},
			})
			cond["$and"] = and
			scope.Params.Cond = cond
			// 输出日志
			kuu.Info(fmt.Sprintf("数据范围控制（uid=%v，org=%v）：%v\n%v", uid, orgID, scope.Schema.Name, kuu.Stringify(scope.Params.Cond, true)))
		}
		return nil
	})
	mongo.AddGlobalHook(mongo.AfterFindEnum, func(scope *mongo.Scope) error {
		scopeCache := *scope.Cache
		loginUID := scopeCache["LoginUID"]
		if skipAuth(loginUID, scope) {
			return nil
		}
		if scope.ListData != nil {
			cache := *scope.Cache
			if cache != nil && cache["rawCond"] != nil {
				scope.ListData["cond"] = cache["rawCond"]
			}
		}
		return nil
	})
}

func skipAuth(loginUID interface{}, scope *mongo.Scope) bool {
	if scope.Schema.NoAuth || loginUID == nil || loginUID == kuu.Data["RootUID"] {
		return true
	}
	return false
}
