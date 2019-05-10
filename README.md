# kuu

[![GoDoc](https://godoc.org/github.com/kuuland/kuu?status.svg)](https://godoc.org/github.com/kuuland/kuu)

Modular Go Web Framework.


## Features

- 🎉 **全局共享应用配置**
- ✨ **定义模块开发规范**
- 🚀 **提供全局日志API**
- 🐠 **提供常用工具函数**
- 👻 **提供常用模块**

## Documentation

- [API Reference](https://godoc.org/github.com/kuuland/kuu)

## Installation

```sh
go get -u github.com/kuuland/kuu
```

## Usage

```go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/kuuland/kuu"
)

type user struct {
	kuu string `rest`
	gorm.Model
	User string
	Pass string
}

func main() {
	defer kuu.Release()
	
	r := gin.Default()
	r.Use(kuu.CORSMiddleware())
	kuu.MountRESTful(r, &user{})
	r.GET("/ping", func(c *gin.Context) {
		kuu.INFO("Hello Kuu")
		var users = []user{}
		kuu.DB().Find(&users)
		kuu.STD(c, users)
	})
	r.Run(":8080")
}

```

kuu.json:

```json
{
  "prefix": "/api",
  "db": {
    "dialect": "postgres",
    "args": "host=127.0.0.1 port=5432 user=root dbname=kuu password=hello sslmode=disable"
  }
}
```

## FAQ

### Why is it called Kuu?

> Kuu is the name of a cat, click to read [The story of Kuu and Shino](http://www.sohu.com/a/225954042_509045).

![kuu](https://raw.githubusercontent.com/kuuland/kuu/master/kuu.png)

## License

Kuu is available under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0.html).
