package kuu

import (
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/guregu/null.v3"
)

func init() {
	paramsUpdateKey := BuildKey("params", "update", "code")
	DefaultCache.Subscribe([]string{paramsUpdateKey}, func(key string, value string) {
		if key == paramsUpdateKey {
			C().LoadFromParams(value)
		}
	})
}

// Model
type Model struct {
	ID          uint        `gorm:"primary_key"`
	CreatedAt   time.Time   `name:"创建时间，ISO字符串（默认字段）"`
	UpdatedAt   time.Time   `name:"修改时间，ISO字符串（默认字段）"`
	DeletedAt   *time.Time  `name:"删除时间，ISO字符串（默认字段）" sql:"index"`
	OrgID       uint        `name:"所属组织ID（默认字段）"`
	CreatedByID uint        `name:"创建人ID（默认字段）"`
	UpdatedByID uint        `name:"修改人ID（默认字段）"`
	DeletedByID uint        `name:"删除人ID（默认字段）"`
	Remark      null.String `name:"备注" gorm:"text"`
	Ts          time.Time   `name:"时间戳"`
	Org         *Org        `gorm:"association_autoupdate:false;association_autocreate:false;foreignkey:id;association_foreignkey:org_id"`
	CreatedBy   *User       `gorm:"association_autoupdate:false;association_autocreate:false;foreignkey:id;association_foreignkey:created_by_id"`
	UpdatedBy   *User       `gorm:"association_autoupdate:false;association_autocreate:false;foreignkey:id;association_foreignkey:updated_by_id"`
	DeletedBy   *User       `gorm:"association_autoupdate:false;association_autocreate:false;foreignkey:id;association_foreignkey:deleted_by_id"`
}

// ModelExDel
type ModelExDel struct {
	ID          uint        `gorm:"primary_key"`
	CreatedAt   time.Time   `name:"创建时间，ISO字符串（默认字段）"`
	UpdatedAt   time.Time   `name:"修改时间，ISO字符串（默认字段）"`
	OrgID       uint        `name:"所属组织ID（默认字段）"`
	CreatedByID uint        `name:"创建人ID（默认字段）"`
	UpdatedByID uint        `name:"修改人ID（默认字段）"`
	Remark      null.String `name:"备注" gorm:"text"`
	Ts          time.Time   `name:"时间戳"`
	Org         *Org        `gorm:"association_autoupdate:false;association_autocreate:false;foreignkey:id;association_foreignkey:org_id"`
	CreatedBy   *User       `gorm:"association_autoupdate:false;association_autocreate:false;foreignkey:id;association_foreignkey:created_by_id"`
	UpdatedBy   *User       `gorm:"association_autoupdate:false;association_autocreate:false;foreignkey:id;association_foreignkey:updated_by_id"`
}

// ModelExOrg
type ModelExOrg struct {
	ID          uint        `gorm:"primary_key"`
	CreatedAt   time.Time   `name:"创建时间，ISO字符串（默认字段）"`
	UpdatedAt   time.Time   `name:"修改时间，ISO字符串（默认字段）"`
	DeletedAt   *time.Time  `name:"删除时间，ISO字符串（默认字段）" sql:"index"`
	CreatedByID uint        `name:"创建人ID（默认字段）"`
	UpdatedByID uint        `name:"修改人ID（默认字段）"`
	DeletedByID uint        `name:"删除人ID（默认字段）"`
	Remark      null.String `name:"备注" gorm:"text"`
	Ts          time.Time   `name:"时间戳"`
	CreatedBy   *User       `gorm:"association_autoupdate:false;association_autocreate:false;foreignkey:id;association_foreignkey:created_by_id"`
	UpdatedBy   *User       `gorm:"association_autoupdate:false;association_autocreate:false;foreignkey:id;association_foreignkey:updated_by_id"`
	DeletedBy   *User       `gorm:"association_autoupdate:false;association_autocreate:false;foreignkey:id;association_foreignkey:deleted_by_id"`
}

