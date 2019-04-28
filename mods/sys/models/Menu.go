package models

import (
	"github.com/globalsign/mgo/bson"
	"github.com/kuuland/kuu"
	"github.com/kuuland/kuu/mods/mongo"
)

// Menu 系统菜单
type Menu struct {
	ID            string `json:"_id" displayName:"系统菜单" noauth:"true"`
	Code          string `name:"菜单编码"`
	Name          string `name:"菜单名称"`
	URI           string `name:"菜单地址"`
	Icon          string `name:"菜单图标"`
	Pid           string `name:"父菜单ID"`
	Group         string `name:"菜单组名"`
	Disable       bool   `name:"是否禁用"`
	IsLink        bool   `name:"是否外链"`
	IsVirtual     bool   `name:"是否虚菜单"`
	Sort          int    `name:"排序号"`
	IsBuiltIn     bool   `name:"是否系统内置"`
	IsDefaultOpen bool   `name:"是否默认打开"`
	Closeable     bool   `name:"是否可关闭"`
	Type          string `name:"菜单类型" dict:"sys_menu_type"`
	// 标准字段
	CreatedBy User   `name:"创建人" join:"User<Username,Name>"`
	CreatedAt int64  `name:"创建时间"`
	UpdatedBy User   `name:"修改人" join:"User<Username,Name>"`
	UpdatedAt int64  `name:"修改时间"`
	IsDeleted bool   `name:"是否已删除"`
	Remark    string `name:"备注"`
}

// BeforeFind 查询前钩子
func (menu *Menu) BeforeFind(scope *mongo.Scope) error {
	scopeCache := *scope.Cache
	routineCache := kuu.GetGoroutineCache()
	loginUID := routineCache["LoginUID"]
	loginOrgID := routineCache["LoginOrgID"]

	if scope.Operation == "List" && loginUID != "" && loginUID != kuu.Data["RootUID"] {
		uid := loginUID.(string)
		orgID := loginOrgID.(string)
		var rule AuthRule
		AuthRule := kuu.Model("AuthRule")
		AuthRule.One(kuu.H{
			"Cond": kuu.H{
				"UID":   uid,
				"OrgID": orgID,
			},
		}, &rule)
		permissions := rule.Permissions
		menusIDs := make([]bson.ObjectId, 0)
		for _, item := range permissions {
			menusIDs = append(menusIDs, bson.ObjectIdHex(item))
		}
		scopeCache["rawCond"] = scope.Params.Cond
		cond := kuu.H{
			"_id": kuu.H{
				"$in": menusIDs,
			},
		}
		scope.Params.Cond = cond
	}
	return nil
}

// AfterFind 查询后钩子
func (menu *Menu) AfterFind(scope *mongo.Scope) error {
	scopeCache := *scope.Cache
	if scopeCache != nil && scopeCache["rawCond"] != nil {
		scope.ListData["cond"] = scopeCache["rawCond"]
	}
	return nil
}
