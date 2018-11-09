package mongo

import (
	"github.com/kuuland/kuu"
)

var k *kuu.Kuu

func init() {
	kuu.On("OnNew", func(args ...interface{}) {
		k = args[0].(*kuu.Kuu)
		if c := k.Config["mongo"]; c != nil {
			uri := c.(string)
			Connect(uri)
		}
	})
}

// All 模块声明
func All() *kuu.Mod {
	return &kuu.Mod{}
}