// ExtendField
type ExtendField struct {
	Def1 uint   `name:"扩展字段1（默认字段）"`
	Def2 string `name:"扩展字段2（默认字段）" gorm:"type:text"`
	Def3 string `name:"扩展字段3（默认字段）"`
	Def4 string `name:"扩展字段4（默认字段）"`
	Def5 string `name:"扩展字段5（默认字段）"`
}

// User
type User struct {
	// 引用Model将无法Preload，故复制字段
	ID          uint       `gorm:"primary_key" rest:"*" displayName:"用户"`
	CreatedAt   time.Time  `name:"创建时间，ISO字符串（默认字段）"`
	UpdatedAt   time.Time  `name:"修改时间，ISO字符串（默认字段）"`
	DeletedAt   *time.Time `name:"删除时间，ISO字符串（默认字段）" sql:"index"`
	Dr          int64      `name:"删除标记（0表示未删除，非0表示删除时间戳）" gorm:"DEFAULT:0;UNIQUE_INDEX:kuu_unique"`
	OrgID       uint       `name:"所属组织ID（默认字段）"`
	CreatedByID uint       `name:"创建人ID（默认字段）"`
	UpdatedByID uint       `name:"修改人ID（默认字段）"`
	DeletedByID uint       `name:"删除人ID（默认字段）"`
	Remark      string     `name:"备注" gorm:"text"`
	Ts          time.Time  `name:"时间戳"`
	Org         *Org       `gorm:"association_autoupdate:false;association_autocreate:false;foreignkey:id;association_foreignkey:org_id"`
	CreatedBy   *User      `gorm:"association_autoupdate:false;association_autocreate:false;foreignkey:id;association_foreignkey:created_by_id"`
	UpdatedBy   *User      `gorm:"association_autoupdate:false;association_autocreate:false;foreignkey:id;association_foreignkey:updated_by_id"`
	DeletedBy   *User      `gorm:"association_autoupdate:false;association_autocreate:false;foreignkey:id;association_foreignkey:deleted_by_id"`

	ExtendField
	Username    string       `name:"账号" gorm:"not null;UNIQUE_INDEX:kuu_unique"`
	Password    string       `name:"密码" gorm:"not null" json:",omitempty" kuu:"password"`
	Name        string       `name:"姓名"`
	Avatar      string       `name:"头像"`
	Sex         int          `name:"性别"`
	Mobile      string       `name:"手机号"`
	Email       string       `name:"邮箱地址"`
	Disable     null.Bool    `name:"是否禁用"`
	RoleAssigns []RoleAssign `name:"已分配角色"`
	IsBuiltIn   null.Bool    `name:"是否内置"`
	SubDocIDs   string       `name:"扩展档案ID映射（令牌类型为键、ID为值的JSON字符串，如{\"ADMIN\":3,\"WECHAT:COMPANY\":2}）"`
	Lang        string       `name:"最近使用语言"`
	DenyLogin   null.Bool    `name:"禁止登录"`
	ActOrgID    uint         `name:"当前组织"`
}

// GetSubDocIDs
func (u *User) GetSubDocIDs() (v map[string]uint, err error) {
	v = make(map[string]uint)
	if u.SubDocIDs == "" {
		u.SubDocIDs = "{}"
	}
	err = JSONParse(u.SubDocIDs, &v)
	return
}

// SetSubDocIDs
func (u *User) SetSubDocIDs(v map[string]uint) {
	u.SubDocIDs = JSONStringify(v)
}

// GetSubDocID
func (u *User) GetSubDocID(signType string) (uint, error) {
	m, err := u.GetSubDocIDs()
	if err != nil {
		return 0, err
	}
	return m[signType], nil
}

// SetSubDocIDs
func (u *User) SetSubDocID(signType string, subDocID uint) error {
	m, err := u.GetSubDocIDs()
	if err != nil {
		return err
	}
	if subDocID == 0 {
		delete(m, signType)
	} else {
		m[signType] = subDocID
	}
	u.SetSubDocIDs(m)
	return nil
}

