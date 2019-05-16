package kuu

//// OrgMiddleware
//func OrgMiddleware(c *gin.Context) {
//	// 从上下文缓存中读取认证信息
//	var sign *SignContext
//	if v, exists := c.Get(SignContextKey); exists {
//		sign = v.(*SignContext)
//	} else {
//		sign, _ = DecodedContext(c)
//	}
//
//	var o LoginOrg
//	DB().Model(&LoginOrg{UID: sign.UID, Token: sign.Token}).Preload("Org").First(&o)
//	if o.IsValid() {
//
//	}
//
//	if record.ID != "" && record.Org.ID != "" {
//		orgIDCacheKey, orgIDCacheVal := "LoginOrgID", record.Org.ID
//		c.Set(orgIDCacheKey, orgIDCacheVal)
//		kuu.SetGoroutineCache(orgIDCacheKey, orgIDCacheVal)
//	}
//	c.Next()
//	reg := regexp.MustCompile("/api/login")
//	if reg.MatchString(c.Request.RequestURI) {
//		err := orgAutoLogin(c)
//		if err != nil {
//			kuu.Error(err)
//			c.AbortWithStatusJSON(http.StatusOK, kuu.StdError(kuu.L(c, "auto_org_login_error", "Org login failed")))
//		}
//	}
//	kuu.ClearGoroutineCache()
//}
//
//func orgAutoLogin(c *gin.Context) error {
//	token := c.GetString("LoginToken")
//	uid := c.GetString("LoginUID")
//	if token == "" || uid == "" {
//		return nil
//	}
//	list := models.GetOrgList(uid)
//	if list != nil && len(list) == 1 && list[0] != nil {
//		item := list[0]
//		if v, ok := item["_id"].(string); ok && v != "" {
//			_, err := models.ExecOrgLogin(c, v, uid, token)
//			if err == nil {
//				c.SetCookie("Org", v, 0, "", "", false, false)
//			}
//			return err
//		}
//	}
//	return nil
//}
//
//
//// GetOrgList 查询指定用户ID的组织列表
//func GetOrgList(uid string) (data []kuu.H) {
//	if uid == kuu.Data["RootUID"] {
//		Org := kuu.Model("Org")
//		Org.List(kuu.H{
//			"Project": map[string]int{
//				"Code": 1,
//				"Name": 1,
//			},
//		}, &data)
//	} else {
//		// 先查授权规则
//		var rules []AuthRule
//		AuthRule := kuu.Model("AuthRule")
//		AuthRule.List(kuu.H{
//			"Cond": kuu.H{
//				"UID": uid,
//			},
//			"Project": map[string]int{
//				"OrgID": 1,
//			},
//		}, &rules)
//		existMap := make(map[string]bool, 0)
//		orgIDs := make([]bson.ObjectId, 0)
//		if rules != nil {
//			for _, rule := range rules {
//				if existMap[rule.OrgID] {
//					continue
//				}
//				existMap[rule.OrgID] = true
//				orgIDs = append(orgIDs, bson.ObjectIdHex(rule.OrgID))
//			}
//		}
//		// 再查ID对应的组织列表
//		Org := kuu.Model("Org")
//		Org.List(kuu.H{
//			"Cond": kuu.H{
//				"_id": kuu.H{
//					"$in": orgIDs,
//				},
//			},
//			"Project": map[string]int{
//				"Code": 1,
//				"Name": 1,
//			},
//		}, &data)
//	}
//	return
//}
//
//// ExecOrgLogin 执行组织登录
//func ExecOrgLogin(orgID string, uid string, token string) (*Org, error) {
//	if orgID == "" {
//		result := kuu.L(c, "body_parse_error")
//		return nil, errors.New(result)
//	}
//
//	OrgModel := kuu.Model("Org")
//	var org Org
//	err := OrgModel.ID(orgID, &org)
//	if org.ID == "" {
//		if err != nil {
//			kuu.Error(err)
//		}
//		result := kuu.L(c, "org_not_exist")
//		return nil, errors.New(result)
//	}
//
//	LoginOrgModel := kuu.Model("LoginOrg")
//	record := &LoginOrg{
//		UID:       uid,
//		Token:     token,
//		Org:       org,
//		CreatedBy: uid,
//		UpdatedBy: uid,
//	}
//	if _, err := LoginOrgModel.Create(record); err != nil {
//		kuu.Error(err)
//		result := kuu.L(c, "org_login_error")
//		return nil, errors.New(result)
//	}
//	return kuu.H{
//		"_id":  org.ID,
//		"Code": org.Code,
//		"Name": org.Name,
//	}, nil
//}
//
