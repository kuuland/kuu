package kuu

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

func RedisPrisKey(sign *SignContext) string {
	return RedisKeyBuilder("privileges", strconv.Itoa(int(sign.UID)), strconv.Itoa(int(sign.OrgID)))
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
	//key := RedisPrisKey(sign)
	//if v, err := RedisClient.Get(key).Result(); err == nil {
	//	Parse(v, &secret)
	//}

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
	return
}