// DelSubDocID
func (u *User) DelSubDocID(signType string) error {
	return u.SetSubDocID(signType, 0)
}

// BeforeSave
func (u *User) BeforeSave(scope *gorm.Scope) (err error) {
	if len(u.Password) == 32 {
		var hashed string
		if hashed, err = GenerateFromPassword(u.Password); err == nil {
			err = scope.SetColumn("Password", hashed)
		}
	}
	if u.ID != 0 {
		DelCache(fmt.Sprintf("user_%d", u.ID))
	}
	return
}

// AfterDelete
func (u *User) AfterDelete() {
	if u.ID != 0 {
		DelCache(fmt.Sprintf("user_%d", u.ID))
	}
}

// Org
type Org struct {
	// 引用Model将无法Preload，故复制字段
	ID          uint       `gorm:"primary_key" rest:"*" displayName:"用户"`
	CreatedAt   time.Time  `name:"创建时间，ISO字符串（默认字段）"`
	UpdatedAt   time.Time  `name:"修改时间，ISO字符串（默认字段）"`
	DeletedAt   *time.Time `name:"删除时间，ISO字符串（默认字段）" sql:"index"`
	Dr          int64      `name:"删除标记（0表示未删除，非0表示删除时间戳）" gorm:"DEFAULT:0;UNIQUE_INDEX:kuu_unique"`
	OrgID       uint       `name:"所属组织ID（默认字段）"`
	CreatedByID uint       `name:"创建人ID（默认字段）"`
	UpdatedByID uint       `name:"修改人ID（默认字段）"`
	DeletedByID uint       `name:"删除人ID（默认字段）"`
	Remark      string     `name:"备注" gorm:"text"`
	Ts          time.Time  `name:"时间戳"`
	Org         *Org       `gorm:"association_autoupdate:false;association_autocreate:false;foreignkey:id;association_foreignkey:org_id"`
	CreatedBy   *User      `gorm:"association_autoupdate:false;association_autocreate:false;foreignkey:id;association_foreignkey:created_by_id"`
	UpdatedBy   *User      `gorm:"association_autoupdate:false;association_autocreate:false;foreignkey:id;association_foreignkey:updated_by_id"`
	DeletedBy   *User      `gorm:"association_autoupdate:false;association_autocreate:false;foreignkey:id;association_foreignkey:deleted_by_id"`

	ExtendField
	Code      string `name:"组织编码" gorm:"not null;UNIQUE_INDEX:kuu_unique"`
	Name      string `name:"组织名称" gorm:"not null"`
	Pid       uint   `name:"父组织ID"`
	Sort      int    `name:"排序值"`
	FullPid   string
	FullName  string
	Class     string
	IsBuiltIn null.Bool `name:"是否内置"`
}

// BeforeCreate
func (o *Org) BeforeCreate() {
	if o.Pid == 0 {
		if desc := GetRoutinePrivilegesDesc(); desc.IsValid() && desc.ActOrgID != 0 {
			o.Pid = desc.ActOrgID
		} else if o.Code != "default" {
			o.Pid = RootOrgID()
		}
	}
}

// OrgIDMap
func OrgIDMap(list []Org) map[uint]Org {
	orgMap := make(map[uint]Org)
	for _, org := range list {
		orgMap[org.ID] = org
	}
	return orgMap
}

// FillOrgFullInfo
func FillOrgFullInfo(list []Org) []Org {
	var (
		childrenMap = make(map[uint][]Org)
		fullPidMap  = make(map[uint]string)
		fullNameMap = make(map[uint]string)
		fall        func([]Org, uint)
	)
	for _, org := range list {
		if org.Pid != org.ID {
			childrenMap[org.Pid] = append(childrenMap[org.Pid], org)
		}
	}
	fall = func(values []Org, pid uint) {
		for _, item := range values {
			if _, exists := fullPidMap[pid]; exists {
				fullPidMap[item.ID] = fmt.Sprintf("%s,%d,", fullPidMap[pid], item.ID)
				fullNameMap[item.ID] = fmt.Sprintf("%s,%s,", fullNameMap[pid], item.Name)
			} else {
				fullPidMap[item.ID] = fmt.Sprintf("%d,", item.ID)
				fullNameMap[item.ID] = fmt.Sprintf("%s,", item.Name)
			}
			children := childrenMap[item.ID]
			if len(children) > 0 {
				fall(children, item.ID)
			}
		}
	}
	fall(childrenMap[0], 0)
	for index, item := range list {
		item.FullPid = fullPidMap[item.ID]
		item.FullName = fullNameMap[item.ID]
		list[index] = item
	}
	return list
}

