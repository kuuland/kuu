package kuu

import (
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
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
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{})
		Logger.SetFormatter(&logrus.TextFormatter{})
	}
	LogDir = C().DefaultGetString("logs", "logs")
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
	if len(extraFields) > 0 {
		if fields == nil {
			fields = make(logrus.Fields)
		}
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
