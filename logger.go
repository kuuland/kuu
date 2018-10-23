package kuu

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	log        = logrus.New()
	dateFormat = time.Now().Format("2006-01-02")
	dateFile   *os.File
	logsDir    = os.Getenv("KUU_LOGS")
)

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	setLogOut(Join("kuu.", dateFormat, ".log"))
	log.AddHook(new(outputHook))
}

func setLogOut(filePath string) {
	if logsDir == "" {
		config, err := LocalConfig()
		if err == nil && config["logsDir"] != nil {
			dir := config["logsDir"].(string)
			if !path.IsAbs(dir) {
				dir = path.Join(ROOT, dir)
			}
			logsDir = dir
		} else {
			logsDir = ROOT
		}
	}
	EnsureDir(logsDir)
	filePath = path.Join(logsDir, filePath)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.Out = file
		if dateFile != nil {
			dateFile.Close()
		}
		dateFile = file
	} else {
		log.Info("Failed to log to file, using default stderr")
	}
}

func dateCheck() {
	f := time.Now().Format("2006-01-02")
	if dateFormat != f {
		dateFormat = f
		setLogOut(Join("kuu.", dateFormat, ".log"))
	}
}

type outputHook struct{}

func (h *outputHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *outputHook) Fire(entry *logrus.Entry) error {
	dateCheck()
	fmt.Println(Join("[KUU-", strings.ToUpper(entry.Level.String()), time.Now().Format(" 2006-01-02 15:04:05] "), entry.Message))
	return nil
}

func split(args ...interface{}) (string, []interface{}) {
	format := args[0].(string)
	var a []interface{}
	if len(args) > 1 {
		a = a[1:len(args)]
	}
	return format, a
}

// Debug 输出Debug级别的日志
func Debug(args ...interface{}) {
	format, a := split(args...)
	log.Debugf(format, a...)
}

// Info 输出Info级别的日志
func Info(args ...interface{}) {
	format, a := split(args...)
	log.Infof(format, a...)
}

// Print 输出Print级别的日志
func Print(args ...interface{}) {
	format, a := split(args...)
	log.Printf(format, a...)
}

// Warn 输出Warn级别的日志
func Warn(args ...interface{}) {
	format, a := split(args...)
	log.Warnf(format, a...)
}

// Error 输出Error级别的日志
func Error(args ...interface{}) {
	format, a := split(args...)
	log.Errorf(format, a...)
}

// Fatal 输出Fatal级别的日志
func Fatal(args ...interface{}) {
	format, a := split(args...)
	log.Fatalf(format, a...)
}

// Panic 输出Panic级别的日志
func Panic(args ...interface{}) {
	format, a := split(args...)
	log.Panicf(format, a...)
}

// Debug 应用实例函数
func (k *Kuu) Debug(args ...interface{}) {
	Debug(args...)
}

// Info 应用实例函数
func (k *Kuu) Info(args ...interface{}) {
	Info(args...)
}

// Print 应用实例函数
func (k *Kuu) Print(args ...interface{}) {
	Print(args...)
}

// Warn 应用实例函数
func (k *Kuu) Warn(args ...interface{}) {
	Warn(args...)
}

// Error 应用实例函数
func (k *Kuu) Error(args ...interface{}) {
	Error(args...)
}

// Fatal 应用实例函数
func (k *Kuu) Fatal(args ...interface{}) {
	Fatal(args...)
}

// Panic 应用实例函数
func (k *Kuu) Panic(args ...interface{}) {
	Panic(args...)
}