// RoleAssign
type RoleAssign struct {
	ModelExOrg `displayName:"用户角色分配"`
	ExtendField
	UserID     uint `name:"用户ID" gorm:"not null"`
	RoleID     uint `name:"角色ID" gorm:"not null"`
	Role       *Role
	ExpireUnix int64
}

// Role
type Role struct {
	// 引用Model将无法Preload，故复制字段
	ID          uint       `gorm:"primary_key" rest:"*" displayName:"角色"`
	CreatedAt   time.Time  `name:"创建时间，ISO字符串（默认字段）"`
	UpdatedAt   time.Time  `name:"修改时间，ISO字符串（默认字段）"`
	DeletedAt   *time.Time `name:"删除时间，ISO字符串（默认字段）" sql:"index"`
	Dr          int64      `name:"删除标记（0表示未删除，非0表示删除时间戳）" gorm:"DEFAULT:0;UNIQUE_INDEX:kuu_unique"`
	OrgID       uint       `name:"所属组织ID（默认字段）"`
	CreatedByID uint       `name:"创建人ID（默认字段）"`
	UpdatedByID uint       `name:"修改人ID（默认字段）"`
	DeletedByID uint       `name:"删除人ID（默认字段）"`
	Remark      string     `name:"备注" gorm:"text"`
	Ts          time.Time  `name:"时间戳"`
	Org         *Org       `gorm:"association_autoupdate:false;association_autocreate:false;foreignkey:id;association_foreignkey:org_id"`
	CreatedBy   *User      `gorm:"association_autoupdate:false;association_autocreate:false;foreignkey:id;association_foreignkey:created_by_id"`
	UpdatedBy   *User      `gorm:"association_autoupdate:false;association_autocreate:false;foreignkey:id;association_foreignkey:updated_by_id"`
	DeletedBy   *User      `gorm:"association_autoupdate:false;association_autocreate:false;foreignkey:id;association_foreignkey:deleted_by_id"`

	ExtendField
	Code                string                `name:"角色编码" gorm:"not null;UNIQUE_INDEX:kuu_unique"`
	Name                string                `name:"角色名称" gorm:"not null"`
	OperationPrivileges []OperationPrivileges `name:"角色操作权限"`
	DataPrivileges      []DataPrivileges      `name:"角色数据权限"`
	IsBuiltIn           null.Bool             `name:"是否内置"`
}

// OperationPrivileges
type OperationPrivileges struct {
	ModelExOrg `displayName:"角色操作权限"`
	ExtendField
	RoleID   uint   `name:"角色ID"`
	MenuCode string `name:"菜单编码"`
}

// DataPrivileges
type DataPrivileges struct {
	Model `rest:"*" displayName:"角色数据权限"`
	ExtendField
	RoleID        uint   `name:"角色ID"`
	TargetOrgID   uint   `name:"目标组织ID"`
	ReadableRange string `name:"可读范围" enum:"DataScope"`
	WritableRange string `name:"可写范围" enum:"DataScope"`
}

// BeforeSave
func (m *DataPrivileges) BeforeSave(scope *gorm.Scope) {
	_ = scope.SetColumn("OrgID", m.TargetOrgID)
}

