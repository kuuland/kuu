package kuu

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin/binding"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/guregu/null.v3"
	"time"
)

type GenTokenDesc struct {
	UID      uint   `binding:"required"`
	Exp      int64  `binding:"required"`
	Type     string `binding:"required"`
	Desc     string
	Payload  jwt.MapClaims
	IsAPIKey bool
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
		Secret:   uuid.NewV4().String(),
		Iat:      iat,
		Exp:      desc.Exp,
		Method:   SignMethodLogin,
		SubDocID: subDocID,
		Desc:     desc.Desc,
		Type:     desc.Type,
		IsAPIKey: null.NewBool(desc.IsAPIKey, true),
	}
	// 签发令牌
	if signed, err := EncodedToken(desc.Payload, secretData.Secret); err != nil {
		return secretData, err
	} else {
		secretData.Token = signed
	}
	desc.Payload[TokenKey] = secretData.Token
	if err = DB().Create(secretData).Error; err != nil {
		return
	}
	// 保存登入历史
	saveHistory(secretData)
	return
}
