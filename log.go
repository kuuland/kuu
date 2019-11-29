package kuu

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/mssola/user_agent"
	uuid "github.com/satori/go.uuid"
	"log"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	LogTypeSign  = "sign"
	LogTypeAPI   = "api"
	LogTypeAudit = "audit"
	LogTypeBiz   = "biz"
)

const (
	AuditTypeCreate = "create"
	AuditTypeUpdate = "update"
	AuditTypeRemove = "remove"
)

var (
	Logger        = logrus.New()
	DailyFileName = fmt.Sprintf("kuu-%s.log", time.Now().Format("2006-01-02"))
	DailyFile     *os.File
	LogDir        string
)

func init() {
	initEnums()
	if C().GetString("env") == "prod" {
		IsProduction = true
	}
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)
	if IsProduction {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		Logger.SetFormatter(&logrus.JSONFormatter{})
		log.Println("==> 生产环境自动启用文件模式存储日志")
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{})
		Logger.SetFormatter(&logrus.TextFormatter{})
	}
	LogDir = C().DefaultGetString("logs", "logs")
	if LogDir != "" {
		Logger.AddHook(&LogDailyFileHook{})
		Logger.AddHook(&LogBizHook{})
	}
}

func initEnums() {
	Enum("LogType", "日志类型").
		Add(LogTypeSign, "登录日志").
		Add(LogTypeAPI, "接口日志").
		Add(LogTypeAudit, "审计日志").
		Add(LogTypeBiz, "业务日志")

	Enum("AuditType", "审计类型").
		Add(AuditTypeCreate, "新增操作").
		Add(AuditTypeUpdate, "修改操作").
		Add(AuditTypeRemove, "删除操作")
}

type LogIgnorer interface {
	IgnoreLog()
}

// Log
type Log struct {
	gorm.Model   `rest:"*" displayName:"系统日志"`
	UUID         string `name:"数据ID（UUID）" rest:"*" displayName:"系统日志"`
	Time         int64  `name:"记录时间（Unix时间戳）"`
	Type         string `name:"日志类型" enum:"LogType"`
	ContentHuman string `name:"日志内容（可读描述）"`
	ContentData  string `name:"日志详情（完整JSON）" gorm:"type:text"`
	// 用户信息
	UID        uint   `name:"用户ID" sql:"index"`
	SubDocID   uint   `name:"用户子档案ID"`
	Username   string `name:"用户账号"`
	RealName   string `name:"姓名"`
	Token      string `name:"使用令牌" gorm:"type:text"`
	ActOrgID   uint   `name:"当前组织ID"`
	ActOrgCode string `name:"当前组织编码"`
	ActOrgName string `name:"当前组织名称"`
	// 认证信息
	SignMethod  string `name:"登录/登出"`
	SignType    string `name:"令牌类型"`
	SignPayload string `name:"登录数据" gorm:"type:text"`
	// 请求信息
	RequestUserAgent                  string        `name:"原始User-Agent" gorm:"type:text"`
	RequestMethod                     string        `name:"请求方法"`
	RequestPath                       string        `name:"请求接口"`
	RequestContentLength              int64         `name:"请求体大小"`
	RequestReferer                    string        `name:"请求Referer"`
	RequestIsWebsocket                bool          `name:"是否websocket请求"`
	RequestIsMobile                   bool          `name:"是否移动端请求"`
	RequestContentType                string        `name:"原始Content-Type"`
	RequestHeaders                    string        `name:"请求头" gorm:"type:text"`
	RequestQuery                      string        `name:"查询参数"`
	RequestCost                       time.Duration `name:"调用耗时"`
	RequestIP                         string        `name:"调用IP"`
	RequestCountry                    string        `name:"调用国家"`
	RequestCity                       string        `name:"调用城市"`
	RequestClientLocalization         string        `name:"终端语言环境"`
	RequestClientBrowserEngine        string        `name:"终端浏览器内核"`
	RequestClientBrowserEngineVersion string        `name:"终端浏览器内核版本"`
	RequestClientBrowserName          string        `name:"终端浏览器名称"`
	RequestClientBrowserVersion       string        `name:"终端浏览器版本"`
	RequestClientPlatform             string        `name:"终端平台"`
	RequestClientOSFullName           string        `name:"终端系统全名"`
	RequestClientOSName               string        `name:"终端系统名称"`
	RequestClientOSVersion            string        `name:"终端系统版本"`
	RequestErrorMessage               string        `name:"请求错误信息"`
	ResponseStatusCode                int           `name:"响应状态码"`
	ResponseBodySize                  int           `name:"响应体大小"`
	// 审计信息
	AuditType    string `name:"操作类型" enum:"AuditType"`
	AuditTag     string `name:"审计标记（自定义扩展，默认为system）"`
	AuditModel   string `name:"模型名称"`
	AuditSQL     string `name:"SQL" gorm:"type:text"`
	AuditSQLVars string `name:"SQL参数" gorm:"type:text"`
	// 业务日志
	Level string `name:"日志级别"`
}

