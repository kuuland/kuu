package kuu

import (
	"fmt"
	"log"
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

// PRINT
func PRINT(args ...interface{}) {
	format, a := split(args...)
	Logger.Printf(format, a...)
}

// DEBUG
func DEBUG(args ...interface{}) {
	format, a := split(args...)
	Logger.Debugf(format, a...)
}

// WARN
func WARN(args ...interface{}) {
	format, a := split(args...)
	Logger.Warnf(format, a...)
}

// INFO
func INFO(args ...interface{}) {
	format, a := split(args...)
	Logger.Infof(format, a...)
}

// ERROR
func ERROR(args ...interface{}) {
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
	switch args[0].(type) {
	case error:
		args[0] = args[0].(error).Error()
	}
	format, a := split(args...)
	Logger.Fatalf(format, a...)
}

// PANIC
func PANIC(args ...interface{}) {
	switch args[0].(type) {
	case error:
		args[0] = args[0].(error).Error()
	}
	format, a := split(args...)
	Logger.Panicf(format, a...)
}