// Menu
type Menu struct {
	ModelExOrg `rest:"*" displayName:"菜单"`
	ExtendField
	Code          string      `name:"菜单编码" gorm:"not null;UNIQUE_INDEX:kuu_unique"`
	Dr            int64       `name:"删除标记（0表示未删除，非0表示删除时间戳）" gorm:"DEFAULT:0;UNIQUE_INDEX:kuu_unique"`
	Name          string      `name:"菜单名称" gorm:"not null"`
	URI           null.String `name:"菜单地址"`
	Icon          null.String `name:"菜单图标"`
	ParentCode    null.String `name:"直接父菜单编号"`
	ParentCodes   null.String `name:"所有父菜单编号"`
	Group         null.String `name:"菜单分组名"`
	Disable       null.Bool   `name:"是否禁用"`
	IsLink        null.Bool   `name:"是否外链"`
	Sort          null.Int    `name:"排序值"`
	IsDefaultOpen null.Bool   `name:"是否默认打开"`
	Closeable     null.Bool   `name:"是否可关闭"`
	LocaleKey     null.String `name:"国际化语言键"`
	IsVirtual     null.Bool   `name:"是否虚菜单"`
}

// BeforeSave
func (m *Menu) BeforeCreate(scope *gorm.Scope) error {
	if m.LocaleKey.String == "" {
		prefix := "menu_"
		code := strings.ToLower(strings.TrimSpace(m.Code))
		if !strings.HasPrefix(code, prefix) {
			code = fmt.Sprintf("%s%s", prefix, code)
		}
		if err := scope.SetColumn("LocaleKey", null.StringFrom(code)); err != nil {
			return err
		}
	}
	if !m.Closeable.Valid {
		if err := scope.SetColumn("Closeable", null.NewBool(true, true)); err != nil {
			return err
		}
	}
	if _, err := m.setParentCodes(scope); err != nil {
		return err
	}
	return nil
}

// BeforeUpdate
func (m *Menu) BeforeUpdate(scope *gorm.Scope) error {
	if m.ID != 0 && m.ParentCode.Valid {
		var old Menu
		if err := scope.NewDB().Model(&Menu{}).Where(&Menu{ModelExOrg: ModelExOrg{ID: m.ID}}).First(&old).Error; err != nil {
			return err
		}
		newParentCodes, err := m.setParentCodes(scope)
		if err != nil {
			return err
		}
		if err := m.updateChildrenParentCodes(scope, old.Code, old.ParentCodes.String, newParentCodes); err != nil {
			return err
		}
	}
	return nil
}

// AfterSave
func (m *Menu) AfterSave(tx *gorm.DB) error {
	if err := m.updatePresetRolePrivileges(tx); err != nil {
		return err
	}
	return nil
}

// BeforeDelete
func (m *Menu) BeforeDelete(scope *gorm.Scope) error {
	tx := scope.NewDB()
	var old Menu
	if err := tx.Model(&Menu{}).Where(&Menu{ModelExOrg: ModelExOrg{ID: m.ID}}).First(&old).Error; err != nil {
		return err
	}
	f, _ := scope.FieldByName("ParentCodes")
	sql := fmt.Sprintf("%s.%s LIKE '%s'", scope.QuotedTableName(), scope.Quote(f.DBName), "%,"+old.Code+",%")
	scope.Search.Or(sql)
	return nil
}

// AfterDelete
func (m *Menu) AfterDelete(tx *gorm.DB) error {
	if err := m.updatePresetRolePrivileges(tx); err != nil {
		return err
	}
	return nil
}

