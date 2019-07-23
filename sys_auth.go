package kuu

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	DataScopePersonal         = "PERSONAL"
	DataScopeCurrent          = "CURRENT"
	DataScopeCurrentFollowing = "CURRENT_FOLLOWING"
)

func init() {
	Enum("DataScope", "数据范围定义").
		Add(DataScopePersonal, "个人范围").
		Add(DataScopeCurrent, "当前组织").
		Add(DataScopeCurrentFollowing, "当前及以下组织")
}

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
		if len(v) > 0 {
			if err := RedisClient.Del(v...).Err(); err != nil {
				ERROR(err)
			} else {
				INFO("清空权限缓存 ALL")
			}
		}
	}
}

// DelPrisCacheBySign 清除用户缓存
func DelPrisCacheBySign(sign *SignContext) {
	if v := RedisClient.Get(RedisUserRolesKey(sign)).Val(); v != "" {
		if err := RedisClient.Del(RedisUserPrisKey(sign, v)).Err(); err != nil {
			ERROR(err)
		}
	}
	if err := RedisClient.Del(RedisUserRolesKey(sign)).Err(); err != nil {
		ERROR(err)
	} else {
		if sign.UID != 0 {
			INFO("清空权限缓存：UID=%d", sign.UID)
			return
		}
	}
}

// DelCurPrisCache 清除当前用户缓存
func DelCurPrisCache() {
	if v, ok := GetGLSValue(GLSSignInfoKey); ok {
		DelPrisCacheBySign(v.(*SignContext))
	}
}

// PrivilegesDesc
type PrivilegesDesc struct {
	UID              uint
	Codes            []string
	Permissions      map[string]int64
	ReadableOrgIDs   []uint
	ReadableOrgIDMap map[uint]bool
	WritableOrgIDs   []uint
	WritableOrgIDMap map[uint]bool
	Valid            bool
	SignOrgID        uint
	SignInfo         *SignContext
}

// IsWritableOrgID
func (desc *PrivilegesDesc) IsWritableOrgID(orgID uint) bool {
	return desc.WritableOrgIDMap[orgID] == true
}

// IsReadableOrgID
func (desc *PrivilegesDesc) IsReadableOrgID(orgID uint) bool {
	return desc.ReadableOrgIDMap[orgID] == true
}

// IsValid
func (desc *PrivilegesDesc) IsValid() bool {
	return desc != nil && desc.Valid && desc.SignInfo != nil && desc.SignInfo.IsValid()
}

// NotRootUser
func (desc *PrivilegesDesc) NotRootUser() bool {
	return desc.IsValid() && desc.UID != RootUID()
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
	// 重新计算
	user, err := GetUserWithRoles(sign.UID)
	if err != nil {
		//ERROR(err)
		return
	}
	desc = &PrivilegesDesc{
		UID:         sign.UID,
		Permissions: make(map[string]int64),
		Valid:       true,
	}
	type orange struct {
		readable string
		writable string
	}
	roleIDs := make([]string, 0)
	orm := make(map[uint]*orange)
	vmap := map[string]int{
		DataScopePersonal:         1,
		DataScopeCurrent:          2,
		DataScopeCurrentFollowing: 3,
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
					readableOrgIDs[child.ID] = true
				}
			}
		}
		// 统计可写
		if vmap[orgRange.writable] == 2 {
			writableOrgIDs[orgID] = true
		} else if vmap[orgRange.writable] == 3 {
			for _, child := range orgList {
				if strings.HasPrefix(child.FullPid, org.FullPid) {
					writableOrgIDs[child.ID] = true
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
	desc.ReadableOrgIDMap = readableOrgIDs
	desc.ReadableOrgIDs = keys(readableOrgIDs)
	desc.WritableOrgIDMap = writableOrgIDs
	desc.WritableOrgIDs = keys(writableOrgIDs)

	desc.SignOrgID = sign.OrgID
	desc.SignInfo = sign
	return
}
