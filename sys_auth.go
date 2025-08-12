package kuu

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"regexp"
	"sort"
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
	UID                      uint
	OrgID                    uint
	Permissions              []string
	PermissionMap            map[string]int64
	ReadableOrgIDs           []uint
	ReadableOrgIDMap         map[uint]Org
	FullReadableOrgIDs       []uint
	FullReadableOrgIDMap     map[uint]Org
	WritableOrgIDs           []uint
	WritableOrgIDMap         map[uint]Org
	PersonalReadableOrgIDs   []uint
	PersonalReadableOrgIDMap map[uint]Org
	PersonalWritableOrgIDs   []uint
	PersonalWritableOrgIDMap map[uint]Org
	LoginableOrgIDs          []uint
	LoginableOrgIDMap        map[uint]Org
	Valid                    bool
	SignInfo                 *SignContext
	ActOrgID                 uint
	ActOrgCode               string
	ActOrgName               string
	RolesCode                []string
}

// IsWritableOrgID
func (desc *PrivilegesDesc) IsWritableOrgID(orgID uint) bool {
	if v, has := desc.WritableOrgIDMap[orgID]; has && v.ID != 0 {
		return true
	}
	return false
}

// IsReadableOrgID
func (desc *PrivilegesDesc) IsReadableOrgID(orgID uint) bool {
	if v, has := desc.ReadableOrgIDMap[orgID]; has && v.ID != 0 {
		return true
	}
	return false
}

// IsLoginableOrgID
func (desc *PrivilegesDesc) IsLoginableOrgID(orgID uint) bool {
	if v, has := desc.LoginableOrgIDMap[orgID]; has && v.ID != 0 {
		return true
	}
	return false
}

// IsValid
func (desc *PrivilegesDesc) IsValid() bool {
	return desc != nil && desc.Valid && desc.SignInfo != nil && desc.SignInfo.IsValid()
}

// NotRootUser
func (desc *PrivilegesDesc) NotRootUser() bool {
	return desc.IsValid() && desc.UID != RootUID()
}

func (desc *PrivilegesDesc) addEmptyOrgs() {
	zeroOrg := Org{ID: 0}

	desc.ReadableOrgIDs = append(desc.ReadableOrgIDs, 0)
	desc.ReadableOrgIDMap[0] = zeroOrg

	desc.FullReadableOrgIDs = append(desc.FullReadableOrgIDs, 0)
	desc.FullReadableOrgIDMap[0] = zeroOrg

	desc.WritableOrgIDs = append(desc.WritableOrgIDs, 0)
	desc.WritableOrgIDMap[0] = zeroOrg

	desc.PersonalReadableOrgIDs = append(desc.PersonalReadableOrgIDs, 0)
	desc.PersonalReadableOrgIDMap[0] = zeroOrg

	desc.PersonalWritableOrgIDs = append(desc.PersonalWritableOrgIDs, 0)
	desc.PersonalWritableOrgIDMap[0] = zeroOrg
}

func (desc *PrivilegesDesc) HasPermission(permission string) bool {
	for _, s := range desc.Permissions {
		if permission == s {
			return true
		}
	}
	return false
}

func (desc *PrivilegesDesc) HasRole(code string) bool {
	for _, item := range desc.RolesCode {
		if item == code {
			return true
		}
	}
	return false
}

// APIPermissionItem API权限项
type APIPermissionItem struct {
	Method  []string `json:"method"`  // HTTP方法数组，如["GET", "POST"]
	Pattern string   `json:"pattern"` // URL匹配模式
}

// APIPermissionConfig API权限配置结构
type APIPermissionConfig struct {
	Enabled      bool     `json:"enabled"`       // 是否启用API权限验证
	IncludeUsers []string `json:"includeUsers"`  // 必须进行API权限验证的用户名列表
	ExcludeUsers []string `json:"excludeUsers"`  // 不需要进行API权限验证的用户名列表
	IncludeAll   bool     `json:"includeAll"`    // 是否对所有用户进行API权限验证
	ExcludeAll   bool     `json:"excludeAll"`    // 是否对所有用户都不进行API权限验证
}

