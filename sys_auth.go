package kuu

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"strconv"
	"strings"
)

const (
	// DataScopePersonal
	DataScopePersonal = "PERSONAL"
	// DataScopeCurrent
	DataScopeCurrent = "CURRENT"
	// DataScopeCurrentFollowing
	DataScopeCurrentFollowing = "CURRENT_FOLLOWING"
)

// ActiveAuthProcessor
var ActiveAuthProcessor = DefaultAuthProcessor{}

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

// AuthProcessor
type AuthProcessor interface {
	AllowCreate(AuthProcessorDesc) error
	AddWritableWheres(AuthProcessorDesc) error
	AddReadableWheres(AuthProcessorDesc) error
}

// AuthProcessorDesc
type AuthProcessorDesc struct {
	Meta                *Metadata
	SubDocIDNames       []string
	Scope               *gorm.Scope
	PrisDesc            *PrivilegesDesc
	HasCreatedByIDField bool
	HasOrgIDField       bool
	CreatedByIDField    *gorm.Field
	OrgIDFieldField     *gorm.Field
	CreatedByID         uint
	OrgID               uint
}

// GetAuthProcessorDesc
func GetAuthProcessorDesc(scope *gorm.Scope, desc *PrivilegesDesc) (auth AuthProcessorDesc) {
	auth.Scope = scope
	auth.PrisDesc = desc
	if scope.Value != nil {
		auth.Meta = Meta(scope.Value)
		auth.SubDocIDNames = auth.Meta.SubDocIDNames

		if field, ok := scope.FieldByName("CreatedByID"); ok {
			auth.CreatedByIDField = field
			auth.HasCreatedByIDField = ok
		}
		if field, ok := scope.FieldByName("OrgID"); ok {
			auth.OrgIDFieldField = field
			auth.HasOrgIDField = ok
		}
	}
	return
}

// InjectCreateAuth
var InjectCreateAuth = func(signType string, auth AuthProcessorDesc) (replace bool, err error) {
	return
}

// InjectWritableAuth
var InjectWritableAuth = func(signType string, auth AuthProcessorDesc) (replace bool, err error) {
	return
}

// InjectReadableAuth
var InjectReadableAuth = func(signType string, auth AuthProcessorDesc) (replace bool, err error) {
	return
}

// DefaultAuthProcessor
type DefaultAuthProcessor struct{}

// AllowCreate
func (por *DefaultAuthProcessor) AllowCreate(auth AuthProcessorDesc) (err error) {
	if auth.PrisDesc.IsValid() {
		signType := auth.PrisDesc.SignInfo.Type
		if replace, custErr := InjectCreateAuth(signType, auth); custErr != nil {
			return custErr
		} else if replace {
			return
		}
		desc := auth.PrisDesc
		if desc.SignInfo.SubDocID != 0 && len(auth.SubDocIDNames) > 0 {
			// 基于扩展档案ID的数据权限
			if auth.HasCreatedByIDField && auth.CreatedByID != desc.UID {
				return fmt.Errorf("用户 %d 只拥有个人可写权限", desc.UID)
			}
		} else {
			// 基于组织的数据权限
			if auth.OrgID == 0 {
				if auth.HasCreatedByIDField && auth.CreatedByID != desc.UID {
					return fmt.Errorf("用户 %d 只拥有个人可写权限", desc.UID)
				}
			} else if auth.HasOrgIDField && !desc.IsWritableOrgID(auth.OrgID) {
				return fmt.Errorf("用户 %d 在组织 %d 中无可写权限", desc.UID, auth.OrgID)
			}
		}
	}
	return
}

// AddWritableWheres
func (por *DefaultAuthProcessor) AddWritableWheres(auth AuthProcessorDesc) (err error) {
	if auth.PrisDesc.IsValid() {
		signType := auth.PrisDesc.SignInfo.Type
		if replace, custErr := InjectWritableAuth(signType, auth); custErr != nil {
			return custErr
		} else if replace {
			return
		}
		sqls, attrs := GetDataScopeWheres(auth.Scope, auth.PrisDesc, auth.PrisDesc.WritableOrgIDs)
		if len(sqls) > 0 {
			auth.Scope.Search.Where(strings.Join(sqls, " OR "), attrs...)
		}
	}
	return
}

