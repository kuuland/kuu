package db

import (
	"context"
	"fmt"
	"github.com/kuuland/kuu/v3"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

var (
	dbMap                 = new(sync.Map)
	defaultDataSourceName = "DEFAULT_DATA_SOURCE_NAME"
)

type Config struct {
	GORMConfig *gorm.Config
	Name       string
}

func Register(dialector gorm.Dialector, config ...*Config) error {
	conf := &Config{
		GORMConfig: &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
			Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
				SlowThreshold:             5 * time.Second,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: false,
				Colorful:                  true,
			}),
		},
		Name: defaultDataSourceName,
	}
	if v := os.Getenv("KUU_DB_LOGGER"); v == "1" {
		conf.GORMConfig.Logger = logger.Default.LogMode(logger.Info)
	}
	if len(config) > 0 && config[0] != nil {
		if config[0].GORMConfig != nil {
			conf.GORMConfig = config[0].GORMConfig
		}
		if config[0].Name != "" {
			conf.Name = config[0].Name
		}
	}

	ins, err := gorm.Open(dialector, conf.GORMConfig)
	if err != nil {
		return err
	}
	_ = ins.Callback().Delete().Before("gorm:before_delete").Register("kuu:before_delete", func(tx *gorm.DB) {
		if tx.Statement.Schema != nil {
			for _, field := range tx.Statement.Schema.Fields {
				switch tx.Statement.ReflectValue.Kind() {
				case reflect.Slice, reflect.Array:
					for i := 0; i < tx.Statement.ReflectValue.Len(); i++ {
						// Set value to field
						if field.Name == "Dr" || field.DBName == "dr" {
							if field.GORMDataType == schema.String {
								if err := field.Set(tx.Statement.Context, tx.Statement.ReflectValue.Index(i), fmt.Sprintf("%d", time.Now().UnixNano())); err != nil {
									_ = tx.AddError(err)
									return
								}
							}
						}
					}
				case reflect.Struct:
					// Set value to field
					if strings.ToUpper(field.Name) == "DR" {
						if field.GORMDataType == schema.String {
							if err := field.Set(tx.Statement.Context, tx.Statement.ReflectValue, fmt.Sprintf("%d", time.Now().UnixNano())); err != nil {
								_ = tx.AddError(err)
								return
							}
						}
					}
				}
			}
		}
	})
	dbMap.Store(conf.Name, ins)
	kuu.Infof("Database is connected! (%s)\n", conf.Name)
	return nil
}

func Unregister(name string) {
	dbMap.Delete(name)
}

func Inst(ctx context.Context, name ...string) *gorm.DB {
	n := defaultDataSourceName
	if len(name) > 0 && name[0] != "" {
		n = name[0]
	}
	v, has := dbMap.Load(n)
	if !has || v == nil {
		panic(fmt.Errorf("data source [%s] is not registered", n))
	}
	inst := v.(*gorm.DB)
	if ctx == nil {
		ctx = context.Background()
	}
	return inst.WithContext(ctx)
}
