package kuu

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	Logger        = logrus.New()
	DailyFileName = fmt.Sprintf("kuu-%s.log", time.Now().Format("2006-01-02"))
	DailyFile     *os.File
	LogDir        string
)

func init() {
	if C().GetString("env") == "prod" {
		IsProduction = true
	}
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)
	if IsProduction {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		Logger.SetFormatter(&logrus.JSONFormatter{})
		log.Println("==> 生产环境自动启用文件模式存储日志")
		LogDir = C().DefaultGetString("logs", "logs")
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{})
		Logger.SetFormatter(&logrus.TextFormatter{})
		LogDir = C().GetString("logs")
	}
	if LogDir != "" {
		Logger.AddHook(new(DailyFileHook))
	}
}

// LogInfo
type LogInfo struct {
	Time *time.Time `name:"记录时间"`
	Type string     `name:"日志类型"`
	// 用户信息
	UID        uint   `name:"用户ID"`
	User       User   `name:"用户对象" gorm:"foreignkey:UID"`
	SubDocID   uint   `name:"用户子档案ID"`
	Token      string `name:"使用令牌"`
	SignMethod string `name:"登录/登出"`
	// 请求信息
	RequestID            string      `name:"请求唯一ID"`
	RequestMethod        string      `name:"请求方法"`
	RequestPath          string      `name:"请求接口"`
	RequestHeaders       http.Header `name:"请求头"`
	RequestQuery         url.Values  `name:"查询参数"`
	RequestCost          float64     `name:"调用耗时"`
	RequestIP            string      `name:"调用IP"`
	RequestAddr          string      `name:"调用地区"`
	RequestClientBrowser string      `name:"终端浏览器"`
	RequestClientType    string      `name:"终端类型"`
	RequestClientSys     string      `name:"终端系统"`
	// 审计信息
	AuditOp   string      `name:"操作类型"`
	ModelName string      `name:"模型名称"`
	SQL       interface{} `name:"SQL"`
	SQLVars   interface{} `name:"SQL参数"`
	// 日志详情
	Level        string `name:"日志级别"`
	ContentHuman string `name:"日志内容（可读描述）"`
	ContentData  string `name:"日志详情（完整JSON）"`
}

// Log4Sign
type Log4Sign struct {
	gorm.Model `rest:"*" displayName:"登录日志"`
	Time       *time.Time `name:"日志时间" sql:"index"`
	UID        uint       `name:"用户ID" sql:"index"`
	User       User       `name:"用户对象" gorm:"foreignkey:UID"`
	SubDocID   uint       `name:"用户子档案ID"`
	Token      string     `name:"使用令牌"`
	SignMethod string     `name:"登录/登出"`
}

// Log4API
type Log4API struct {
	gorm.Model `rest:"*" displayName:"API日志"`
	Time       *time.Time `name:"日志时间" sql:"index"`
	// 用户信息
	UID      uint   `name:"用户ID" sql:"index"`
	User     User   `name:"用户对象" gorm:"foreignkey:UID"`
	SubDocID uint   `name:"用户子档案ID"`
	Token    string `name:"使用令牌"`
	// 请求信息
	RequestID      string      `name:"请求唯一ID" sql:"index"`
	RequestMethod  string      `name:"请求方法" sql:"index"`
	RequestPath    string      `name:"请求接口" sql:"index"`
	RequestHeaders http.Header `name:"请求头"`
	RequestQuery   url.Values  `name:"查询参数"`
	RequestCost    float64     `name:"调用耗时"`
	RequestIP      string      `name:"调用IP"`
	RequestAddr    string      `name:"调用地区"`
}

// Log4Audit
type Log4Audit struct {
	gorm.Model `rest:"*" displayName:"审计日志"`
	Time       *time.Time `name:"日志时间" sql:"index"`
	// 用户信息
	UID       uint        `name:"用户ID" sql:"index"`
	User      User        `name:"用户对象" gorm:"foreignkey:UID"`
	SubDocID  uint        `name:"用户子档案ID"`
	Token     string      `name:"使用令牌"`
	AuditOp   string      `name:"操作类型"`
	ModelName string      `name:"模型名称"`
	SQL       interface{} `name:"SQL"`
	SQLVars   interface{} `name:"SQL参数"`
	Content   string      `name:"日志内容"`
}

// Log4Biz
type Log4Biz struct {
	gorm.Model `rest:"*" displayName:"业务日志"`
	Time       *time.Time `name:"日志时间" sql:"index"`
	Level      string     `name:"日志级别"`
	Content    string     `name:"日志内容"`
}

type DailyFileHook struct{}

func (h *DailyFileHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *DailyFileHook) Fire(entry *logrus.Entry) error {
	now := fmt.Sprintf("kuu-%s.log", time.Now().Format("2006-01-02"))
	if now != DailyFileName || DailyFile == nil {
		DailyFileName = now
		changeLoggerOutput(now)
	}
	log.Println(fmt.Sprintf("[KUU-%s] %s", strings.ToUpper(entry.Level.String()), entry.Message))
	return nil
}

func changeLoggerOutput(filePath string) {
	if LogDir == "" {
		return
	}
	EnsureDir(LogDir)
	filePath = path.Join(LogDir, filePath)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		Logger.Out = file
		if DailyFile != nil {
			_ = DailyFile.Close()
		}
		DailyFile = file
	} else {
		ERROR("创建日志文件失败，使用标准输出流输出日志")
	}
}

func split(args ...interface{}) (string, []interface{}) {
	format := args[0].(string)
	var a []interface{}
	if len(args) > 1 {
		a = args[1:len(args)]
	}
	return format, a
}

func isBlankArgs(args ...interface{}) bool {
	if len(args) == 0 || (len(args) == 1 || args[0] == nil) {
		return true
	}
	return false
}

// PRINT
func PRINT(args ...interface{}) {
	if isBlankArgs(args...) {
		return
	}
	format, a := split(args...)
	Logger.Printf(format, a...)
}

// DEBUG
func DEBUG(args ...interface{}) {
	if isBlankArgs(args...) {
		return
	}
	format, a := split(args...)
	Logger.Debugf(format, a...)
}

// WARN
func WARN(args ...interface{}) {
	if isBlankArgs(args...) {
		return
	}
	format, a := split(args...)
	Logger.Warnf(format, a...)
}

// INFO
func INFO(args ...interface{}) {
	if isBlankArgs(args...) {
		return
	}
	format, a := split(args...)
	Logger.Infof(format, a...)
}

// ERROR
func ERROR(args ...interface{}) {
	if isBlankArgs(args...) {
		return
	}
	switch args[0].(type) {
	case error:
		args[0] = args[0].(error).Error()
	case []error:
		for _, e := range args[0].([]error) {
			ERROR(e)
		}
		return
	}
	format, a := split(args...)
	Logger.Errorf(format, a...)
}

// FATAL
func FATAL(args ...interface{}) {
	if isBlankArgs(args...) {
		return
	}
	switch args[0].(type) {
	case error:
		args[0] = args[0].(error).Error()
	}
	format, a := split(args...)
	Logger.Fatalf(format, a...)
}

// PANIC
func PANIC(args ...interface{}) {
	if isBlankArgs(args...) {
		return
	}
	switch args[0].(type) {
	case error:
		args[0] = args[0].(error).Error()
	}
	format, a := split(args...)
	Logger.Panicf(format, a...)
}
