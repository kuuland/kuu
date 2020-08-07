package kuu

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/jtolds/gls"
)

var (
	dataSourcesMap sync.Map
	mgr            = gls.NewContextManager()
	singleDSName   = "kuu_default_db"
)

type dataSource struct {
	Name    string
	Dialect string
	Args    string
}

func (ds *dataSource) isBlank() bool {
	return !(ds != nil && ds.Dialect != "" && ds.Args != "")
}

type DBTypeRepairer interface {
	RepairDBTypes()
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
		if r := GetRoutineRequestID(); r != "" {
			tmp := []interface{}{fmt.Sprintf("%s=%s ", GLSRequestIDKey, r)}
			tmp = append(tmp, messages...)
			messages = tmp
		}
		Logger.Info(messages...)
	}
}

func initDataSources() {
	raw, has := C().Get("db")
	if !has {
		return
	}
	var dataSources []dataSource
	if err := json.Unmarshal(raw, &dataSources); err != nil {
		var dataSource dataSource
		if err := json.Unmarshal(raw, &dataSource); err == nil {
			dataSources = append(dataSources, dataSource)
		}
	}
	if len(dataSources) == 0 {
		return
	}
	var firstDSName string
	for _, ds := range dataSources {
		if ds.Name == "" {
			ds.Name = singleDSName
		}
		if _, ok := dataSourcesMap.Load(ds.Name); ok {
			continue
		}
		if firstDSName == "" {
			firstDSName = ds.Name
			singleDSName = firstDSName
		}
		openDB(ds)
	}
}

func openDB(ds dataSource) {
	db, err := gorm.Open(ds.Dialect, ds.Args)
	if err != nil {
		panic(err)
	} else {
		connectedPrint(strings.Title(db.Dialect().GetName()), db.Dialect().CurrentDatabase())
		dataSourcesMap.Store(ds.Name, db)
		if gin.IsDebugging() {
			db.LogMode(true)
			db.SetLogger(dbLogger{})
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
		return db.
			Set("gorm:association_autoupdate", false).
			Set("gorm:association_autocreate", false)
	}
	PANIC("No data source named \"%s\"", name)
	return nil
}

// WithTransaction
func WithTransaction(fn func(*gorm.DB) error) (err error) {
	tx := DB().Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	return tx.Rollback().Error
}

func releaseDB() {
	dataSourcesMap.Range(func(_, value interface{}) bool {
		db := value.(*gorm.DB)
		if err := db.Close(); err != nil {
			ERROR(err)
		}
		return true
	})
}

// AutoMigrate
func AutoMigrate(values ...interface{}) *gorm.DB {
	return DB().AutoMigrate(values...)
}
