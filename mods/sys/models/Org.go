package models

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
	"github.com/kuuland/kuu"
)

// Org 系统组织
type Org struct {
	ID           string `json:"_id" displayName:"系统组织"`
	Code         string `name:"组织编码"`
	Name         string `name:"组织名称"`
	Pid          string `name:"父组织ID"`
	Sort         int    `name:"排序号"`
	FullPathPid  string `name:"完整层级路径"`
	FullPathName string `name:"完整层级路径"`
	// 标准字段
	CreatedBy User   `name:"创建人" join:"User<Username,Name>"`
	CreatedAt int64  `name:"创建时间"`
	UpdatedBy User   `name:"修改人" join:"User<Username,Name>"`
	UpdatedAt int64  `name:"修改时间"`
	IsDeleted bool   `name:"是否已删除"`
	Remark    string `name:"备注"`
}

// GetOrgList 查询指定用户ID的组织列表
func GetOrgList(uid string) (data []kuu.H) {
	if uid == kuu.Data["RootUID"] {
		Org := kuu.Model("Org")
		Org.List(kuu.H{
			"Project": map[string]int{
				"Code": 1,
				"Name": 1,
			},
		}, &data)
	} else {
		// 先查授权规则
		var rules []AuthRule
		AuthRule := kuu.Model("AuthRule")
		AuthRule.List(kuu.H{
			"Cond": kuu.H{
				"UID": uid,
			},
			"Project": map[string]int{
				"OrgID": 1,
			},
		}, &rules)
		existMap := make(map[string]bool, 0)
		orgIDs := make([]bson.ObjectId, 0)
		if rules != nil {
			for _, rule := range rules {
				if existMap[rule.OrgID] {
					continue
				}
				existMap[rule.OrgID] = true
				orgIDs = append(orgIDs, bson.ObjectIdHex(rule.OrgID))
			}
		}
		// 再查ID对应的组织列表
		Org := kuu.Model("Org")
		Org.List(kuu.H{
			"Cond": kuu.H{
				"_id": kuu.H{
					"$in": orgIDs,
				},
			},
			"Project": map[string]int{
				"Code": 1,
				"Name": 1,
			},
		}, &data)
	}
	return
}

// ExecOrgLogin 执行组织登录
func ExecOrgLogin(c *gin.Context, orgID string, uid string, token string) (kuu.H, error) {
	if orgID == "" {
		result := kuu.L(c, "body_parse_error")
		return nil, errors.New(result)
	}

	OrgModel := kuu.Model("Org")
	var org Org
	err := OrgModel.ID(orgID, &org)
	if org.ID == "" {
		if err != nil {
			kuu.Error(err)
		}
		result := kuu.L(c, "org_not_exist")
		return nil, errors.New(result)
	}

	LoginOrgModel := kuu.Model("LoginOrg")
	record := &LoginOrg{
		UID:       uid,
		Token:     token,
		Org:       org,
		CreatedBy: uid,
		UpdatedBy: uid,
	}
	if _, err := LoginOrgModel.Create(record); err != nil {
		kuu.Error(err)
		result := kuu.L(c, "org_login_error")
		return nil, errors.New(result)
	}
	return kuu.H{
		"_id":  org.ID,
		"Code": org.Code,
		"Name": org.Name,
	}, nil
}
