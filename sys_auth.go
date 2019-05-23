package kuu

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
	"time"
)

// RedisUserPrisKey
func RedisUserPrisKey(sign *SignContext, roleIDs string) string {
	return RedisKeyBuilder("privileges", strconv.Itoa(int(sign.UID)), roleIDs, strconv.Itoa(int(sign.OrgID)))
}

// RedisUserRolesKey
func RedisUserRolesKey(sign *SignContext) string {
	return RedisKeyBuilder("privileges_roles", strconv.Itoa(int(sign.UID)))
}

func setPrisCache(sign *SignContext, desc *PrivilegesDesc, roleIDs []string) {
	roleIDsStr := strings.Join(roleIDs, ",")
	value := Stringify(desc)
	// 添加缓存
	if !RedisClient.SetNX(RedisUserRolesKey(sign), roleIDsStr, time.Second*time.Duration(ExpiresSeconds)*7).Val() {
		ERROR("用户角色缓存失败")
	}
	if !RedisClient.SetNX(RedisUserPrisKey(sign, roleIDsStr), value, time.Second*time.Duration(ExpiresSeconds)*7).Val() {
		ERROR("用户权限缓存失败")
	}
}

func getPrisCache(sign *SignContext) (desc *PrivilegesDesc) {
	if v := RedisClient.Get(RedisUserRolesKey(sign)).Val(); v != "" {
		if v := RedisClient.Get(RedisUserPrisKey(sign, v)).Val(); v != "" {
			Parse(v, &desc)
			if desc.UID != 0 {
				return
			}
		}
	}
	return
}

func delPrisCache() {
	if v, err := RedisClient.Keys(RedisKeyBuilder("privileges*")).Result(); err == nil {
		if err := RedisClient.Del(v...).Err(); err != nil {
			ERROR(err)
		}
	}
}

// PrivilegesDesc
type PrivilegesDesc struct {
	UID            uint
	Permissions    map[string]int64
	ReadableOrgIDs []uint
	WritableOrgIDs []uint
}

// GetPrivilegesDesc
func GetPrivilegesDesc(c *gin.Context) (desc *PrivilegesDesc) {
	sign := GetSignContext(c)
	// 从缓存取
	if desc = getPrisCache(sign); desc != nil {
		return
	}
	// 重新计算
	user, err := GetUserRoles(c, sign.UID)
	if err != nil {
		ERROR(err)
		return
	}
	desc = &PrivilegesDesc{
		UID: sign.UID,
	}
	type orange struct {
		readable string
		writable string
	}
	roleIDs := make([]string, 0)
	orm := make(map[uint]*orange)
	vmap := map[string]int{
		"PERSONAL":          1,
		"CURRENT":           2,
		"CURRENT_FOLLOWING": 3,
	}
	for _, assign := range user.RoleAssigns {
		if assign.Role == nil {
			continue
		}
		roleIDs = append(roleIDs, strconv.Itoa(int(assign.Role.ID)))
		for _, op := range assign.Role.OperationPrivileges {
			if op.MenuCode != "" {
				desc.Permissions[op.MenuCode] = assign.ExpireUnix
			}
		}
		for _, dp := range assign.Role.DataPrivileges {
			if dp.TargetOrgID == 0 {
				continue
			}
			or := orm[dp.TargetOrgID]
			dp.ReadableRange = strings.ToUpper(dp.ReadableRange)
			dp.WritableRange = strings.ToUpper(dp.WritableRange)
			if or == nil {
				or.readable = dp.ReadableRange
				or.writable = dp.WritableRange
			} else {
				if vmap[dp.ReadableRange] > vmap[or.readable] {
					or.readable = dp.ReadableRange
				}
				if vmap[dp.WritableRange] > vmap[or.writable] {
					or.writable = dp.WritableRange
				}
			}
			orm[dp.TargetOrgID] = or
		}
	}
	var (
		orgList []Org
		orgMap  = make(map[uint]Org)
	)
	if err := DB().Find(&orgList).Error; err != nil {
		ERROR("组织列表查询失败")
		return
	}
	for _, org := range orgList {
		orgMap[org.ID] = org
	}

	var (
		readableOrgIDs = make(map[uint]bool)
		writableOrgIDs = make(map[uint]bool)
	)
	for orgID, orgRange := range orm {
		org := orgMap[orgID]
		// 统计可读
		if vmap[orgRange.readable] == 2 {
			readableOrgIDs[orgID] = true
		} else if vmap[orgRange.readable] == 3 {
			for _, child := range orgList {
				if strings.HasPrefix(child.FullPid, org.FullPid) {
					readableOrgIDs[orgID] = true
				}
			}
		}
		// 统计可写
		if vmap[orgRange.writable] == 2 {
			writableOrgIDs[orgID] = true
		} else if vmap[orgRange.writable] == 3 {
			for _, child := range orgList {
				if strings.HasPrefix(child.FullPid, org.FullPid) {
					writableOrgIDs[orgID] = true
				}
			}
		}
	}
	keys := func(m map[uint]bool) (a []uint) {
		for key, _ := range m {
			a = append(a, key)
		}
		return
	}
	desc.ReadableOrgIDs = keys(readableOrgIDs)
	desc.WritableOrgIDs = keys(writableOrgIDs)

	// 添加缓存
	setPrisCache(sign, desc, roleIDs)
	return
}
