# kuu

[![GoDoc](https://godoc.org/github.com/kuuland/kuu?status.svg)](https://godoc.org/github.com/kuuland/kuu)

Modular Go Web Framework based on [GORM](https://github.com/jinzhu/gorm) and [Gin](https://github.com/gin-gonic/gin).

## Contents

- [Installation](#installation)
- [Quick start](#quick-start)
- [Features](#features)
    - [Global configuration](#global-configuration)
    - [RESTful API for structs](#restful-api-for-structs)
    - [Modular specification](#modular-specification)
    - [Global log API](#global-log-api)
    - [Standard response format](#standard-response-format)
    - [Common Functions](#common-functions)
- [API Reference](https://godoc.org/github.com/kuuland/kuu)

## Installation

```sh
go get -u github.com/kuuland/kuu
```

## Quick start

```sh
# assume the following codes in kuu.json file
$ cat kuu.json
```

```json
{
  "prefix": "/api",
  "db": {
    "dialect": "postgres",
    "args": "host=127.0.0.1 port=5432 user=root dbname=kuu password=hello sslmode=disable"
  }
}
```

```sh
# assume the following codes in main.go file
$ cat main.go
```

```go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
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
	kuu.RESTful(r, &user{})
	r.GET("/ping", func(c *gin.Context) {
		kuu.INFO("Hello Kuu")
		var users []user
		kuu.DB().Find(&users)
		kuu.STD(c, users)
	})
	r.Run(":8080")
}
```

```sh
# run main.go and visit 0.0.0.0:8080/ping on browser
$ go run example.go
```

## Features

### Global configuration

```sh
# assume the following codes in kuu.json file
$ cat kuu.json
```

```json
{
  "string": "/api",
  "boolean": false,
  "number": 320,
  "digit": 45.22,
  "db": {
    "dialect": "postgres",
    "args": "host=127.0.0.1 port=5432 user=root dbname=kuu password=hello sslmode=disable"
  }
}
```

```go
func main() {
	kuu.C().Get("prefix")         // output "/api"
	kuu.C().GetString("string")   // output "/api"
	kuu.C().GetBool("boolean")    // output false
	kuu.C().GetInt("number")      // output 320
	kuu.C().GetFloat64("digit")   // output 45.22
}
```

### RESTful API for structs

Automatically mount RESTful API for structs:

```go
type user struct {
	kuu string `rest`
	gorm.Model
	User string
	Pass string
}

func main() {
	kuu.RESTful(r, &user{})
}
```

```text
[GIN-debug] POST   /api/user  --> github.com/kuuland/kuu.RESTful.func1 (4 handlers)
[GIN-debug] DELETE /api/user  --> github.com/kuuland/kuu.RESTful.func2 (4 handlers)
[GIN-debug] GET    /api/user  --> github.com/kuuland/kuu.RESTful.func3 (4 handlers)
[GIN-debug] PUT    /api/user  --> github.com/kuuland/kuu.RESTful.func4 (4 handlers)
```

You can also change the request method:

```go
type user struct {
	kuu string `rest:"C:POST;U:PUT;R:GET;D:DELETE"`
	gorm.Model
	User string
	Pass string
}

func main() {
	kuu.RESTful(r, &user{})
}
```

Or unmount:

```go
type user struct {
	kuu string `rest:"C:-;U:PUT;R:GET;D:-"` // unmount all: `rest:"-"`
	gorm.Model
	User string
	Pass string
}

func main() {
	kuu.RESTful(r, &user{})
}
```

### Modular specification

```go
type User struct {
	kuu string `rest`
	gorm.Model
	Username string
	Password string
}

type Profile struct {
	kuu string `rest`
	gorm.Model
	Nickname string
	Age int
}

func All() *kuu.Mod {
	return &kuu.Mod{
		Models: []interface{}{
			&User{},
			&Profile{},
		},
		Middleware: kuu.Middleware{
			func (c *gin.Context) {
                // Auth middleware
            },
		},
		Routes: kuu.Routes{
			kuu.RouteInfo{
                Method: "POST",
                Path:   "/login",
                Handler: func(c *gin.Context) {
                    // POST /login
                }
            },
			kuu.RouteInfo{
                Method: "POST",
                Path:   "/logout",
                Handler: func(c *gin.Context) {
                    // POST /logout
                }
            },
		},
	}
}

func main() {
	defer kuu.Release()
	r := gin.Default()
	kuu.Import(r, All()) // import custom module
	kuu.Import(r, accounts.All(), sys.All()) // import preset modules
}
```

### Global log API

```go
func main() {
	kuu.PRINT("Hello Kuu")  // PRINT[0000] Hello Kuu
	kuu.DEBUG("Hello Kuu")  // DEBUG[0000] Hello Kuu
	kuu.WARN("Hello Kuu")   // WARN[0000] Hello Kuu
	kuu.INFO("Hello Kuu")   // INFO[0000] Hello Kuu
	kuu.FATAL("Hello Kuu")  // FATAL[0000] Hello Kuu
	kuu.PANIC("Hello Kuu")  // PANIC[0000] Hello Kuu
}
```

Or with params:

```go
func main() {
	kuu.INFO("Hello %s", "Kuu")  // INFO[0000] Hello Kuu
}
```

### Standard response format

```go
func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		// 'kuu.STD' Can only be called once
        kuu.STD(c, "hello")                      // response: {"data":"hello","code":0,"msg":""}
        kuu.STD(c, "hello", "Success")           // response: {"data":"hello","code":0,"msg":"Success"}
        kuu.STD(c, 200)                          // response: {"data":200,"code":0,"msg":""}
        kuu.STDErr(c, "New record failed")       // response: {"data":null,"code":-1,"msg":"New record failed"}
        kuu.STDErr(c, "New record failed", 555)  // response: {"data":null,"code":555,"msg":"New record failed"}
    })
}
```

### Common Functions

```go
func main() {
	r := gin.Default()
	// Cross-domain support
	r.Use(kuu.CORSMiddleware())
	// Parse JSON from string
	var params map[string]string
	kuu.Parse(`{"user":"kuu","pass":"123"}`, &params)
	// Formatted as JSON
	kuu.Stringify(&params, true)
}
```

## FAQ

### Why is it called Kuu?

> Kuu is the name of a cat, click to read [The story of Kuu and Shino](http://www.sohu.com/a/225954042_509045).

![kuu](https://raw.githubusercontent.com/kuuland/kuu/master/kuu.png)

## License

Kuu is available under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0.html).
