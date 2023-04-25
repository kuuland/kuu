package kuu

import (
	"errors"
	"fmt"
	"github.com/samber/lo"
	"github.com/yukitsune/lokirus"
	"os"
	"reflect"
	"runtime"
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
	// logrus默认实例用于输出到控制台
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{})
	if C().GetString("env") == "prod" {
		IsProduction = true
	}
	if IsProduction {
		Logger.SetFormatter(&logrus.JSONFormatter{})
		logrus.Info("==> 生产环境自动启用文件模式存储日志")
	} else {
		Logger.SetFormatter(&logrus.TextFormatter{})
	}
	var loki = struct {
		Endpoint string
		Labels   lokirus.Labels
		Auth     struct {
			Username string
			Password string
		}
	}{}
	C().GetInterface("loki", &loki)
	if loki.Endpoint != "" {
		opts := lokirus.NewLokiHookOptions().
			WithLevelMap(lokirus.LevelMap{logrus.PanicLevel: "critical"}).
			WithFormatter(&logrus.TextFormatter{})
		if len(loki.Labels) > 0 {
			opts = opts.WithStaticLabels(loki.Labels)
		}
		if loki.Auth.Username != "" {
			opts = opts.WithBasicAuth(loki.Auth.Username, loki.Auth.Password)
		}
		hook := lokirus.NewLokiHookWithOpts(loki.Endpoint, opts)
		Logger.AddHook(hook)

	}
	LogDir = C().GetString("logs")
	if LogDir != "" {
		Logger.AddHook(&LogDailyFileHook{})
	}
}

func parseArgs(v []interface{}) (fields logrus.Fields, format string, args []interface{}, err error) {
	if len(v) == 0 || (len(v) == 1 && IsNil(v[0])) {
		return
	}
	if vv, ok := v[0].(string); ok {
		format = vv
		if len(v) > 1 {
			args = v[1:]
		}
	} else if vv, ok := v[0].(error); ok {
		format = vv.Error()
		if len(v) > 1 {
			args = v[1:]
		}
	} else if vv, ok := v[0].([]error); ok {
		format = fmt.Sprintf("%v", vv)
		if len(v) > 1 {
			args = v[1:]
		}
	} else {
		indirectValue := reflect.Indirect(reflect.ValueOf(v[0]))
		switch indirectValue.Kind() {
		case reflect.Map, reflect.Struct:
			_ = Copy(v[0], &fields)
		default:
			fields = make(logrus.Fields)
			fields["data"] = fmt.Sprintf("%v", v[0])
		}
		if len(v) > 1 {
			if vv, ok := v[1].(string); ok {
				format = vv
				if len(v) > 2 {
					args = v[2:]
				}
			}
		}
	}
	if len(fields) == 0 && format == "" {
		return fields, format, args, errors.New("unsupported parameter format")
	}
	return
}

// PRINT
func PRINT(v ...interface{}) {
	PRINTWithFields(nil, v...)
}

func PRINTWithFields(fields logrus.Fields, v ...interface{}) {
	logWithFields(fields, func(fields logrus.Fields, format string, args ...interface{}) {
		Logger.WithFields(fields).Printf(format, args...)
	}, v...)
}

// DEBUG
func DEBUG(v ...interface{}) {
	DEBUGWithFields(nil, v...)
}

func DEBUGWithFields(fields logrus.Fields, v ...interface{}) {
	logWithFields(nil, func(fields logrus.Fields, format string, args ...interface{}) {
		Logger.WithFields(fields).Debugf(format, args...)
	}, v...)
}

// WARN
func WARN(v ...interface{}) {
	WARNWithFields(nil, v...)
}

func WARNWithFields(fields logrus.Fields, v ...interface{}) {
	logWithFields(fields, func(fields logrus.Fields, format string, args ...interface{}) {
		Logger.WithFields(fields).Warnf(format, args...)
	}, v...)
}

// INFO
func INFO(v ...interface{}) {
	INFOWithFields(nil, v...)
}

func INFOWithFields(fields logrus.Fields, v ...interface{}) {
	logWithFields(fields, func(fields logrus.Fields, format string, args ...interface{}) {
		Logger.WithFields(fields).Infof(format, args...)
	}, v...)
}

// ERROR
func ERROR(v ...interface{}) {
	ERRORWithFields(nil, v...)
}

func ERRORWithFields(fields logrus.Fields, v ...interface{}) {
	logWithFields(fields, func(fields logrus.Fields, format string, args ...interface{}) {
		Logger.WithFields(fields).Errorf(format, args...)
	}, v...)
}

// FATAL
func FATAL(v ...interface{}) {
	FATALWithFields(nil, v...)
}

func FATALWithFields(fields logrus.Fields, v ...interface{}) {
	logWithFields(fields, func(fields logrus.Fields, format string, args ...interface{}) {
		Logger.WithFields(fields).Fatalf(format, args...)
	}, v...)
}

// PANIC
func PANIC(v ...interface{}) {
	PANICWithFields(nil, v...)
}

func PANICWithFields(fields logrus.Fields, v ...interface{}) {
	logWithFields(fields, func(fields logrus.Fields, format string, args ...interface{}) {
		Logger.WithFields(fields).Panicf(format, args...)
	}, v...)
}

func logWithFields(extraFields logrus.Fields, callback func(fields logrus.Fields, format string, args ...interface{}), v ...interface{}) {
	fields, format, args, err := parseArgs(v)
	if err != nil {
		return
	}
	if fields == nil {
		fields = make(logrus.Fields)
	}
	if extraFields == nil {
		extraFields = make(logrus.Fields)
	}
	funcname := getCaller()
	if funcname != "" {
		extraFields["method"] = funcname
	}
	if len(extraFields) > 0 {
		for k, v := range extraFields {
			if vv, exists := fields[k]; exists {
				fields[fmt.Sprintf("fields.%s", k)] = vv
			}
			fields[k] = v
		}
	}

	if callback != nil {
		callback(fields, format, args...)
	}
}

func getCaller() string {
	var callers []*runtime.Func
	var index = 2
	for true {
		pc, _, _, ok := runtime.Caller(index)
		if ok {
			funcname := runtime.FuncForPC(pc)
			if !strings.HasPrefix(funcname.Name(), "github.com/kuuland/kuu") {
				break
			}
			callers = append(callers, funcname)
		}
		index += 1
	}

	var method string
	var ci = 3
	if len(callers) == 1 {
		ci = 3
	}
	var checks = [][]string{
		{"PRINTWithFields", "PRINT"},
		{"DEBUGWithFields", "DEBUG"},
		{"WARNWithFields", "WARN"},
		{"INFOWithFields", "INFO"},
		{"ERRORWithFields", "ERROR"},
		{"FATALWithFields", "FATAL"},
		{"PANICWithFields", "PANIC"},
	}
	if len(callers) == 2 {
		for _, check := range checks {
			if check[0] == _getMethodByCaller(callers[0]) && check[1] == _getMethodByCaller(callers[1]) {
				ci = 4
				break
			}
		}
	}
	if len(callers) == 3 {
		ci = 5
	}
	pc, _, line, ok := runtime.Caller(ci)
	if ok {
		funcname := runtime.FuncForPC(pc)
		method = fmt.Sprintf("%s:L%d", funcname.Name(), line)
	}
	if strings.HasPrefix(method, "github.com/kuuland/kuu") {
		method = ""
	}
	if strings.HasPrefix(method, "runtime.") {
		method = ""
	}
	return method
}

func _getMethodByCaller(fn *runtime.Func) string {
	splits := strings.Split(fn.Name(), ".")
	method, _ := lo.Last(splits)
	return method
}
