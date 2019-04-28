package models

import (
	"fmt"
	"github.com/kuuland/kuu"
	"strings"
)

// AuthRule 授权规则
// 自动维护时间：
// 	1.角色编辑
// 	1.用户角色授权变更
// 	2.用户角色授权自动过期
type AuthRule struct {
	ID                string      `json:"_id" displayName:"授权规则" noauth:"true"`
	UID               string      `name:"用户ID"`
	Username          string      `name:"用户账号"`
	Name              string      `name:"用户姓名"`
	OrgID             string      `name:"登录组织"`
	OrgName           string      `name:"登录组织"`
	ObjectName        string      `name:"访问实体名称"`
	ObjectDisplayName string      `name:"访问实体显示名称"`
	ReadableScope     string      `name:"可读范围" dict:"sys_data_range"`
	WritableScope     string      `name:"可写范围" dict:"sys_data_range"`
	ReadableOrgIDs    []string    `name:"可读数据的组织ID"`
	WritableOrgIDs    []string    `name:"可写数据的组织ID"`
	HitAssign         interface{} `name:"命中规则"`
	Permissions       []string    `name:"可访问操作权限"`
}

// UpdateAllAuthRules 更新所有用户的授权规则
func UpdateAllAuthRules() {
	UserModel := kuu.Model("User")
	var list []User
	UserModel.List(kuu.H{
		"Project": map[string]int{
			"_id":      1,
			"Username": 1,
		},
	}, &list)
	if list != nil {
		for _, user := range list {
			UpdateAuthRules(user.ID)
		}
	}
	kuu.Info("所有用户的授权规则已更新")
}

// UpdateAuthRules 生成并保存授权规则
func UpdateAuthRules(uid string) {
	if uid == "" {
		return
	}
	// 查询用户角色列表
	roles, user := GetUserRoles(uid)
	// 查询用户权限列表
	permissions := GetUserPermissions(uid, roles)
	if roles == nil || len(roles) == 0 {
		return
	}
	// 查询用户组织列表
	orgs := GetUserOrgs(uid, roles)
	// 查询所有组织
	OrgModel := kuu.Model("Org")
	var totalOrgs []Org
	OrgModel.List(kuu.H{
		"Project": map[string]int{
			"_id":         1,
			"FullPathPid": 1,
		},
	}, &totalOrgs)
	// 构建规则列表
	orgPrivilegesMap := getOrgPrivilegesMap(roles)
	rules := []AuthRule{}
	for _, org := range orgs {
		//【直接授权】：直接针对组织的授权称为“直接授权”；
		//【间接授权】：在授权上级组织时选择了“当前及以下组织”而获得的称为“间接授权”。
		// 1.首先取直接授权
		// 2.若无直接授权，沿着组织树向上查询最近的一个“current_following”授权
		privilege := orgPrivilegesMap[org.ID]
		privilegeGetter := func(callback func(DataPrivileges) bool) {
			pids := strings.Split(org.FullPathPid, ",")
			for _, pid := range pids {
				if pid == org.ID {
					continue
				}
				p := orgPrivilegesMap[pid]
				if p.OrgID != "" && callback(p) {
					return
				}
			}
		}
		if privilege.OrgID == "" {
			privilegeGetter(func(p DataPrivileges) bool {
				if p.AllWritableRange == "current_following" || p.AllWritableRange == "current_following" {
					privilege = p
					return true
				}
				return false
			})
		}
		if privilege.AllReadableRange == "" {
			privilegeGetter(func(p DataPrivileges) bool {
				if p.AllReadableRange == "current_following" {
					privilege = p
					return true
				}
				return false
			})
		}
		if privilege.AllWritableRange == "" {
			privilegeGetter(func(p DataPrivileges) bool {
				if p.AllWritableRange == "current_following" {
					privilege = p
					return true
				}
				return false
			})
		}
		if privilege.OrgID == "" {
			continue
		}
		authObjectsMaps := map[string]AuthObject{}
		if privilege.AuthObjects != nil {
			for _, authObject := range privilege.AuthObjects {
				authObjectsMaps[authObject.Name] = authObject
			}
		}
		for name, schema := range kuu.Schemas {
			if schema.NoAuth {
				continue
			}
			authObject := authObjectsMaps[name]
			if authObject.Name == "" {
				authObject.Name = schema.Name
				authObject.DisplayName = schema.DisplayName
				authObject.ObjReadableRange = privilege.AllReadableRange
				authObject.ObjWritableRange = privilege.AllWritableRange
			}
			if authObject.ObjReadableRange == "" {
				authObject.ObjReadableRange = privilege.AllReadableRange
			}
			if authObject.ObjWritableRange == "" {
				authObject.ObjWritableRange = privilege.AllWritableRange
			}
			authObjectsMaps[name] = authObject
			rule := AuthRule{
				UID:               uid,
				Username:          user.Username,
				Name:              user.Name,
				OrgID:             org.ID,
				OrgName:           org.Name,
				ObjectName:        authObject.Name,
				ObjectDisplayName: authObject.DisplayName,
				ReadableScope:     authObject.ObjReadableRange,
				WritableScope:     authObject.ObjWritableRange,
				ReadableOrgIDs:    []string{},
				WritableOrgIDs:    []string{},
				HitAssign:         authObject,
				Permissions:       permissions,
			}
			switch rule.ReadableScope {
			case "current":
				rule.ReadableOrgIDs = append(rule.ReadableOrgIDs, rule.OrgID)
			case "current_following":
				for _, childOrg := range totalOrgs {
					if strings.HasPrefix(childOrg.FullPathPid, org.FullPathPid) {
						rule.ReadableOrgIDs = append(rule.ReadableOrgIDs, childOrg.ID)
					}
				}
			}
			switch rule.WritableScope {
			case "current":
				rule.WritableOrgIDs = append(rule.WritableOrgIDs, rule.OrgID)
			case "current_following":
				for _, childOrg := range totalOrgs {
					if strings.HasPrefix(childOrg.FullPathPid, org.FullPathPid) {
						rule.WritableOrgIDs = append(rule.WritableOrgIDs, childOrg.ID)
					}
				}
			}
			rules = append(rules, rule)
		}
	}
	AuthRule := kuu.Model("AuthRule")
	// 删除旧规则
	AuthRule.PhyRemoveAll(kuu.H{"UID": uid})
	// 新增旧规则
	AuthRule.Create(rules)
	kuu.Info(fmt.Sprintf("已更新用户 %v 的授权规则", uid))
}

func getOrgPrivilegesMap(roles []Role) map[string]DataPrivileges {
	groups := map[string]DataPrivileges{}
	for _, role := range roles {
		if role.DataPrivileges != nil {
			for _, privilege := range role.DataPrivileges {
				groups[privilege.OrgID] = privilege
			}
		}
	}
	return groups
}
