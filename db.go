package kuu

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/gin-gonic/gin"
	mysql_driver "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/jtolds/gls"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"sync"
)

var (
	dataSourcesMap sync.Map
	mgr            = gls.NewContextManager()
	singleDSName   = "kuu_default_db"
)

type dataSource struct {
	Name      string
	Dialect   string
	Args      string
	EnableTLS bool
	TLSName   string
	CAPath    string
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

	if ds.EnableTLS {
		RegisterDBTLS(ds.TLSName, ds.CAPath)
		ds.Args = ResetDSN(ds.Dialect, ds.Args, ds.TLSName)
	}

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
		err = tx.Error
		return
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit().Error
		}
	}()
	err = fn(tx)
	return
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

func RegisterDBTLS(tlsName string, caPath string) {
	var err error
	fmt.Printf("using TLS for database connection\n")

	// 路径为空或文件不存在
	if caPath == "" {
		fmt.Printf("DBCAPath is empty\n")
		panic(fmt.Errorf("DBCAPath is empty"))
	}

	var info os.FileInfo
	info, err = os.Stat(caPath)
	if os.IsNotExist(err) {
		fmt.Printf("DBCAPath is not exist\n")
		panic(fmt.Errorf("DBCAPath is not exist"))
	}
	if info.IsDir() {
		fmt.Printf("DBCAPath is a directory\n")
		panic(fmt.Errorf("DBCAPath is a directory"))
	}

	rootCertPool := x509.NewCertPool()
	pem, err := ioutil.ReadFile(caPath)
	if err != nil {
		panic(err)
	}
	if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
		panic("Failed to append PEM.")
	}

	tlsConfig := &tls.Config{
		RootCAs: rootCertPool,
	}

	err = mysql_driver.RegisterTLSConfig(tlsName, tlsConfig)
	if err != nil {
		panic("Failed to register TLS config " + err.Error())
	}
}

func ResetDSN(driver string, dsn string, tlsName string) string {
	// 使用正则表达式删除 tls=xxxxx
	// 匹配 tls= 后面的任意字符，直到遇到 & 或字符串结尾
	switch driver {
	case "mysql":
		re := regexp.MustCompile(`tls=[^&]*`)
		cleanedDsn := re.ReplaceAllString(dsn, "")

		// 如果 tls=xxxxx 是第一个参数，需要删除后面的 &，否则会导致多余的 & 出现
		// 例如：?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai
		// 应该变为：?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai

		if strings.Contains(cleanedDsn, "?") {
			dsn = cleanedDsn + fmt.Sprintf("&tls=%s", tlsName)
		} else {
			dsn = cleanedDsn + fmt.Sprintf("?tls=%s", tlsName)
		}
	default:
		fmt.Printf("driver %s not support tls\n", driver)
	}

	return dsn
}
