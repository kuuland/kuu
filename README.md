# Kuu

[![GoDoc](https://godoc.org/github.com/kuuland/kuu?status.svg)](https://godoc.org/github.com/kuuland/kuu)

Modular Go Web Framework based on [GORM](https://github.com/jinzhu/gorm) and [Gin](https://github.com/gin-gonic/gin).

## Contents

- [Installation](#installation)
- [Quick start](#quick-start)
- [Features](#features)
    - [Global configuration](#global-configuration)
    - [RESTful APIs for struct](#restful-apis-for-struct)
    - [Modular project structure](#modular-project-structure)
    - [Global log API](#global-log-api)
    - [Standard response format](#standard-response-format)
    - [Common functions](#common-functions)
    - [Preset modules](#preset-modules)
- [FAQ](#faq)
- [License](#license)

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
	gorm.Model `rest:"*"`
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

### RESTful APIs for struct

Automatically mount RESTful APIs for struct:

```go
type User struct {
	gorm.Model `rest:"*"`
	Code string
	Name string
}

func main() {
	kuu.RESTful(r, &User{})
}
```

```text
[GIN-debug] POST   /api/user  --> github.com/kuuland/kuu.RESTful.func1 (4 handlers)
[GIN-debug] DELETE /api/user  --> github.com/kuuland/kuu.RESTful.func2 (4 handlers)
[GIN-debug] GET    /api/user  --> github.com/kuuland/kuu.RESTful.func3 (4 handlers)
[GIN-debug] PUT    /api/user  --> github.com/kuuland/kuu.RESTful.func4 (4 handlers)
```

On other fields:

```go
type User struct {
	gorm.Model
	Code string `rest:"*"`
	Name string
}
```

You can also change the default request method:

```go
type User struct {
	gorm.Model `rest:"C:POST;U:PUT;R:GET;D:DELETE"`
	Code string
	Name string
}

func main() {
	kuu.RESTful(r, &User{})
}
```

Or change route path:

```go
type User struct {
	gorm.Model `rest:"*" route:"profile"`
	Code string
	Name string
}

func main() {
	kuu.RESTful(r, &User{})
}
```

```text
[GIN-debug] POST   /api/profile  --> github.com/kuuland/kuu.RESTful.func1 (4 handlers)
[GIN-debug] DELETE /api/profile  --> github.com/kuuland/kuu.RESTful.func2 (4 handlers)
[GIN-debug] GET    /api/profile  --> github.com/kuuland/kuu.RESTful.func3 (4 handlers)
[GIN-debug] PUT    /api/profile  --> github.com/kuuland/kuu.RESTful.func4 (4 handlers)
```

Or unmount:

```go
type User struct {
	gorm.Model `rest:"C:-;U:PUT;R:GET;D:-"` // unmount all: `rest:"-"`
	Code string
	Name string
}

func main() {
	kuu.RESTful(r, &User{})
}
```

#### Create Record

```sh
curl -X POST \
  http://localhost:8080/api/user \
  -H 'Content-Type: application/json' \
  -d '{
    "user": "test",
    "pass": "123"
}'
```

#### Batch Create

```sh
curl -X POST \
  http://localhost:8080/api/user \
  -H 'Content-Type: application/json' \
  -d '[
    {
        "user": "test1",
        "pass": "123456"
    },
    {
        "user": "test2",
        "pass": "123456"
    },
    {
        "user": "test3",
        "pass": "123456"
    }
]'
```

#### Query

Request querystring parameters:

```sh
curl -X GET \
  'http://localhost:8080/api/user?cond={"user":"test"}&sort=id&project=pass'
```

|  Key  |  Desc  | Default | Example |
| ------ | ------ | ------ | ------ |
| range | data range, allow `ALL` and `PAGE` | `PAGE` | `range=ALL` |
| cond | query condition, JSON string | - | `cond={"user":"test"}` |
| sort | order fields | - | `sort=id,-user` |
| project | select fields | - | `project=user,pass` |
| page | current page(required in `PAGE` mode) | 1 | `page=2` |
| size | record size per page(required in `PAGE` mode) | 30 | `size=100` |

Query operators:

| Operator  |  Desc  | Example |
| ------ | ------ | ------ |
| `$regex` | LIKE | `cond={"user":{"$regex":"^test$"}}` |
| `$in` | IN | `cond={"id":{"$in":[1,2,5]}}` |
| `$nin` | NOT IN | `cond={"id":{"$nin":[1,2,5]}}` |
| `$eq` | Equal | `cond={"id":{"$eq":5}}` equivalent to `cond={"id":5}` |
| `$ne` | NOT Equal | `cond={"id":{"$ne":5}}` |
| `$exists` | IS NOT NULL | `cond={"pass":{"$exists":true}}` |
| `$gt` | Greater Than | `cond={"id":{"$gt":5}}` |
| `$gte` | Greater Than or Equal | `cond={"id":{"$gte":5}}` |
| `$lt` | Less Than | `cond={"id":{"$lt":20}}` |
| `$lte` | Less Than or Equal | `cond={"id":{"$lte":20}}`, `cond={"id":{"$gte":5,"$lte":20}}` |
| `$and` | AND | `cond={"user":"root","$and":[{"pass":"123"},{"pass":{"$regex":"^333"}}]}` |
| `$or` | OR | `cond={"user":"root","$or":[{"pass":"123"},{"pass":{"$regex":"^333"}}]}` |

Response JSON body:

```json
{
    "data": {
        "cond": {
            "user": "test"
        },
        "list": [
            {
                "ID": 3,
                "CreatedAt": "2019-05-10T09:19:40.437816Z",
                "UpdatedAt": "2019-05-12T07:04:13.583093Z",
                "DeletedAt": null,
                "User": "test",
                "Pass": "123456"
            },
            {
                "ID": 5,
                "CreatedAt": "2019-05-10T10:31:43.203526Z",
                "UpdatedAt": "2019-05-12T07:04:13.583093Z",
                "DeletedAt": null,
                "User": "test",
                "Pass": "111222333"
            }
        ],
        "page": 1,
        "range": "PAGE",
        "size": 30,
        "sort": "id",
        "totalpages": 1,
        "totalrecords": 2
    },
    "code": 0,
    "msg": ""
}
```

|  Key  |  Desc  | Default |
| ------ | ------ | ------ |
| list | data list | `[]` |
| range | data range, same as request | `PAGE` |
| cond | query condition, same as request | - |
| sort | order fields, same as request | - |
| project | select fields, same as request | - |
| totalrecords | total records | 0 |
| page | current page, exist in `PAGE` mode | 1 |
| size | record size per page, exist in `PAGE` mode | 30 |
| totalpages | total pages, exist in `PAGE` mode  | 0 |

#### Update Fields

```sh
curl -X PUT \
  http://localhost:8080/api/user \
  -H 'Content-Type: application/json' \
  -d '{
    "cond": {
        "id": 5
    },
    "doc": {
        "user": "new username"
    }
}'
```

#### Batch Updates

```sh
curl -X PUT \
  http://localhost:8080/api/user \
  -H 'Content-Type: application/json' \
  -d '{
    "cond": {
        "user": "test"
    },
    "doc": {
        "pass": "newpass"
    },
    "multi": true
}'
```

#### Delete Record

```sh
curl -X DELETE \
  http://localhost:8080/api/user \
  -H 'Content-Type: application/json' \
  -d '{
    "cond": {
        "id": 5
    }
}'
```

#### Batch Delete

```sh
curl -X DELETE \
  http://localhost:8080/api/user \
  -H 'Content-Type: application/json' \
  -d '{
    "cond": {
        "user": "test"
    },
    "multi": true
}'
```



### Modular project structure

Kuu will automatically mount routes, middleware and struct RESTful APIs after `kuu.Import`:

```go
type User struct {
	gorm.Model `rest`
	Username string
	Password string
}

type Profile struct {
	gorm.Model `rest`
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
	kuu.Import(r, All())                       // import custom module
	kuu.Import(r, accounts.All(), sys.All())   // import preset modules
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

### Common functions

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

### Preset modules

- [Accounts module](https://github.com/kuuland/accounts) - [JWT-based](https://jwt.io/) token issuance, login authentication, etc.
- [System module](https://github.com/kuuland/sys) - Menu, [admin](https://github.com/kuuland/ui), roles, organization, etc.
- [Internationalization(i18n)](https://github.com/kuuland/i18n) - Translation tools for multilingual applications.
- [Admin](https://github.com/kuuland/ui) - A React boilerplate.

## FAQ

### Why called Kuu?

> Kuu is a lovely cat, click to read [The story of Kuu and Shino](https://www.youtube.com/results?search_query=kuu+shino).

![kuu](https://raw.githubusercontent.com/kuuland/kuu/master/kuu.png)

## License

Kuu is available under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0.html).
