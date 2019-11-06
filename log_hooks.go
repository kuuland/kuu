package kuu

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

// LogDailyFileHook
type LogDailyFileHook struct{}

// Levels
func (h *LogDailyFileHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire
func (h *LogDailyFileHook) Fire(entry *logrus.Entry) error {
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

// LogBizHook
type LogBizHook struct{}

// Levels
func (h *LogBizHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire
func (h *LogBizHook) Fire(entry *logrus.Entry) error {
	now := fmt.Sprintf("kuu-%s.log", time.Now().Format("2006-01-02"))
	if now != DailyFileName || DailyFile == nil {
		DailyFileName = now
		changeLoggerOutput(now)
	}
	info := NewLog(LogTypeBiz)
	info.Level = entry.Level.String()
	info.ContentHuman = entry.Message
	info.ContentData = entry.Message
	info.Save2Cache()
	return nil
}