// getAPIPermissionConfig 获取API权限配置
func (desc *PrivilegesDesc) getAPIPermissionConfig() APIPermissionConfig {
	// 获取API权限配置
	apiPermissionConfigStr := C().GetString("ApiPermissionConfig")
	if apiPermissionConfigStr != "" {
		var config APIPermissionConfig
		if err := json.Unmarshal([]byte(apiPermissionConfigStr), &config); err == nil {
			return config
		}
	}

	// 默认配置：不启用API权限验证
	return APIPermissionConfig{Enabled: false}
}

// shouldCheckAPIPermission 判断是否需要对用户进行API权限验证
func (desc *PrivilegesDesc) shouldCheckAPIPermission(config APIPermissionConfig, username string) bool {
	// 如果功能未启用，不需要验证
	if !config.Enabled {
		return false
	}

	// 如果ExcludeAll为true，所有用户都不需要验证
	if config.ExcludeAll {
		return false
	}

	// 如果用户在排除列表中，不需要验证
	for _, excludeUser := range config.ExcludeUsers {
		if excludeUser == username {
			return false
		}
	}

	// 如果IncludeAll为true，所有用户都需要验证（除了已经被排除的）
	if config.IncludeAll {
		return true
	}

	// 检查用户是否在包含列表中
	for _, includeUser := range config.IncludeUsers {
		if includeUser == username {
			return true
		}
	}

	// 默认不需要验证
	return false
}

// HasAPIPermission 检查用户是否有API访问权限
func (desc *PrivilegesDesc) HasAPIPermission(method, path string) bool {
	// 如果是root用户，直接允许
	if desc.UID == RootUID() {
		return true
	}

	// 获取API权限配置
	apiPermissionConfig := desc.getAPIPermissionConfig()
	
	// 如果未启用API权限验证，直接允许
	if !apiPermissionConfig.Enabled {
		return true
	}

	// 获取当前用户的用户名
	var currentUser User
	if err := DB().Select("username").Where("id = ?", desc.UID).First(&currentUser).Error; err != nil {
		return true // 查询用户失败，默认允许
	}

	// 检查是否需要进行API权限验证
	needCheck := desc.shouldCheckAPIPermission(apiPermissionConfig, currentUser.Username)

	if !needCheck {
		return true // 不需要验证，默认允许
	}

	// 获取用户有权限的菜单列表
	var menus []Menu
	IgnoreAuth()
	defer IgnoreAuth(true)

	if err := DB().Where("code IN (?)", desc.Permissions).Find(&menus).Error; err != nil {
		return false
	}

	// 检查每个菜单的API权限配置
	for _, menu := range menus {
		if menu.APIPattern.Valid && menu.APIPattern.String != "" {
			var apiItems []APIPermissionItem
			if err := json.Unmarshal([]byte(menu.APIPattern.String), &apiItems); err != nil {
				continue
			}

			for _, item := range apiItems {
				// 检查HTTP方法是否匹配
				methodMatch := false
				for _, allowedMethod := range item.Method {
					if allowedMethod == "*" || strings.ToUpper(allowedMethod) == strings.ToUpper(method) {
						methodMatch = true
						break
					}
				}

				if methodMatch {
					// 检查URL模式是否匹配
					if matched, _ := regexp.MatchString(item.Pattern, path); matched {
						return true
					}
				}
			}
		}
	}

	return false // 没有找到匹配的权限，拒绝访问
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
	if auth.Meta == nil || !auth.PrisDesc.IsValid() {
		return
	}
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
	return
}

