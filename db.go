package kuu

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"sync"
)

var (
	dataSourcesMap sync.Map
	singleDSName   = "kuu_default_db"
)

type datasource struct {
	Name    string
	Dialect string
	Args    string
}

func initDataSources() {
	pairs := C().Keys
	dbConfig, has := pairs["db"]
	if !has {
		return
	}
	if _, ok := dbConfig.([]interface{}); ok {
		// Multiple data sources
		var dsArr []datasource
		GetSoul(dbConfig, &dsArr)
		if len(dsArr) > 0 {
			var first string
			for _, ds := range dsArr {
				if IsBlank(ds) || ds.Name == "" {
					continue
				}
				if _, ok := dataSourcesMap.Load(ds.Name); ok {
					continue
				}
				if first == "" {
					first = ds.Name
				}
				db, err := gorm.Open(ds.Dialect, ds.Args)
				if err != nil {
					panic(err)
				} else {
					dataSourcesMap.Store(ds.Name, db)
					if gin.IsDebugging() {
						db.LogMode(true)
					}
				}
			}
			if first != "" {
				singleDSName = first
			}
		}
	} else {
		// Single data source
		var ds datasource
		GetSoul(dbConfig, &ds)
		if !IsBlank(ds) {
			if ds.Name == "" {
				ds.Name = singleDSName
			} else {
				singleDSName = ds.Name
			}
			db, err := gorm.Open(ds.Dialect, ds.Args)
			if err != nil {
				panic(err)
			} else {
				dataSourcesMap.Store(ds.Name, db)
				if gin.IsDebugging() {
					db.LogMode(true)
				}
			}
		}
	}
}

// DB
func DB(c ...*gin.Context) *gorm.DB {
	return DBWithName("", c...)
}

// DBWithName
func DBWithName(name string, c ...*gin.Context) *gorm.DB {
	if name == "" {
		name = singleDSName
	}
	if v, ok := dataSourcesMap.Load(name); ok {
		db := v.(*gorm.DB)
		if len(c) > 0 && c[0] != nil {
			//ctx := c[0]
			//orgID := ParseOrgID(ctx)
			//sign := ensureLogged(ctx)
			//if v, exists := ctx.Get(SignContextKey); exists {
			//	sign = v.(*SignContext)
			//}
			//// TODO 过滤数据范围
			//if orgID != 0 {
			//	db.Where("org_id in (?) or createdby_id = ?", orgID)
			//}
		}
		return db
	}
	PANIC("No data source named \"%s\"", name)
	return nil
}

// Init
func Init() {
	initDataSources()
}

// Release
func Release() {
	dataSourcesMap.Range(func(_, value interface{}) bool {
		db := value.(*gorm.DB)
		db.Close()
		return true
	})
}