// BeforeCreate
func (log *Log) BeforeCreate() {
	if log.UUID == "" {
		log.UUID = uuid.NewV4().String()
	}
	if log.Type == LogTypeAudit && log.AuditTag == "" {
		log.AuditTag = "system"
	}
}

// IgnoreLog
func (log *Log) IgnoreLog() {}

// BeforeCreate
func (log *Log) RepairDBTypes() {
	var (
		db        = DB()
		scope     = db.NewScope(log)
		tableName = scope.QuotedTableName()
		textType  = C().DefaultGetString("logs:textType", "MEDIUMTEXT")
	)
	if db.Dialect().GetName() == "mysql" {
		fields := []string{
			"content_data",
			"token",
			"sign_payload",
			"request_user_agent",
			"request_headers",
			"audit_sql",
			"audit_sql_vars",
		}
		for _, item := range fields {
			sql := fmt.Sprintf("ALTER TABLE %s MODIFY %s %s NULL", tableName, scope.Quote(item), textType)
			ERROR(db.Exec(sql).Error)
		}
	}
}

// CacheKey
func (log *Log) CacheKey() string {
	return BuildKey(
		"log",
		time.Unix(log.Time, 0).Format("20060102"),
		log.Type,
		log.UUID,
	)
}

// Save2Cache
func (log *Log) Save2Cache() {
	if key := log.CacheKey(); key != "" {
		SetCacheString(key, Stringify(log))
	}
}

// NewLog
func NewLog(logType string, context ...*gin.Context) (log *Log) {
	log = &Log{
		UUID: uuid.NewV4().String(),
		Time: time.Now().Unix(),
		Type: logType,
	}

	var c *gin.Context

	if len(context) > 0 {
		c = context[0]
	}

	if c == nil {
		if kc := GetRoutineRequestContext(); kc != nil {
			c = kc.Context

			signContext := kc.SignInfo
			if signContext == nil {
				if v, exists := c.Get(SignContextKey); exists {
					signContext = v.(*SignContext)
				}
			}

			if signContext != nil {
				user := GetUserFromCache(signContext.UID)
				log.UID = signContext.UID
				log.SubDocID = signContext.SubDocID
				log.Token = signContext.Token
				log.SignType = signContext.Type
				log.SignPayload = Stringify(signContext.Payload)
				log.Username = user.Username
				log.RealName = user.Name
			}

			if desc := kc.PrisDesc; desc != nil {
				log.ActOrgID = desc.ActOrgID
				log.ActOrgCode = desc.ActOrgCode
				log.ActOrgName = desc.ActOrgName
			}
		}
	}

	if desc := GetRoutinePrivilegesDesc(); desc != nil {
		log.ActOrgID = desc.ActOrgID
		log.ActOrgCode = desc.ActOrgCode
		log.ActOrgName = desc.ActOrgName
	}

	if c != nil {
		ua := user_agent.New(c.Request.UserAgent())
		engine, engineVersion := ua.Engine()
		browserName, browserVersion := ua.Browser()
		osInfo := ua.OSInfo()
		// 请求信息
		log.RequestUserAgent = c.Request.UserAgent()
		log.RequestMethod = c.Request.Method
		log.RequestPath = c.Request.URL.Path
		log.RequestContentLength = c.Request.ContentLength
		log.RequestReferer = c.Request.Referer()
		log.RequestIsWebsocket = c.IsWebsocket()
		log.RequestIsMobile = ua.Mobile()
		log.RequestContentType = c.ContentType()
		log.RequestHeaders = Stringify(c.Request.Header)
		log.RequestQuery = c.Request.URL.RawQuery
		log.RequestCost = 0 // 请求补全
		log.RequestIP = c.ClientIP()
		log.RequestCountry = "" // 请求补全
		log.RequestCity = ""    // 请求补全
		log.RequestClientLocalization = ua.Localization()
		log.RequestClientBrowserEngine = engine
		log.RequestClientBrowserEngineVersion = engineVersion
		log.RequestClientBrowserName = browserName
		log.RequestClientBrowserVersion = browserVersion
		log.RequestClientPlatform = ua.Platform()
		log.RequestClientOSFullName = osInfo.FullName
		log.RequestClientOSName = osInfo.Name
		log.RequestClientOSVersion = osInfo.Version
		log.RequestErrorMessage = "" // 请求补全
		log.ResponseStatusCode = 0   // 请求补全
		log.ResponseBodySize = 0     // 请求补全
	}
	return
}

