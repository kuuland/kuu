package kuu

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"os"
	"sync"

	"github.com/jinzhu/gorm"
)

var (
	pairs          map[string]interface{}
	inst           *Config
	dataSourcesMap sync.Map
	singleDSName   = "kuu_default_db"
)

// Parse config file when init
func init() {
	filePath := os.Getenv("CONFIG_FILE")
	if IsBlank(filePath) {
		filePath = "kuu.json"
	}
	pairs = make(map[string]interface{})
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return
	}
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		ERROR(err)
	} else {
		err := json.Unmarshal(data, &pairs)
		if err != nil {
			ERROR(err)
		}
	}
	initDataSources()

}

type Config struct {
	Keys map[string]interface{}
}

func C() *Config {
	if inst == nil {
		inst = &Config{Keys: pairs}
	}
	return inst
}

func (c *Config) Get(key string) (value interface{}, exists bool) {
	value, exists = c.Keys[key]
	return
}

// GetString returns the value associated with the key as a string.
func (c *Config) GetString(key string) (s string) {
	if val, ok := c.Get(key); ok && val != nil {
		s, _ = val.(string)
	}
	return
}

// GetBool returns the value associated with the key as a boolean.
func (c *Config) GetBool(key string) (b bool) {
	if val, ok := c.Get(key); ok && val != nil {
		b, _ = val.(bool)
	}
	return
}

// GetInt returns the value associated with the key as an integer.
func (c *Config) GetInt(key string) (i int) {
	if val, ok := c.Get(key); ok && val != nil {
		i, _ = val.(int)
	}
	return
}

// GetInt64 returns the value associated with the key as an integer.
func (c *Config) GetInt32(key string) (i32 int32) {
	if val, ok := c.Get(key); ok && val != nil {
		i32, _ = val.(int32)
	}
	return
}

// GetInt64 returns the value associated with the key as an integer.
func (c *Config) GetInt64(key string) (i64 int64) {
	if val, ok := c.Get(key); ok && val != nil {
		i64, _ = val.(int64)
	}
	return
}

// GetFloat64 returns the value associated with the key as a float32.
func (c *Config) GetFloat32(key string) (f32 float32) {
	if val, ok := c.Get(key); ok && val != nil {
		f32, _ = val.(float32)
	}
	return
}

// GetFloat64 returns the value associated with the key as a float64.
func (c *Config) GetFloat64(key string) (f64 float64) {
	if val, ok := c.Get(key); ok && val != nil {
		f64, _ = val.(float64)
	}
	return
}

type datasource struct {
	Name    string
	Dialect string
	Args    string
}

func initDataSources() {
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
func DB(name ...string) *gorm.DB {
	key := singleDSName
	if len(name) > 0 {
		key = name[0]
	}
	if v, ok := dataSourcesMap.Load(key); ok {
		return v.(*gorm.DB)
	}
	PANIC("No data source named \"%s\"", key)
	return nil
}

// Release
func Release() {
	dataSourcesMap.Range(func(_, value interface{}) bool {
		db := value.(*gorm.DB)
		db.Close()
		return true
	})
}