// AddReadableWheres
func (por *DefaultAuthProcessor) AddReadableWheres(auth AuthProcessorDesc) (err error) {
	if auth.PrisDesc.IsValid() {
		signType := auth.PrisDesc.SignInfo.Type
		if replace, custErr := InjectReadableAuth(signType, auth); custErr != nil {
			return custErr
		} else if replace {
			return
		}
		sqls, attrs := GetDataScopeWheres(auth.Scope, auth.PrisDesc, auth.PrisDesc.ReadableOrgIDs)
		if len(sqls) > 0 {
			auth.Scope.Search.Where(strings.Join(sqls, " OR "), attrs...)
		}
	}
	return
}

// GetDataScopeWheres
func GetDataScopeWheres(scope *gorm.Scope, desc *PrivilegesDesc, orgIDs []uint) (sqls []string, attrs []interface{}) {
	if scope.Value == nil || !desc.IsValid() {
		return
	}
	meta := Meta(scope.Value)
	caches := GetRoutineCaches()
	if caches != nil {
		// 有忽略标记时
		if _, ignoreAuth := caches[GLSIgnoreAuthKey]; ignoreAuth {
			return
		}
		// 查询用户菜单时
		if _, queryUserMenus := caches[GLSUserMenusKey]; queryUserMenus {
			if desc.NotRootUser() {
				_, hasCodeField := scope.FieldByName("Code")
				_, hasCreatedByIDField := scope.FieldByName("CreatedByID")
				if hasCodeField && hasCreatedByIDField {
					// 菜单数据权限控制与组织无关，且只有两种情况：
					// 1.自己创建的，一定看得到
					// 2.别人创建的，必须通过分配操作权限才能看到
					scope.Search.Where("(code in (?)) OR (created_by_id = ?)", desc.Codes, desc.UID)
				}
			}
			return
		}
	}

	subDocIDNames := meta.SubDocIDNames
	if desc.SignInfo.SubDocID != 0 && len(subDocIDNames) > 0 {
		// 基于扩展档案ID的数据权限
		for _, name := range subDocIDNames {
			if f, ok := scope.FieldByName(name); ok {
				sqls = append(sqls, "("+f.DBName+" = ?)")
				attrs = append(attrs, desc.SignInfo.SubDocID)
			}
		}
	} else {
		// 基于组织的数据权限
		if f, ok := scope.FieldByName("OrgID"); ok && len(orgIDs) > 0 {
			sqls = append(sqls, "("+f.DBName+" in (?))")
			attrs = append(attrs, orgIDs)
		} else {
			if f, ok := scope.FieldByName("CreatedByID"); ok {
				sqls = append(sqls, "("+f.DBName+" = ?)")
				attrs = append(attrs, desc.UID)
			}
		}
		if names := meta.OrgIDNames; len(names) > 0 {
			for _, name := range names {
				if f, ok := scope.FieldByName(name); ok {
					sqls = append(sqls, "("+f.DBName+" in (?))")
					attrs = append(attrs, orgIDs)
				}
			}
		}
		if names := meta.UIDNames; len(names) > 0 {
			for _, name := range names {
				if f, ok := scope.FieldByName(name); ok {
					sqls = append(sqls, "("+f.DBName+" = ?)")
					attrs = append(attrs, desc.UID)
				}
			}
		}
	}
	if meta.Name == "User" {
		sqls = append(sqls, "id = ?")
		attrs = append(attrs, desc.UID)
	}
	return
}

// CountWheres
func CountWheres(valueOrName interface{}, db *gorm.DB) *gorm.DB {
	var (
		meta  = Meta(valueOrName)
		scope = db.NewScope(meta.NewValue())
		desc  = GetRoutinePrivilegesDesc()
	)
	if desc != nil {
		sqls, attrs := GetDataScopeWheres(scope, desc, desc.ReadableOrgIDs)
		if len(sqls) > 0 {
			db = db.Where(strings.Join(sqls, " OR "), attrs...)
		}
	}
	return db
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
