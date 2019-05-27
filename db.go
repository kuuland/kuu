package kuu

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/jtolds/gls"
	"sync"
)

var (
	dataSourcesMap sync.Map
	mgr            = gls.NewContextManager()
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
					connectedPrint(Capitalize(db.Dialect().GetName()), db.Dialect().CurrentDatabase())
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
				connectedPrint(Capitalize(db.Dialect().GetName()), db.Dialect().CurrentDatabase())
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
			// 查询授权规则
			c := ginContext[0]
			desc := GetPrivilegesDesc(c)
			db = DefaultWhereHandler(db, desc, c)
		}
		return db
	}
	PANIC("No data source named \"%s\"", name)
	return nil
}

// WithTransaction
func WithTransaction(fn func(*gorm.DB) error, with ...*gorm.DB) error {
	var (
		tx  *gorm.DB
		out bool
	)
	if len(with) > 0 {
		tx = with[0]
		out = true
	} else {
		tx = DB().Begin()
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := tx.Error; err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		return err
	}
	if out {
		return nil
	}
	if errs := tx.GetErrors(); len(errs) > 0 {
		if err := tx.Rollback().Error; err != nil {
			return err
		}
	}
	return tx.Commit().Error
}

// Release
func Release() {
	dataSourcesMap.Range(func(_, value interface{}) bool {
		db := value.(*gorm.DB)
		db.Close()
		return true
	})
}