func (m *Menu) updatePresetRolePrivileges(tx *gorm.DB) error {
	IgnoreAuth()
	defer IgnoreAuth(true)

	if err := tx.Model(&OperationPrivileges{}).Unscoped().Where(OperationPrivileges{RoleID: RootRoleID()}).Delete(OperationPrivileges{}).Error; err != nil {
		return err
	}
	var menus []Menu
	if err := tx.Model(&Menu{}).Find(&menus).Error; err != nil {
		return err
	}
	for _, menu := range menus {
		if err := tx.Model(&OperationPrivileges{}).Create(&OperationPrivileges{
			ModelExOrg: ModelExOrg{
				CreatedByID: RootUID(),
				UpdatedByID: RootUID(),
			},
			RoleID:   RootRoleID(),
			MenuCode: menu.Code,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}
func (m *Menu) setParentCodes(scope *gorm.Scope) (string, error) {
	var newParentCodes string
	if m.ParentCode.String != "" {
		var parent Menu
		IgnoreAuth()
		if err := scope.NewDB().Model(&Menu{}).Where(&Menu{Code: m.ParentCode.String}).First(&parent).Error; err != nil {
			return newParentCodes, err
		}
		IgnoreAuth(true)
		newParentCodes = fmt.Sprintf("%s%s,", parent.ParentCodes.String, parent.Code)
	} else {
		newParentCodes = "root,"
	}
	f, _ := scope.FieldByName("ParentCodes")
	if err := scope.SetColumn(f.DBName, newParentCodes); err != nil {
		return newParentCodes, err
	}
	return newParentCodes, nil
}

func (m *Menu) updateChildrenParentCodes(scope *gorm.Scope, oldCode, oldParentCodes, newParentCodes string) error {
	f, _ := scope.FieldByName("ParentCodes")
	quoteColumnName := scope.Quote(f.DBName)
	sql := fmt.Sprintf("UPDATE %s SET %s = REPLACE(%s, ?, ?) WHERE %s LIKE '%s'",
		scope.QuotedTableName(),
		quoteColumnName,
		quoteColumnName,
		quoteColumnName,
		"%,"+oldCode+",%",
	)
	if err := scope.NewDB().Exec(sql, oldParentCodes, newParentCodes).Error; err != nil {
		return err
	}
	return nil
}

// File
type File struct {
	Model `rest:"*" displayName:"文件"`
	ExtendField
	Class     string `name:"文件分类"`
	OwnerID   uint   `name:"归属数据ID"`
	OwnerType string `name:"归属数据类型"`

	UID  string `name:"文件唯一ID" json:"uid" gorm:"not null"`
	Type string `name:"文件Mine-Type" json:"type" gorm:"not null"`
	Size int64  `name:"文件大小" json:"size" gorm:"not null"`
	Name string `name:"文件名称" json:"name" gorm:"not null"`
	URL  string `name:"文件下载路径" json:"url" gorm:"not null"`
	Path string `name:"文件存储路径" json:"path" gorm:"not null"`
}

// BeforeCreate
func (f *File) BeforeCreate() {
	if f.UID == "" {
		f.UID = strings.ReplaceAll(uuid.NewV4().String(), "-", "")
	}
}

// Param
type Param struct {
	gorm.Model `rest:"*" displayName:"参数"`
	ExtendField
	Dr        int64     `name:"删除标记（0表示未删除，非0表示删除时间戳）" gorm:"DEFAULT:0;UNIQUE_INDEX:kuu_unique"`
	Code      string    `name:"参数编码" gorm:"not null;UNIQUE_INDEX:kuu_unique"`
	Name      string    `name:"参数名称" gorm:"not null"`
	Value     string    `name:"参数值" gorm:"type:text"`
	Type      string    `name:"参数类型"`
	IsBuiltIn null.Bool `name:"是否预置"`
}

// RepairDBTypes
func (l *Param) RepairDBTypes() {
	var (
		db        = DB()
		scope     = db.NewScope(l)
		tableName = scope.QuotedTableName()
		textType  = "MEDIUMTEXT"
	)
	if db.Dialect().GetName() == "mysql" {
		fields := []string{
			"value",
		}
		for _, item := range fields {
			sql := fmt.Sprintf("ALTER TABLE %s MODIFY %s %s NULL", tableName, scope.Quote(item), textType)
			ERROR(db.Exec(sql).Error)
		}
	}
}

func (l *Param) AfterSave(tx *gorm.DB) error {
	go func() {
		// 同时兼容单机和微服务模式
		time.Sleep(time.Second * 5)
		C().LoadFromParams(l.Code)
		DefaultCache.Publish(BuildKey("params", "update", "code"), l.Code)
	}()
	return nil
}