// AddWritableWheres
func (por *DefaultAuthProcessor) AddWritableWheres(auth AuthProcessorDesc) (err error) {
	if auth.Meta == nil || !auth.PrisDesc.IsValid() {
		return
	}
	signType := auth.PrisDesc.SignInfo.Type
	if replace, custErr := InjectWritableAuth(signType, auth); custErr != nil {
		return custErr
	} else if replace {
		return
	}
	sqls, attrs := GetDataScopeWheres(auth.Scope, auth.PrisDesc, auth.PrisDesc.WritableOrgIDs, auth.PrisDesc.PersonalWritableOrgIDMap)
	if len(sqls) > 0 {
		auth.Scope.Search.Where(strings.Join(sqls, " OR "), attrs...)
	}
	return
}

// AddReadableWheres
func (por *DefaultAuthProcessor) AddReadableWheres(auth AuthProcessorDesc) (err error) {
	if auth.Meta == nil || !auth.PrisDesc.IsValid() {
		return
	}
	signType := auth.PrisDesc.SignInfo.Type
	if replace, custErr := InjectReadableAuth(signType, auth); custErr != nil {
		return custErr
	} else if replace {
		return
	}
	sqls, attrs := GetDataScopeWheres(auth.Scope, auth.PrisDesc, auth.PrisDesc.ReadableOrgIDs, auth.PrisDesc.PersonalWritableOrgIDMap)
	if len(sqls) > 0 {
		auth.Scope.Search.Where(strings.Join(sqls, " OR "), attrs...)
	}
	return
}

// GetDataScopeWheres
func GetDataScopeWheres(scope *gorm.Scope, desc *PrivilegesDesc, orgIDs []uint, personalOrgIDMap map[uint]Org) (sqls []string, attrs []interface{}) {
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
		if meta.Name == "Menu" {
			if desc.NotRootUser() {
				codeField, hasCodeField := scope.FieldByName("Code")
				createdByIDField, hasCreatedByIDField := scope.FieldByName("CreatedByID")
				if hasCodeField && hasCreatedByIDField {
					// 菜单数据权限控制与组织无关，且只有两种情况：
					// 1.自己创建的，一定看得到
					// 2.别人创建的，必须通过分配操作权限才能看到
					scope.Search.Where(fmt.Sprintf("(%v.%v IN (?)) OR (%v.%v = ?)",
						scope.QuotedTableName(),
						scope.Quote(codeField.DBName),
						scope.QuotedTableName(),
						scope.Quote(createdByIDField.DBName),
					), desc.Permissions, desc.UID)
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
				sqls = append(sqls, fmt.Sprintf("(%v.%v = ?)",
					scope.QuotedTableName(),
					scope.Quote(f.DBName),
				))
				attrs = append(attrs, desc.SignInfo.SubDocID)
			}
		}
	} else {
		// 基于组织的数据权限
		if orgIDField, has := scope.FieldByName("OrgID"); has && len(orgIDs) > 0 {
			dbName := orgIDField.DBName
			if meta.Name == "Org" {
				dbName = "id"
			}
			sqls = append(sqls, fmt.Sprintf("(%v.%v IN (?))",
				scope.QuotedTableName(),
				scope.Quote(dbName),
			))
			attrs = append(attrs, orgIDs)

			// 空组织
			sqls = append(sqls, fmt.Sprintf("(%v.%v IS NULL)",
				scope.QuotedTableName(),
				scope.Quote(dbName),
			))
		}
		if len(personalOrgIDMap) > 0 && personalOrgIDMap[desc.ActOrgID].ID != 0 {
			if f, ok := scope.FieldByName("CreatedByID"); ok {
				sqls = append(sqls, fmt.Sprintf("(%v.%v = ?)",
					scope.QuotedTableName(),
					scope.Quote(f.DBName),
				))
				attrs = append(attrs, desc.UID)
			}
			if len(meta.UIDNames) > 0 {
				for _, name := range meta.UIDNames {
					if f, ok := scope.FieldByName(name); ok {
						sqls = append(sqls, fmt.Sprintf("(%v.%v = ?)",
							scope.QuotedTableName(),
							scope.Quote(f.DBName),
						))
						attrs = append(attrs, desc.UID)
					}
				}
			}
		}
		if len(meta.OrgIDNames) > 0 {
			for _, name := range meta.OrgIDNames {
				if f, ok := scope.FieldByName(name); ok {
					sqls = append(sqls, fmt.Sprintf("(%v.%v IN (?))",
						scope.QuotedTableName(),
						scope.Quote(f.DBName),
					))
					attrs = append(attrs, orgIDs)
					// 空组织
					sqls = append(sqls, fmt.Sprintf("(%v.%v IS NULL)",
						scope.QuotedTableName(),
						scope.Quote(f.DBName),
					))
				}
			}
		}
	}
	if meta.Name == "User" {
		sqls = append(sqls, fmt.Sprintf("(%v.%v = ?)",
			scope.QuotedTableName(),
			scope.Quote("id"),
		))
		attrs = append(attrs, desc.UID)
	}
	return
}

