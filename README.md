# kuu

Pluggable Go Web Framework.

## Documentation

- [API Reference](https://godoc.org/github.com/kuuland/kuu)
- [Examples](https://godoc.org/github.com/kuuland/kuu#pkg-examples)

## Installation

```sh
go get -u github.com/kuuland/kuu
```

## Features

- 🎉 **插件式设计** -  灵活的插件机制
- ✨ **领域建模** - 面向数据模型设计
- 🚀 **配套增删改查API** - 数据模型自动注册CURD路由
- 🐠 **配套管理UI*** - 自带后台基础管理框架

![plan](https://raw.githubusercontent.com/kuuland/kuu/master/plan.png)

## Example

```go
package main

import (
	"fmt"

	"github.com/kuuland/kuu"
	"github.com/kuuland/kuu/plugins/mongo"
	"github.com/kuuland/kuu/plugins/task"
)

func main() {
	kuu.Import(mongo.All(), task.All())
	k := kuu.New(kuu.H{
		"mongo":   "mongodb://root:kuuland@127.0.0.1:27017/kuu?authSource=admin&maxPoolSize=50"
	})
	kuu.Info("Hello Kuu.")
	k.Run(":8080")
}

```
