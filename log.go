package kuu

import (
	"github.com/sirupsen/logrus"
	"os"
)

func init() {
	logrus.SetOutput(os.Stdout)
}

var Logger = logrus.New()

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
	format, a := split(args...)
	Logger.Fatalf(format, a...)
}

// PANIC
func PANIC(args ...interface{}) {
	format, a := split(args...)
	Logger.Panicf(format, a...)
}