// GetPrivilegesDesc
func GetPrivilegesDesc(signOrContextOrUID interface{}) (desc *PrivilegesDesc) {
	IgnoreAuth()
	defer IgnoreAuth(true)

	var (
		sign *SignContext
		uid  uint
	)
	if v, ok := signOrContextOrUID.(*Context); ok {
		sign = v.SignInfo
	} else if v, ok := signOrContextOrUID.(*SignContext); ok {
		sign = v
	} else if v, ok := signOrContextOrUID.(uint); ok {
		uid = v
	} else {
		ERROR("unsupported parameter: %v", signOrContextOrUID)
		return
	}
	if sign == nil && uid == 0 {
		return
	} else if sign != nil {
		uid = sign.UID
	}
	// 重新计算
	user, err := GetUserWithRoles(uid)
	if err != nil {
		return
	}
	desc = &PrivilegesDesc{
		UID:           uid,
		OrgID:         user.OrgID,
		PermissionMap: make(map[string]int64),
		Valid:         true,
		SignInfo:      sign,
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
	// 单独统计个人针对组织的可读写权限
	var (
		personalReadableOrgIDMap = make(map[uint]Org)
		personalWritableOrgIDMap = make(map[uint]Org)
	)
	for _, assign := range user.RoleAssigns {
		if assign.Role == nil {
			continue
		}
		desc.RolesCode = append(desc.RolesCode, assign.Role.Code)
		roleIDs = append(roleIDs, strconv.Itoa(int(assign.Role.ID)))
		for _, op := range assign.Role.OperationPrivileges {
			if op.MenuCode != "" {
				desc.PermissionMap[op.MenuCode] = assign.ExpireUnix
			}
		}
		for _, dp := range assign.Role.DataPrivileges {
			if dp.TargetOrgID == 0 {
				continue
			}
			or := orm[dp.TargetOrgID]
			dp.ReadableRange = strings.ToUpper(dp.ReadableRange)
			dp.WritableRange = strings.ToUpper(dp.WritableRange)

			if dp.ReadableRange == DataScopePersonal {
				personalReadableOrgIDMap[dp.TargetOrgID] = Org{ID: dp.TargetOrgID}
			}

			if dp.WritableRange == DataScopePersonal {
				personalWritableOrgIDMap[dp.TargetOrgID] = Org{ID: dp.TargetOrgID}
			}

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
		readableOrgIDMap  = make(map[uint]Org)
		writableOrgIDMap  = make(map[uint]Org)
		loginableOrgIDMap = make(map[uint]Org)
	)
	for orgID, orgRange := range orm {
		var org Org
		if v, has := orgMap[orgID]; !has {
			continue
		} else {
			org = v
		}
		loginableOrgIDMap[orgID] = org
		// 统计可读
		if vmap[orgRange.readable] == 2 {
			readableOrgIDMap[orgID] = org
		} else if vmap[orgRange.readable] == 3 {
			for _, child := range orgList {
				//if strings.HasPrefix(child.FullPid, org.FullPid) {
				if strings.HasPrefix(child.FullPid, org.FullPid) || child.FullPid == org.FullPid {
					readableOrgIDMap[child.ID] = org
				}
			}
		}
		// 统计可写
		if vmap[orgRange.writable] == 2 {
			writableOrgIDMap[orgID] = org
		} else if vmap[orgRange.writable] == 3 {
			for _, child := range orgList {
				if strings.HasPrefix(child.FullPid, org.FullPid) || child.FullPid == org.FullPid {
					writableOrgIDMap[child.ID] = org
				}
			}
		}
	}
	// 个人读写单独统计
	for orgID := range personalReadableOrgIDMap {
		personalReadableOrgIDMap[orgID] = orgMap[orgID]
	}
	for orgID := range personalWritableOrgIDMap {
		personalWritableOrgIDMap[orgID] = orgMap[orgID]
	}
	keys := func(m map[uint]Org) (a []uint) {
		for key, _ := range m {
			a = append(a, key)
		}
		return
	}

	for code := range desc.PermissionMap {
		desc.Permissions = append(desc.Permissions, code)
	}
	desc.FullReadableOrgIDMap = readableOrgIDMap
	desc.FullReadableOrgIDs = keys(readableOrgIDMap)
	desc.WritableOrgIDMap = writableOrgIDMap
	desc.WritableOrgIDs = keys(writableOrgIDMap)
	desc.LoginableOrgIDMap = loginableOrgIDMap
	desc.LoginableOrgIDs = keys(loginableOrgIDMap)

	desc.PersonalReadableOrgIDMap = personalReadableOrgIDMap
	desc.PersonalReadableOrgIDs = keys(personalReadableOrgIDMap)
	desc.PersonalWritableOrgIDMap = personalWritableOrgIDMap
	desc.PersonalWritableOrgIDs = keys(personalWritableOrgIDMap)

	// 排序
	sortIDs(&desc.FullReadableOrgIDs)
	sortIDs(&desc.WritableOrgIDs)
	sortIDs(&desc.LoginableOrgIDs)
	sortIDs(&desc.PersonalReadableOrgIDs)
	sortIDs(&desc.PersonalWritableOrgIDs)
	// 计算ActOrgID
	var actOrg Org
	if user.ActOrgID != 0 && desc.IsLoginableOrgID(user.ActOrgID) {
		actOrg = orgMap[user.ActOrgID]
	} else if len(desc.LoginableOrgIDs) > 0 {
		actOrg = orgMap[desc.LoginableOrgIDs[0]]
		// 取最顶级组织为默认值
		for _, orgID := range desc.LoginableOrgIDs {
			orgItem := orgMap[orgID]
			if len(orgItem.FullPid) < len(actOrg.FullPid) {
				actOrg = orgItem
			}
		}

	} else {
		actOrg = orgMap[desc.OrgID]
	}
	desc.ActOrgID = actOrg.ID
	desc.ActOrgCode = actOrg.Code
	desc.ActOrgName = actOrg.Name
	// 限制读取组织为当前组织或当前组织及以下（不能跨组织树分支或上级组织）
	filteredReadableOrgIDs := map[uint]Org{actOrg.ID: actOrg}
	for itemID, itemOrg := range desc.FullReadableOrgIDMap {
		if strings.HasPrefix(itemOrg.FullPid, actOrg.FullPid) {
			filteredReadableOrgIDs[itemID] = itemOrg
		}
	}
	desc.ReadableOrgIDMap = filteredReadableOrgIDs
	desc.ReadableOrgIDs = keys(filteredReadableOrgIDs)

	desc.addEmptyOrgs()

	return
}

func sortIDs(arr *[]uint) {
	newIDs := make([]int, len(*arr))
	for index, item := range *arr {
		newIDs[index] = int(item)
	}
	sort.Ints(newIDs)
	for index, item := range newIDs {
		(*arr)[index] = uint(item)
	}
}
