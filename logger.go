package kuu

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// log 日志实例
var log = logrus.New()

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	path := Join("kuu.", time.Now().Format("2006-01-02"), ".log")

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.Out = file
	} else {
		log.Info("Failed to log to file, using default stderr")
	}
}

func splitArgs(args ...interface{}) (string, []interface{}) {
	format := args[0].(string)
	var a []interface{}
	if len(args) > 1 {
		a = a[1:len(args)]
	}
	return format, a
}

// Debug Logger.Debug别名
func Debug(args ...interface{}) {
	format, a := splitArgs(args...)
	log.Debugf(format, a...)
}

// Info Logger.Info别名
func Info(args ...interface{}) {
	format, a := splitArgs(args...)
	log.Infof(format, a...)
}

// Print Logger.Print别名
func Print(args ...interface{}) {
	format, a := splitArgs(args...)
	log.Printf(format, a...)
}

// Warn Logger.Warn别名
func Warn(args ...interface{}) {
	format, a := splitArgs(args...)
	log.Warnf(format, a...)
}

// Error Logger.Error别名
func Error(args ...interface{}) {
	format, a := splitArgs(args...)
	log.Errorf(format, a...)
}

// Fatal Logger.Fatal别名
func Fatal(args ...interface{}) {
	format, a := splitArgs(args...)
	log.Fatalf(format, a...)
}

// Panic Logger.Panic别名
func Panic(args ...interface{}) {
	format, a := splitArgs(args...)
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
