package kuu

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"strings"
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
func DBWithName(name string, ginContext ...*gin.Context) *gorm.DB {
	if name == "" {
		name = singleDSName
	}
	if v, ok := dataSourcesMap.Load(name); ok {
		db := v.(*gorm.DB)
		if len(ginContext) > 0 && ginContext[0] != nil {
			// 解析登录信息
			c := ginContext[0]
			var sign *SignContext
			if v, exists := c.Get(SignContextKey); exists {
				sign = v.(*SignContext)
			} else {
				if v, err := DecodedContext(c); err == nil {
					sign = v
				}
			}
			orgID := ParseOrgID(c)
			// 查询授权规则
			var rule AuthRule
			queryDB := v.(*gorm.DB)
			if err := queryDB.Where(&AuthRule{UID: sign.UID, TargetOrgID: orgID}).First(&rule).Error; err == nil {
				var orgIDs []uint
				if rule.ReadableOrgIDs != "" {
					for _, item := range strings.Split(rule.ReadableOrgIDs, ",") {
						if v := ParseID(item); v != 0 {
							orgIDs = append(orgIDs, uint(v))
						}
					}
				}
				db.Where("(org_id IS NULL) OR (org_id in (?)) OR (created_by_id = ?)", orgIDs, sign.UID)
			}
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
