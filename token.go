package kuu

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin/binding"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/guregu/null.v3"
	"strings"
	"time"
)

type GenTokenDesc struct {
	UID          uint
	Username     string
	Exp          int64 `binding:"required"`
	Type         string
	Desc         string
	Payload      jwt.MapClaims
	IsAPIKey     bool
	ForcePayload bool // 是否强制使用Payload参数值作为jwt的payload
}

// GenToken
func GenToken(desc GenTokenDesc) (secretData *SignSecret, err error) {
	if err := binding.Validator.ValidateStruct(&desc); err != nil {
		return nil, err
	}
	if desc.IsAPIKey && desc.Desc == "" {
		return nil, errors.New("API Keys needs a description")
	}

	// 设置JWT令牌信息
	iat := time.Now().Unix()
	desc.Payload["iat"] = iat      // 签发时间：必须用全小写iat
	desc.Payload["exp"] = desc.Exp // 过期时间：必须用全小写exp
	desc.Payload["_k"] = strings.ReplaceAll(uuid.NewV4().String(), "-", "")
	// 兼容未传递SubDocID时自动查询
	var (
		subDocID uint
		user     = GetUserFromCache(desc.UID)
	)
	if v, err := user.GetSubDocID(desc.Type); err != nil {
		return nil, err
	} else {
		subDocID = v
	}
	// 生成新密钥
	secretData = &SignSecret{
		UID:      desc.UID,
		Username: desc.Username,
		Secret:   strings.ReplaceAll(uuid.NewV4().String(), "-", ""),
		Iat:      iat,
		Exp:      desc.Exp,
		Method:   SignMethodLogin,
		SubDocID: subDocID,
		Payload:  JSONStringify(desc.Payload),
		Desc:     desc.Desc,
		Type:     desc.Type,
		IsAPIKey: null.NewBool(desc.IsAPIKey, true),
	}
	// 签发令牌
	var tokenPayload jwt.MapClaims
	if desc.ForcePayload {
		tokenPayload = desc.Payload
	} else {
		tokenPayload = jwt.MapClaims{
			"_k":  desc.Payload["_k"],
			"UID": desc.Payload["UID"],
			"iat": desc.Payload["iat"],
			"exp": desc.Payload["exp"],
		}
	}
	if signed, err := EncodedToken(tokenPayload, secretData.Secret); err != nil {
		return secretData, err
	} else {
		secretData.Token = signed
	}
	desc.Payload[TokenKey] = secretData.Token
	if err = DB().Create(secretData).Error; err != nil {
		return
	}
	// 将secret存入缓存
	expDur := time.Unix(secretData.Exp, 0).Sub(time.Unix(secretData.Iat, 0))
	SetCacheString(secretData.Token, JSONStringify(secretData, false), expDur)
	// 保存登入历史
	saveHistory(secretData)
	return
}
