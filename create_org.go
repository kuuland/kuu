package kuu

import (
	"fmt"
	"github.com/imdario/mergo"
	"github.com/jinzhu/gorm"
	"gopkg.in/guregu/null.v3"
	"strings"
)

type CreateOrgArgs struct {
	Tx               *gorm.DB
	OrgCode          string
	OrgName          string
	ParentOrgID      uint
	ParentOrgCode    string
	AdminUID         uint
	AdminUsername    string
	AdminPassword    string
	GeneratePassword bool

	ExtraAdminUserInfo   User
	ExtraAdminRoleInfo   Role
	ExtraAssignRoleCodes []string
}

type CreateOrgReply struct {
	OrgID                      uint
	Org                        *Org
	AdminUID                   uint
	AdminUser                  *User
	GeneratedPlaintextPassword string
}

func CreateOrg(args *CreateOrgArgs) (reply *CreateOrgReply, err error) {

	tx := args.Tx
	if tx == nil {
		tx = DB()
		args.Tx = tx
	}
	reply = new(CreateOrgReply)
	// 1.创建组织
	var parentOrg Org
	if args.ParentOrgID != 0 {
		if err := tx.Model(&Org{}).Where(&Org{ID: args.ParentOrgID}).First(&parentOrg).Error; err != nil {
			return nil, err
		}
	} else if args.ParentOrgCode != "" {
		if err := tx.Model(&Org{}).Where(&Org{Code: args.ParentOrgCode}).First(&parentOrg).Error; err != nil {
			return nil, err
		}
	}
	org := Org{
		CreatedByID: RootUID(),
		UpdatedByID: RootUID(),
		Code:        args.OrgCode,
		Name:        args.OrgName,
		Pid:         parentOrg.ID,
		IsBuiltIn:   null.BoolFrom(true),
	}
	if err := tx.Model(&Org{}).Create(&org).Error; err != nil {
		return nil, err
	}
	reply.Org = &org
	reply.OrgID = org.ID
	if args.AdminUsername != "" || args.AdminUID > 0 {
		var adminUser User
		if args.AdminUID > 0 {
			if err := tx.Model(&User{}).Where(&User{ID: args.AdminUID}).First(&adminUser).Error; err != nil && err != gorm.ErrRecordNotFound {
				return nil, err
			}
		} else {
			if err := tx.Model(&User{}).Where(&User{Username: args.AdminUsername}).First(&adminUser).Error; err != nil && err != gorm.ErrRecordNotFound {
				return nil, err
			}
		}
		// 2.创建管理用户
		if adminUser.ID == 0 {
			var password string
			if args.GeneratePassword {
				reply.GeneratedPlaintextPassword = GenPassword()
				password = MD5(reply.GeneratedPlaintextPassword)
			} else {
				password = args.AdminPassword
			}
			adminUser = User{
				OrgID:       reply.OrgID,
				CreatedByID: RootUID(),
				UpdatedByID: RootUID(),
				Username:    args.AdminUsername,
				Password:    password,
				IsBuiltIn:   null.BoolFrom(true),
			}
			if err := mergo.Merge(&adminUser, args.ExtraAdminUserInfo); err != nil {
				return nil, err
			}
			if err := tx.Model(&User{}).Create(&adminUser).Error; err != nil {
				return nil, err
			}
		}
		reply.AdminUID = adminUser.ID
		reply.AdminUser = &adminUser
		// 3.创建管理用户角色
		adminRole := Role{
			CreatedByID: adminUser.ID,
			UpdatedByID: adminUser.ID,
			OrgID:       reply.OrgID,
			Code:        fmt.Sprintf("org_admin:%d", org.ID),
			Name:        strings.TrimSpace(fmt.Sprintf("%s (Admin)", org.Name)),
			IsBuiltIn:   null.BoolFrom(true),
		}
		if err := mergo.Merge(&adminRole, args.ExtraAdminRoleInfo); err != nil {
			return nil, err
		}
		if err := tx.Model(&Role{}).Create(&adminRole).Error; err != nil {
			return nil, err
		}
		// 4.创建数据权限记录
		if err := tx.Model(&DataPrivileges{}).Create(&DataPrivileges{
			Model: Model{
				CreatedByID: reply.AdminUID,
				UpdatedByID: reply.AdminUID,
				OrgID:       reply.OrgID,
			},
			RoleID:        adminRole.ID,
			TargetOrgID:   reply.OrgID,
			ReadableRange: DataScopeCurrentFollowing,
			WritableRange: DataScopeCurrentFollowing,
		}).Error; err != nil {
			return nil, err
		}
		// 5.创建角色授权记录
		if err := tx.Model(&RoleAssign{}).Create(&RoleAssign{
			ModelExOrg: ModelExOrg{
				CreatedByID: reply.AdminUID,
				UpdatedByID: reply.AdminUID,
			},
			RoleID: adminRole.ID,
			UserID: reply.AdminUID,
		}).Error; err != nil {
			return nil, err
		}
		if len(args.ExtraAssignRoleCodes) > 0 {
			var extraRoles []Role
			if err := tx.Model(&Role{}).Where(fmt.Sprintf("%s IN (?)", tx.Dialect().Quote("code")), args.ExtraAssignRoleCodes).Find(&extraRoles).Error; err != nil {
				return nil, err
			}
			for _, role := range extraRoles {
				if err := tx.Model(&RoleAssign{}).Create(&RoleAssign{
					ModelExOrg: ModelExOrg{
						CreatedByID: reply.AdminUID,
						UpdatedByID: reply.AdminUID,
					},
					RoleID: role.ID,
					UserID: adminUser.ID,
				}).Error; err != nil {
					return nil, err
				}
			}
		}
	}

	return reply, nil
}
