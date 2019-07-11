<img src="/docs/logo.png" alt="Modular Go Web Framework"/>

[![GoDoc](https://godoc.org/github.com/kuuland/kuu?status.svg)](https://godoc.org/github.com/kuuland/kuu)

Modular Go Web Framework based on [GORM](https://github.com/jinzhu/gorm) and [Gin](https://github.com/gin-gonic/gin).

## Contents

- [Installation](#installation)
- [Quick start](#quick-start)
- [Features](#features)
    - [Global configuration](#global-configuration)
    - [Data source management](#data-source-management)
    - [RESTful APIs for struct](#restful-apis-for-struct)
    - [Password field filter](#password-field-filter)
    - [Associations](#associations)
    - [Global default callbacks](#global-default-callbacks)
    - [Struct validation](#struct-validation)
    - [Modular project structure](#modular-project-structure)
    - [Global log API](#global-log-api)
    - [Standard response format](#standard-response-format)
    - [Common functions](#common-functions)
    - [Get login context](#get-login-context)
    - [Goroutine local storage](#goroutine-local-storage)
    - [Whitelist](#whitelist)
    - [Preset modules](#preset-modules)
    - [Security framework](#security-framework)
- [FAQ](#faq)
    - [Why called Kuu?](#why-called-kuu)
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
  },
  "redis": {
    "addr": "127.0.0.1:6379"
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
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/kuuland/kuu"
)

func main() {
	r := kuu.Default()
	r.Import(kuu.Acc(), kuu.Sys())
	r.Run()
}
```

```sh
# run main.go and visit 0.0.0.0:8080 on browser
$ go run main.go
```

## Features

### Global configuration

```sh
# assume the following codes in kuu.json file
$ cat kuu.json
```

```json
{
  "prefix": "/api",
  "cors": true,
  "gzip": true,
  "gorm:migrate": false,
  "db": {
    "dialect": "postgres",
    "args": "host=127.0.0.1 port=5432 user=root dbname=kuu password=hello sslmode=disable"
  },
  "redis": {
    "addr": "127.0.0.1:6379"
  },
  "statics": {
    "/assets": "assets/",
    "/drone_yml": ".drone.yml"
  }
}
```

```go
func main() {
	kuu.C().Get("prefix")              // output "/api"
	kuu.C().GetBool("cors")            // output true
	kuu.C().GetBool("gorm:migrate")    // output true
}
```

List of preset config:

- `prefix` - Global routes prefix for `kuu.Mod`'s Routes.
- `gorm:migrate` - Enable GORM's auto migration for Mod's Models.
- `audit:callbacks` - Register audit callbacks, default is `true`.
- `audit:format` - Formatted output audit info.
- `db` - DB configs.
- `redis` - Redis configs.
- `cors` - Attaches the official [CORS](https://github.com/gin-contrib/cors) gin's middleware.
- `gzip` - Attaches the gin middleware to enable [GZIP](https://github.com/gin-contrib/gzip) support.
- `statics` - Static serves files from the given file system root or serve a single file.
- `whitelist:prefix` - Let whitelist also matches paths with global prefix, default is `true`.
- `ignoreDefaultRootRoute` - Do not mount the default root route, default is `false`.

> Notes: Static paths are automatically added to the [whitelist](#whitelist).

### Data Source Management

Single data source:

```json
{
  "db": {
    "dialect": "postgres",
    "args": "host=127.0.0.1 port=5432 user=root dbname=db1 password=hello sslmode=disable"
  }
}
```

```go
r.GET("/ping", func(c *kuu.Context) {
    var users []user
    kuu.DB().Find(&users)
    c.STD(&users)
})
```

Multiple data source:

```json
{
  "db": [
    {
      "name": "ds1",
      "dialect": "postgres",
      "args": "host=127.0.0.1 port=5432 user=root dbname=db1 password=hello sslmode=disable"
    },
    {
      "name": "ds2",
      "dialect": "postgres",
      "args": "host=127.0.0.1 port=5432 user=root dbname=db1 password=hello sslmode=disable"
    }
  ]
}
```

```go
r.GET("/ping", func(c *kuu.Context) {
    var users []user
    kuu.DB("ds1").Find(&users)
    c.STD(&users)
})
```


### RESTful APIs for struct

Automatically mount RESTful APIs for struct:

```go
type User struct {
	kuu.Model `rest:"*"`
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
	kuu.Model
	Code string `rest:"*"`
	Name string
}
```

You can also change the default request method:

```go
type User struct {
	kuu.Model `rest:"C:POST;U:PUT;R:GET;D:DELETE"`
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
	kuu.Model `rest:"*" route:"profile"`
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
	kuu.Model `rest:"C:-;U:PUT;R:GET;D:-"` // unmount all: `rest:"-"`
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
| preload | preload fields | - | `preload=CreditCards,UserAddresses` |
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
| preload | preload fields, same as request | - |
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

#### UnSoft Delete

```sh
curl -X DELETE \
  http://localhost:8080/api/user \
  -H 'Content-Type: application/json' \
  -d '{
    "cond": {
        "user": "test"
    },
    "unsoft": true
}'
```

### Password field filter

```go
type User struct {
	Model   `rest:"*" displayName:"用户"`
	Username    string  `name:"账号"`
	Password    string  `name:"密码" kuu:"password"`
}

users := []User{
    {Username: "root", Password: "xxx"},
    {Username: "admin", Password: "xxx"},
} 
users = kuu.OmitPassword("User", users) // => []User{ { Username: "root" }, { Username: "admin" } } 
```

### Associations

1. set `"auto":true` to enable auto save associations
1. if association has a primary key, Kuu will call Update to save it, otherwise it will be created
1. set `"preload=field1,field2"` to preload associations

#### Create associations

```sh
curl -X PUT \
  http://localhost:8080/api/user \
  -H 'Content-Type: application/json' \
  -d '{
    "cond": {
        "ID": 50
    },
    "doc": {
        "Emails": [
            {
                "Email": "test1@example.com"
            },
            {
                "Email": "test2@example.com"
            }
        ]
    },
    "auto": true
}'
```

#### Update associations

`ID` is required:

```sh
curl -X PUT \
  http://localhost:8080/api/user \
  -d '{
    "cond": {
        "ID": 50
    },
    "doc": {
        "Emails": [
            {
                "ID": 101,
                "Email": "test111@example.com"
            },
            {
                "ID": 159,
                "Email": "test222@example.com"
            }
        ]
    },
    "auto": true
}'
```

#### Delete associations

`ID` and `DeletedAt` are required:

```sh
curl -X PUT \
  http://localhost:8080/api/user \
  -H 'Content-Type: application/json' \
  -d '{
    "cond": {
        "ID": 50
    },
    "doc": {
        "Emails": [
            {
                "ID": 101,
                "DeletedAt": "2019-06-25T17:05:06.000Z",
                "Email": "test111@example.com"
            },
            {
                "ID": 159,
                "DeletedAt": "2019-06-25T17:05:06.000Z",
                "Email": "test222@example.com"
            }
        ]
    },
    "auto": true
}'
```

#### Query associations

set `"preload=Emails"` to preload associations:

```sh
curl -X GET \
  'http://localhost:8080/api/user?cond={"user":"test"}&sort=id&project=pass&preload=Emails'
```

### Global default callbacks

You can override the default callbacks:

```go
// Default create callback
kuu.CreateCallback = func(scope *gorm.Scope) {
	if !scope.HasError() {
		if desc := GetRoutinePrivilegesDesc(); desc != nil {
			if orgIDField, ok := scope.FieldByName("OrgID"); ok {
				if orgIDField.IsBlank {
					if err := orgIDField.Set(desc.SignOrgID); err != nil {
						ERROR("自动设置组织ID失败：%s", err.Error())
					}
				}
			}
			if createdByField, ok := scope.FieldByName("CreatedByID"); ok {
				if err := createdByField.Set(desc.UID); err != nil {
					ERROR("自动设置创建人ID失败：%s", err.Error())
				}
			}

			if updatedByField, ok := scope.FieldByName("UpdatedByID"); ok {
				if err := updatedByField.Set(desc.UID); err != nil {
					ERROR("自动设置修改人ID失败：%s", err.Error())
				}
			}
		}
	}
}


// Default delete callback
kuu.DeleteCallback = func(scope *gorm.Scope) {
	if !scope.HasError() {
		var extraOption string
		if str, ok := scope.Get("gorm:delete_option"); ok {
			extraOption = fmt.Sprint(str)
		}

		deletedAtField, hasDeletedAtField := scope.FieldByName("DeletedAt")

		if !scope.Search.Unscoped && hasDeletedAtField {
			var sql string
			if desc := GetRoutinePrivilegesDesc(); desc != nil {
				deletedByField, hasDeletedByField := scope.FieldByName("DeletedByID")
				if !scope.Search.Unscoped && hasDeletedByField {
					sql = fmt.Sprintf(
						"UPDATE %v SET %v=%v,%v=%v%v%v",
						scope.QuotedTableName(),
						scope.Quote(deletedByField.DBName),
						scope.AddToVars(desc.UID),
						scope.Quote(deletedAtField.DBName),
						scope.AddToVars(gorm.NowFunc()),
						AddExtraSpaceIfExist(scope.CombinedConditionSql()),
						AddExtraSpaceIfExist(extraOption),
					)
				}
			}
			if sql == "" {
				sql = fmt.Sprintf(
					"UPDATE %v SET %v=%v%v%v",
					scope.QuotedTableName(),
					scope.Quote(deletedAtField.DBName),
					scope.AddToVars(gorm.NowFunc()),
					AddExtraSpaceIfExist(scope.CombinedConditionSql()),
					AddExtraSpaceIfExist(extraOption),
				)
			}
			scope.Raw(sql).Exec()
		} else {
			scope.Raw(fmt.Sprintf(
				"DELETE FROM %v%v%v",
				scope.QuotedTableName(),
				AddExtraSpaceIfExist(scope.CombinedConditionSql()),
				AddExtraSpaceIfExist(extraOption),
			)).Exec()
		}
	}
}

// Default update callback
kuu.UpdateCallback = func(scope *gorm.Scope) {
	if !scope.HasError() {
		desc := GetRoutinePrivilegesDesc()
		if desc != nil {
			if err := scope.SetColumn("UpdatedByID", desc.UID); err != nil {
				ERROR("自动设置修改人ID失败：%s", err.Error())
			}
		}
	}
}

// Default query callback
kuu.QueryCallback = func(scope *gorm.Scope) {
	if !scope.HasError() {
		desc := GetRoutinePrivilegesDesc()
		caches := GetRoutineCaches()

		if desc == nil {
			// 无登录登录态时
			return
		}

		if caches != nil {
			// 有忽略标记时
			if _, ignoreAuth := caches[GLSIgnoreAuthKey]; ignoreAuth {
				return
			}
			// 查询用户菜单时
			if _, queryUserMenus := caches[GLSUserMenusKey]; queryUserMenus {
				if desc.NotRootUser() {
					_, hasCodeField := scope.FieldByName("Code")
					_, hasCreatedByIDField := scope.FieldByName("CreatedByID")
					if hasCodeField && hasCreatedByIDField {
						// 菜单数据权限控制与组织无关，且只有两种情况：
						// 1.自己创建的，一定看得到
						// 2.别人创建的，必须通过分配操作权限才能看到
						scope.Search.Where("(code in (?)) OR (created_by_id = ?)", desc.Codes, desc.UID)
					}
				}
				return
			}
		}

		// 有登录态时的常规情况
		if desc.NotRootUser() {
			_, hasOrgIDField := scope.FieldByName("OrgID")
			_, hasCreatedByIDField := scope.FieldByName("CreatedByID")
			if hasOrgIDField && hasCreatedByIDField {
				scope.Search.Where("(org_id IS NULL) OR (org_id = 0) OR (org_id in (?)) OR (created_by_id = ?)", desc.ReadableOrgIDs, desc.UID)
			}
		}
	}
}

// Default validate callback
kuu.ValidateCallback = func(scope *gorm.Scope) {
	if _, ok := scope.Get("gorm:update_column"); !ok {
		if result, ok := scope.DB().Get(skipValidations); !(ok && result.(bool)) {
			if !scope.HasError() {
				scope.CallMethod("Validate")
				if scope.Value != nil {
					resource := scope.IndirectValue().Interface()
					_, validatorErrors := govalidator.ValidateStruct(resource)
					if validatorErrors != nil {
						if errors, ok := validatorErrors.(govalidator.Errors); ok {
							for _, err := range FlatValidatorErrors(errors) {
								if err := scope.DB().AddError(formattedValidError(err, resource)); err != nil {
									ERROR("添加验证错误信息失败：%s", err.Error())
								}

							}
						} else {
							if err := scope.DB().AddError(validatorErrors); err != nil {
								ERROR("添加验证错误信息失败：%s", err.Error())
							}
						}
					}
				}
			}
		}
	}
}
```

### Struct validation

base on [govalidator](https://github.com/asaskevich/govalidator):

```go
// this struct definition will fail govalidator.ValidateStruct() (and the field values do not matter):
type exampleStruct struct {
  Name  string ``
  Email string `valid:"email"`
}

// this, however, will only fail when Email is empty or an invalid email address:
type exampleStruct2 struct {
  Name  string `valid:"-"`
  Email string `valid:"email"`
}

// lastly, this will only fail when Email is an invalid email address but not when it's empty:
type exampleStruct2 struct {
  Name  string `valid:"-"`
  Email string `valid:"email,optional"`
}

// Validate
func (e *exampleStruct2) Validate () error {
}
```

### Modular project structure

Kuu will automatically mount routes, middlewares and struct RESTful APIs after `Import`:

```go
type User struct {
	kuu.Model `rest`
	Username string
	Password string
}

type Profile struct {
	kuu.Model `rest`
	Nickname string
	Age int
}

func MyMod() *kuu.Mod {
	return &kuu.Mod{
		Models: []interface{}{
			&User{},
			&Profile{},
		},
		Middlewares: gin.HandlersChain{
			func(c *gin.Context) {
				// Auth middleware
			},
		},
		Routes: kuu.RoutesInfo{
			kuu.RouteInfo{
				Method: "POST",
				Path:   "/login",
				HandlerFunc: func(c *kuu.Context) {
					// POST /login
				},
			},
			kuu.RouteInfo{
				Method: "POST",
				Path:   "/logout",
				HandlerFunc: func(c *kuu.Context) {
					// POST /logout
				},
			},
		},
	}
}

func main() {
	r := kuu.Default()
	r.Import(kuu.Acc(), kuu.Sys())     // import preset modules
	r.Import(MyMod())                       // import custom module
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
	r := kuu.Default()
	r.GET("/ping", func(c *kuu.Context) {
        c.STD("hello")                      // response: {"data":"hello","code":0}
        c.STD("hello", "Success")           // response: {"data":"hello","code":0,"msg":"Success"}
        c.STD(200)                          // response: {"data":200,"code":0}
        c.STDErr("New record failed")       // response: {"code":-1,"msg":"New record failed"}
        c.STDErr("New record failed", err)  // response: {"code":-1,"msg":"New record failed","data":"错误详细描述信息，对应err.Error()"}
        c.STDErrHold("Token decoding failed", err).Code(555).Render()  // response: {"code":555,"msg":"Token decoding failed","data":"错误详细描述信息，对应err.Error()"}
    })
}
```

**Notes:**

- If `data == error`, Kuu will call `ERROR(data)` to output the log.
- All message will call `kuu.L(c, msg)` for i18n before the response.

### Common functions

```go
func main() {
	r := kuu.Default()
	// Parse JSON from string
	var params map[string]string
	kuu.Parse(`{"user":"kuu","pass":"123"}`, &params)
	// Formatted as JSON
	kuu.Stringify(&params)
}
```

### Get login context

```go
r.GET(func (c *kuu.Context){
	c.SignInfo // Login user info
	c.PrisDesc // Login user privileges
})
```

### Goroutine local storage

```go
// preset caches
kuu.GetRoutinePrivilegesDesc()
kuu.GetRoutineValues()
kuu.GetRoutineRequestContext()

// custom caches
kuu.GetRoutineCaches()
kuu.SetRoutineCache(key, value)
kuu.GetRoutineCache(key)
kuu.DelRoutineCache(key)

// Ignore default data filters
kuu.IgnoreAuth() // Equivalent to c.IgnoreAuth/kuu.GetRoutineValues().IgnoreAuth
```

### Whitelist

All routes are blocked by the authentication middleware by default. If you want to ignore some routes, please configure the whitelist:

```go
kuu.AddWhitelist("GET /", "GET /user")
kuu.AddWhitelist(regexp.MustCompile("/user"))
```

> Notes: Whitelist also matches paths with global `prefix`. If you don't want this feature, please set `"whitelist:prefix":false`.

### Preset modules

- [Accounts module](https://github.com/kuuland/kuu/blob/master/acc.go#L153) - JWT-based token issuance, login authentication, etc.
- [System module](https://github.com/kuuland/kuu/blob/master/sys.go#L564) - Menu, admin, roles, organization, etc.
- [Admin](https://github.com/kuuland/ui) - A React boilerplate.

### Security framework

<img src="/docs/kuu_security_framework.png" alt="Kuu Security framework"/>

## FAQ

### Why called Kuu?

> [Kuu and Shino](https://www.youtube.com/results?search_query=kuu+shino)

<img src="/docs/kuu_and_shino.png" alt="Kuu and Shino"/>

## License
 
Kuu is available under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0.html).