// 日志序列化任务
func LogPersistenceTask() {
	err := WithTransaction(func(tx *gorm.DB) error {
		data := HasPrefixCache(BuildKey("log"), 5000)

		if len(data) == 0 {
			return nil
		}

		var (
			totalKeys  []string
			insertBase = fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s) VALUES ",
				tx.Dialect().Quote("sys_Log"),
				tx.Dialect().Quote("uuid"),
				tx.Dialect().Quote("time"),
				tx.Dialect().Quote("type"),
				tx.Dialect().Quote("content_human"),
				tx.Dialect().Quote("content_data"),
				// 用户信息
				tx.Dialect().Quote("uid"),
				tx.Dialect().Quote("sub_doc_id"),
				tx.Dialect().Quote("username"),
				tx.Dialect().Quote("real_name"),
				tx.Dialect().Quote("token"),
				// 认证信息
				tx.Dialect().Quote("sign_method"),
				tx.Dialect().Quote("sign_type"),
				tx.Dialect().Quote("sign_payload"),
				// 请求信息
				tx.Dialect().Quote("request_user_agent"),
				tx.Dialect().Quote("request_method"),
				tx.Dialect().Quote("request_path"),
				tx.Dialect().Quote("request_content_length"),
				tx.Dialect().Quote("request_referer"),
				tx.Dialect().Quote("request_is_websocket"),
				tx.Dialect().Quote("request_is_mobile"),
				tx.Dialect().Quote("request_content_type"),
				tx.Dialect().Quote("request_headers"),
				tx.Dialect().Quote("request_query"),
				tx.Dialect().Quote("request_cost"),
				tx.Dialect().Quote("request_ip"),
				tx.Dialect().Quote("request_country"),
				tx.Dialect().Quote("request_city"),
				tx.Dialect().Quote("request_client_localization"),
				tx.Dialect().Quote("request_client_browser_engine"),
				tx.Dialect().Quote("request_client_browser_engine_version"),
				tx.Dialect().Quote("request_client_browser_name"),
				tx.Dialect().Quote("request_client_browser_version"),
				tx.Dialect().Quote("request_client_platform"),
				tx.Dialect().Quote("request_client_os_full_name"),
				tx.Dialect().Quote("request_client_os_name"),
				tx.Dialect().Quote("request_client_os_version"),
				tx.Dialect().Quote("request_error_message"),
				tx.Dialect().Quote("response_status_code"),
				tx.Dialect().Quote("response_body_size"),
				// 审计信息
				tx.Dialect().Quote("audit_type"),
				tx.Dialect().Quote("audit_tag"),
				tx.Dialect().Quote("audit_model"),
				tx.Dialect().Quote("audit_sql"),
				tx.Dialect().Quote("audit_sql_vars"),
				// 业务日志
				tx.Dialect().Quote("level"),
			)
			insertItems []BatchInsertItem
		)
		for key, value := range data {
			totalKeys = append(totalKeys, key)
			var item Log
			if err := Parse(value, &item); err == nil && item.UUID != "" {
				insertItems = append(insertItems, BatchInsertItem{
					SQL: "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
					Vars: []interface{}{
						item.UUID,
						item.Time,
						item.Type,
						item.ContentHuman,
						item.ContentData,
						// 用户信息
						item.UID,
						item.SubDocID,
						item.Username,
						item.RealName,
						item.Token,
						// 认证信息
						item.SignMethod,
						item.SignType,
						item.SignPayload,
						// 请求信息
						item.RequestUserAgent,
						item.RequestMethod,
						item.RequestPath,
						item.RequestContentLength,
						item.RequestReferer,
						item.RequestIsWebsocket,
						item.RequestIsMobile,
						item.RequestContentType,
						item.RequestHeaders,
						item.RequestQuery,
						item.RequestCost,
						item.RequestIP,
						item.RequestCountry,
						item.RequestCity,
						item.RequestClientLocalization,
						item.RequestClientBrowserEngine,
						item.RequestClientBrowserEngineVersion,
						item.RequestClientBrowserName,
						item.RequestClientBrowserVersion,
						item.RequestClientPlatform,
						item.RequestClientOSFullName,
						item.RequestClientOSName,
						item.RequestClientOSVersion,
						item.RequestErrorMessage,
						item.ResponseStatusCode,
						item.ResponseBodySize,
						// 审计信息
						item.AuditType,
						item.AuditTag,
						item.AuditModel,
						item.AuditSQL,
						item.AuditSQLVars,
						// 业务日志
						item.Level,
					},
				})
			}
		}

		if err := BatchInsert(tx, insertBase, insertItems, 200); err != nil {
			return err
		} else {
			DelCache(totalKeys...)
		}

		return tx.Error
	})

	if err != nil {
		fmt.Println(err)
	}
}

