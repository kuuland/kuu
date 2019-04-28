package sys

import (
	"github.com/globalsign/mgo/bson"
	"github.com/kuuland/kuu"
	"github.com/kuuland/kuu/mods/sys/models"
	"github.com/kuuland/kuu/mods/sys/utils"
)

const initCode = "sys:init"

func preflight() bool {
	Param := kuu.Model("Param")
	ret := &models.Param{}
	Param.One(kuu.H{"Cond": kuu.H{"Code": initCode, "IsBuiltIn": true, "IsDeleted": kuu.H{"$ne": true}}}, ret)
	if ret != nil && ret.ID != "" {
		return true
	}
	return false
}

func getRootUser() models.User {
	User := kuu.Model("User")
	var rootUser models.User
	User.One(kuu.H{"Username": "root"}, &rootUser)
	return rootUser
}

func createRootUser() models.User {
	User := kuu.Model("User")
	rootUser := &models.User{
		Username:  "root",
		Name:      "预置用户",
		Password:  utils.MD5("kuu"),
		IsBuiltIn: true,
	}
	if _, err := User.Create(rootUser); err != nil {
		kuu.Error("初始化预置用户失败：%s", err.Error())
	}
	return getRootUser()
}

func createPresetDicts() {
	Dict := kuu.Model("Dict")
	items := []models.Dict{
		{
			Code:      "sys_menu_type",
			Name:      "菜单类型",
			IsBuiltIn: true,
			Values: []models.DictValue{
				{
					Label: "菜单",
					Value: "menu",
					Sort:  100,
				},
				{
					Label: "权限",
					Value: "permission",
					Sort:  200,
				},
			},
		},
		{
			Code:      "sys_data_range",
			Name:      "数据范围",
			IsBuiltIn: true,
			Values: []models.DictValue{
				{
					Label: "个人范围",
					Value: "personal",
					Sort:  100,
				},
				{
					Label: "当前组织范围",
					Value: "current",
					Sort:  200,
				},
				{
					Label: "当前及以下组织范围",
					Value: "current_following",
					Sort:  300,
				},
			},
		},
	}
	if _, err := Dict.Create(items); err != nil {
		kuu.Error("初始化预置菜单失败：%s", err.Error())
	}
}

func createPresetMenus() {
	Menu := kuu.Model("Menu")
	rootMenuID := bson.NewObjectId().Hex()
	sysMenuID := bson.NewObjectId().Hex()
	orgMenuID := bson.NewObjectId().Hex()
	permissionMenuID := bson.NewObjectId().Hex()
	settingMenuID := bson.NewObjectId().Hex()
	items := []models.Menu{
		{
			ID:        rootMenuID,
			Name:      "主导航菜单",
			Sort:      100,
			Type:      "menu",
			IsBuiltIn: true,
		},
		{
			ID:        sysMenuID,
			Pid:       rootMenuID,
			Name:      "系统管理",
			Icon:      "setting",
			Sort:      100,
			Type:      "menu",
			IsBuiltIn: true,
		},
		{
			ID:        orgMenuID,
			Pid:       sysMenuID,
			Name:      "组织管理",
			Icon:      "appstore",
			Sort:      100,
			Type:      "menu",
			IsBuiltIn: true,
		},
		{
			Pid:       orgMenuID,
			Name:      "用户管理",
			Icon:      "user",
			URI:       "/sys/user",
			Sort:      100,
			Type:      "menu",
			IsBuiltIn: true,
			Closeable: true,
		},
		{
			Pid:       orgMenuID,
			Name:      "组织机构",
			Icon:      "cluster",
			URI:       "/sys/org",
			Sort:      200,
			Type:      "menu",
			IsBuiltIn: true,
			Closeable: true,
		},
		{
			ID:        permissionMenuID,
			Pid:       sysMenuID,
			Name:      "权限管理",
			Icon:      "dropbox",
			Sort:      200,
			Type:      "menu",
			IsBuiltIn: true,
		},
		{
			Pid:       permissionMenuID,
			Name:      "角色管理",
			Icon:      "team",
			URI:       "/sys/role",
			Sort:      100,
			Type:      "menu",
			IsBuiltIn: true,
			Closeable: true,
		},
		{
			ID:        settingMenuID,
			Pid:       sysMenuID,
			Name:      "系统设置",
			Icon:      "tool",
			Sort:      300,
			Type:      "menu",
			IsBuiltIn: true,
		},
		{
			Pid:       settingMenuID,
			Name:      "菜单管理",
			Icon:      "bars",
			URI:       "/sys/menu",
			Sort:      100,
			Type:      "menu",
			IsBuiltIn: true,
			Closeable: true,
		},
		{
			Pid:       settingMenuID,
			Name:      "参数管理",
			Icon:      "profile",
			URI:       "/sys/param",
			Sort:      200,
			Type:      "menu",
			IsBuiltIn: true,
			Closeable: true,
		},
		{
			Pid:       settingMenuID,
			Name:      "字典管理",
			Icon:      "build",
			URI:       "/sys/dict",
			Sort:      300,
			Type:      "menu",
			IsBuiltIn: true,
			Closeable: true,
		},
		{
			Pid:       settingMenuID,
			Name:      "审计日志",
			Icon:      "book",
			URI:       "/sys/audit",
			Sort:      400,
			Type:      "menu",
			IsBuiltIn: true,
			Closeable: true,
		},
		{
			Pid:       settingMenuID,
			Name:      "文件",
			Icon:      "file",
			URI:       "/sys/file",
			Sort:      500,
			Type:      "menu",
			IsBuiltIn: true,
			Closeable: true,
		},
		{
			Pid:       settingMenuID,
			Name:      "国际化",
			Icon:      "global",
			URI:       "/sys/i18n",
			Sort:      600,
			Type:      "menu",
			IsBuiltIn: true,
			Closeable: true,
		},
		{
			Pid:       settingMenuID,
			Name:      "消息",
			Icon:      "message",
			URI:       "/sys/message",
			Sort:      700,
			Type:      "menu",
			IsBuiltIn: true,
			Closeable: true,
		},
	}
	if _, err := Menu.Create(items); err != nil {
		kuu.Error("初始化预置菜单失败：%s", err.Error())
	}
}

func saveInitParam() {
	Param := kuu.Model("Param")
	rootUser := getRootUser()
	ret := &models.Param{
		Code:      initCode,
		IsBuiltIn: true,
		Name:      "系统初始化标记",
		Value:     "ok",
		CreatedBy: rootUser,
		UpdatedBy: rootUser,
	}
	if _, err := Param.Create(ret); err != nil {
		kuu.Error("保存初始化标记失败：%s", err.Error())
	}
}

func sysInit() {
	// 1.默认菜单/系统模块/(菜单管理、参数管理、字典管理、角色管理、用户管理、国际化管理)
	if preflight() {
		return
	}
	// 初始化预置用户
	createRootUser()
	// 初始化字典
	createPresetDicts()
	// 初始化菜单
	createPresetMenus()
	// 保存初始化标记
	saveInitParam()
}
