package kuu

import (
	"errors"
	"fmt"
	"reflect"
	"time"
)

var (
	ErrTokenNotFound       = errors.New("token not found")
	ErrSecretNotFound      = errors.New("secret not found")
	ErrInvalidToken        = errors.New("invalid token")
	ErrAffectedSaveToken   = errors.New("未新增或修改任何记录，请检查更新条件或数据权限")
	ErrAffectedDeleteToken = errors.New("未删除任何记录，请检查更新条件或数据权限")
)

// CatchError error which one from outside of recovery pluigns, this rec just for kuu
// you can CatchError if your error code does not affect the next plug-in
// sometime you should handler all error in plugin
func CatchError(funk interface{}) (err error) {
	defer func() {
		var ok bool
		if ret := recover(); ret != nil {
			err, ok = ret.(error)
			if !ok {
				err = fmt.Errorf("%v", ret)
			}
			ERROR("%s kuu panic recovered:\n%s\n%s", timeFormat(time.Now()), err, stack(3))
		}
	}()
	assert1(isFunc(funk), fmt.Errorf("funk %v in CatchError should be func type", reflect.TypeOf(funk)))
	reflect.ValueOf(funk).Call([]reflect.Value{})
	return
}