// LogCleanupTask
func LogCleanupTask() {
	var (
		dest   = (24 * time.Hour) * 100 // 默认清除100天前的记录
		divide = time.Now().Add(-dest).Unix()
		db     = DB()
	)

	err := db.Unscoped().
		Where(fmt.Sprintf("%s < ?", db.Dialect().Quote("time")), divide).
		Delete(&Log{}).Error

	if err != nil {
		fmt.Println(err)
	}
}

func split(args ...interface{}) (string, []interface{}) {
	var (
		format string
		a      []interface{}
	)
	if v, ok := args[0].(string); ok {
		format = v
	} else {
		data, err := json.Marshal(v)
		if err == nil {
			format = string(data)
		}
	}
	if len(args) > 1 {
		a = args[1:]
	}
	return format, a
}

func isBlankArgs(args ...interface{}) bool {
	if len(args) == 0 || (len(args) == 1 && IsNil(args[0])) {
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
	if format != "" {
		Logger.Printf(format, a...)
	}
}

// DEBUG
func DEBUG(args ...interface{}) {
	if isBlankArgs(args...) {
		return
	}
	format, a := split(args...)
	if format != "" {
		Logger.Debugf(format, a...)
	}
}

// WARN
func WARN(args ...interface{}) {
	if isBlankArgs(args...) {
		return
	}
	format, a := split(args...)
	if format != "" {
		Logger.Warnf(format, a...)
	}
}

// INFO
func INFO(args ...interface{}) {
	if isBlankArgs(args...) {
		return
	}
	format, a := split(args...)
	if format != "" {
		Logger.Infof(format, a...)
	}
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
	if format != "" {
		Logger.Errorf(format, a...)
	}
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
	if format != "" {
		Logger.Fatalf(format, a...)
	}
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
	if format != "" {
		Logger.Panicf(format, a...)
	}
}
