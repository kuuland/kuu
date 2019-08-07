package kuu

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
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
	SignInfo         *SignContext
	ActOrgID         uint
	ActOrgCode       string
	ActOrgName       string
	RolesCode        []string
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
		SignInfo:    sign,
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
		desc.RolesCode = append(desc.RolesCode, assign.Role.Code)
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
	// 计算ActOrgID
	var actOrg Org
	if user.ActOrgID != 0 && desc.ReadableOrgIDMap[user.ActOrgID] {
		actOrg = orgMap[user.ActOrgID]
	} else if len(desc.ReadableOrgIDs) > 0 {
		actOrg = orgMap[desc.ReadableOrgIDs[0]]
	}
	desc.ActOrgID = actOrg.ID
	desc.ActOrgCode = actOrg.Code
	desc.ActOrgName = actOrg.Name
	return
}
