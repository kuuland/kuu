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

// SetPrisCache
func SetPrisCache(sign *SignContext, desc *PrivilegesDesc, roleIDs []string) {
	roleIDsStr := strings.Join(roleIDs, ",")
	value := Stringify(desc)
	// 添加缓存
	if err := RedisClient.SetNX(RedisUserRolesKey(sign), roleIDsStr, time.Second*time.Duration(ExpiresSeconds)*7).Err(); err != nil {
		ERROR("用户角色缓存失败：%s", err.Error())
	} else {
		if err := RedisClient.SetNX(RedisUserPrisKey(sign, roleIDsStr), value, time.Second*time.Duration(ExpiresSeconds)*7).Err(); err != nil {
			ERROR("用户权限缓存失败：%s", err.Error())
			RedisClient.Del(RedisUserRolesKey(sign))
		} else {
			INFO("设置权限缓存：UID=%d", sign.UID)
		}
	}
}

// GetPrisCache
func GetPrisCache(sign *SignContext) (desc *PrivilegesDesc) {
	if v := RedisClient.Get(RedisUserRolesKey(sign)).Val(); v != "" {
		if v := RedisClient.Get(RedisUserPrisKey(sign, v)).Val(); v != "" {
			Parse(v, &desc)
			if desc.UID != 0 {
				INFO("获取权限缓存：UID=%d", sign.UID)
				return
			}
		}
	}
	return
}

// DelPrisCache
func DelPrisCache() {
	if v, err := RedisClient.Keys(RedisKeyBuilder("privileges*")).Result(); err == nil {
		if err := RedisClient.Del(v...).Err(); err != nil {
			ERROR(err)
		} else {
			INFO("清空权限缓存")
		}
	}
}

// PrivilegesDesc
type PrivilegesDesc struct {
	UID            uint
	OrgID          uint
	Codes          []string
	Permissions    map[string]int64
	ReadableOrgIDs []uint
	WritableOrgIDs []uint
}

// LoginOrgFilter
func LoginOrgFilter(desc *PrivilegesDesc, sign *SignContext) {
	if desc == nil || sign == nil || sign.OrgID == 0 {
		return
	}

	if C().DefaultGetBool("login:org:filter", true) {
		readableOrgIDs := make([]uint, 0)
		for _, item := range desc.ReadableOrgIDs {
			if item == sign.OrgID {
				readableOrgIDs = append(readableOrgIDs, item)
			}
		}
		writableOrgIDs := make([]uint, 0)
		for _, item := range desc.WritableOrgIDs {
			if item == sign.OrgID {
				writableOrgIDs = append(writableOrgIDs, item)
			}
		}
		desc.ReadableOrgIDs = readableOrgIDs
		desc.WritableOrgIDs = writableOrgIDs
	}
}

// GetPrivilegesDesc
func GetPrivilegesDesc(c *gin.Context) (desc *PrivilegesDesc) {
	if c == nil {
		return
	}

	sign := GetSignContext(c)
	if sign == nil {
		return
	}

	// 从缓存取
	if desc = GetPrisCache(sign); desc != nil {
		desc.OrgID = sign.OrgID
		LoginOrgFilter(desc, sign)
		return
	}
	// 重新计算
	user, err := GetUserRoles(c, sign.UID)
	if err != nil {
		//ERROR(err)
		return
	}
	desc = &PrivilegesDesc{
		UID:         sign.UID,
		OrgID:       sign.OrgID,
		Permissions: make(map[string]int64),
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
				or = &orange{
					readable: dp.ReadableRange,
					writable: dp.WritableRange,
				}
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
	var orgList []Org
	if err := DB().Find(&orgList).Error; err != nil {
		ERROR("组织列表查询失败")
		return
	}
	orgList = FillOrgFullInfo(orgList)
	orgMap := OrgIDMap(orgList)

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

	for code, _ := range desc.Permissions {
		desc.Codes = append(desc.Codes, code)
	}
	desc.ReadableOrgIDs = keys(readableOrgIDs)
	desc.WritableOrgIDs = keys(writableOrgIDs)

	// 添加缓存
	SetPrisCache(sign, desc, roleIDs)
	LoginOrgFilter(desc, sign)
	return
}
