package kuu

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/jtolds/gls"
	"log"
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

type dbLogger struct {
	gorm.LogWriter
}

// Print format & print log
func (logger dbLogger) Print(values ...interface{}) {
	logger.Println(values...)
}

// Println format & print log
func (logger dbLogger) Println(values ...interface{}) {
	messages := gorm.LogFormatter(values...)
	if len(messages) > 0 {
		INFO(messages...)
		log.Println(messages...)
	}
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
		if err := Copy(dbConfig, &dsArr); err != nil {
			ERROR(err)
		}
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
						//db.SetLogger(dbLogger{})
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
		if err := Copy(dbConfig, &ds); err != nil {
			ERROR(err)
		}
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
					//db.SetLogger(dbLogger{})
				}
			}
		}
	}
}

// DB
func DB() *gorm.DB {
	return DS("")
}

// DS
func DS(name string) *gorm.DB {
	if name == "" {
		name = singleDSName
	}
	if v, ok := dataSourcesMap.Load(name); ok {
		db := v.(*gorm.DB)
		return db
	}
	PANIC("No data source named \"%s\"", name)
	return nil
}

// WithTransaction
func WithTransaction(fn func(*gorm.DB) (*gorm.DB, error), with ...*gorm.DB) error {
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
			ERROR(r)
			tx.Rollback()
		}
	}()
	if db, err := fn(tx); err != nil {
		return err
	} else if db.Error != nil {
		return db.Error
	} else {
		tx = db
	}
	if err := tx.Error; err != nil {
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
		if err := db.Close(); err != nil {
			ERROR(err)
		}
		return true
	})
}
