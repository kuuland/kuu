package kuu

import (
	"errors"
)

var (
	ErrTokenNotFound       = errors.New("token not found")
	ErrSecretNotFound      = errors.New("secret not found")
	ErrInvalidToken        = errors.New("invalid token")
	ErrAffectedSaveToken   = errors.New("未新增或修改任何记录，请检查更新条件或数据权限")
	ErrAffectedDeleteToken = errors.New("未删除任何记录，请检查更新条件或数据权限")
)
