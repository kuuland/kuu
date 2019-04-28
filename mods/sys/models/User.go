package models

import (
	"github.com/globalsign/mgo/bson"
	"github.com/kuuland/kuu"
	"github.com/kuuland/kuu/mods/mongo"
	"github.com/kuuland/kuu/mods/sys/utils"
	"time"
)

// User 系统用户
type User struct {
	ID          string       `json:"_id" displayName:"系统用户"`
	Username    string       `name:"账号"`
	Password    string       `name:"密码"`
	Name        string       `name:"姓名"`
	Avatar      string       `name:"头像"`
	Sex         int          `name:"性别"`
	Mobile      string       `name:"手机号"`
	Email       string       `name:"邮箱"`
	Language    string       `name:"语言"`
	Disable     bool         `name:"是否禁用"`
	RoleAssigns []RoleAssign `name:"角色分配"`
	IsBuiltIn   bool         `name:"是否系统内置"`
	// 标准字段
	CreatedBy interface{} `name:"创建人" join:"User<Username,Name>"`
	CreatedAt int64       `name:"创建时间"`
	UpdatedBy interface{} `name:"修改人" join:"User<Username,Name>"`
	UpdatedAt int64       `name:"修改时间"`
	IsDeleted bool        `name:"是否已删除"`
	Remark    string      `name:"备注"`
}

// RoleAssign 用户角色分配
type RoleAssign struct {
	RoleID     string `name:"角色ID"`
	ExpiryUnix int64  `name:"过期时间"`
}

// BeforeCreate 新增前钩子
func (u *User) BeforeCreate(scope *mongo.Scope) (err error) {
	arr := *scope.CreateData
	for index, item := range arr {
		var doc kuu.H
		kuu.JSONConvert(item, &doc)
		doc["Password"] = utils.GenerateFromPassword(doc["Password"].(string))
		arr[index] = doc
	}
	return
}

// BeforeUpdate 更新前钩子
func (u *User) BeforeUpdate(scope *mongo.Scope) (err error) {
	doc := *scope.UpdateDoc
	_set := doc["$set"].(kuu.H)
	if _set["Password"] != nil {
		_set["Password"] = utils.GenerateFromPassword(_set["Password"].(string))
	}
	if _set["RoleAssigns"] != nil {
		cache := *scope.Cache
		cache["UpdateRules"] = true
	}
	doc["$set"] = _set
	return
}

// AfterUpdate 更新后钩子
func (u *User) AfterUpdate(scope *mongo.Scope) (err error) {
	scopeCache := *scope.Cache
	routineCache := kuu.GetGoroutineCache()
	loginUID := routineCache["LoginUID"]

	if scopeCache["UpdateRules"] != nil && loginUID != kuu.Data["RootUID"] {
		if v, ok := loginUID.(string); ok && v != "" {
			UpdateAuthRules(v)
		}
	}
	return
}

// GetUserRoles 查询用户有效角色
func GetUserRoles(uid string) (roles []Role, user User) {
	// 查询用户档案
	User := kuu.Model("User")
	User.ID(uid, &user)
	if user.ID == "" {
		return
	}
	// 过滤有效的角色分配
	validAssigns := []RoleAssign{}
	roleIDs := []bson.ObjectId{}

	if user.RoleAssigns != nil {
		for _, assign := range user.RoleAssigns {
			if assign.ExpiryUnix <= 0 || time.Now().Before(time.Unix(assign.ExpiryUnix, 0)) {
				validAssigns = append(validAssigns, assign)
				roleIDs = append(roleIDs, bson.ObjectIdHex(assign.RoleID))
			}
		}
	}
	if len(validAssigns) == 0 {
		return
	}
	// 查询角色档案
	Role := kuu.Model("Role")
	Role.List(kuu.H{
		"Cond": kuu.H{
			"_id": kuu.H{
				"$in": roleIDs,
			},
		},
	}, &roles)
	return
}

// GetUserPermissions 查询用户权限
func GetUserPermissions(uid string, roles []Role) []string {
	if roles == nil {
		roles, _ = GetUserRoles(uid)
	}
	permissions := []string{}
	if roles != nil {
		for _, role := range roles {
			if role.OperationPrivileges == nil || len(role.OperationPrivileges) == 0 {
				continue
			}
			for _, privilege := range role.OperationPrivileges {
				permissions = append(permissions, privilege.Permission)
			}
		}
	}
	return permissions
}

// GetUserOrgs 查询用户组织
func GetUserOrgs(uid string, roles []Role) (orgs []Org) {
	if roles == nil {
		roles, _ = GetUserRoles(uid)
	}
	if roles == nil {
		return
	}
	// 提取组织ID
	orgIDs := make([]bson.ObjectId, 0)
	for _, role := range roles {
		if role.DataPrivileges != nil {
			for _, item := range role.DataPrivileges {
				if item.OrgID == "" {
					continue
				}
				orgIDs = append(orgIDs, bson.ObjectIdHex(item.OrgID))
			}
		}
	}
	// 查询组织列表
	Org := kuu.Model("Org")
	Org.List(kuu.H{
		"cond": kuu.H{
			"_id": kuu.H{
				"$in": orgIDs,
			},
		},
	}, &orgs)
	return
}
