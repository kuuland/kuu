package plugins

import (
	"github.com/kuuland/kuu/plugins/accounts"
	"github.com/kuuland/kuu/plugins/mongo"
	"github.com/kuuland/kuu/plugins/task"
)

// 汇总导出
var (
	Mongo    = mongo.P
	Accounts = accounts.P
	Task     = task.P
)
