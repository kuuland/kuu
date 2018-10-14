package plugins

import (
	"github.com/kuuland/kuu"
	"github.com/kuuland/kuu/plugins/accounts"
	"github.com/kuuland/kuu/plugins/mongo"
)

// Mongo MongoDB插件别名
func Mongo() *kuu.Plugin {
	return mongo.Install()
}

// Accounts 账户插件别名
func Accounts() *kuu.Plugin {
	return accounts.Install()
}
