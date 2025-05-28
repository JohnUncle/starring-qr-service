package main

import (
	"log"
	"os"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *log.Logger

func InitLogger(logDir string) {
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		os.MkdirAll(logDir, 0755)
	}
	logPath := logDir + "/starring-qr-service-" + time.Now().Format("2006-01-02") + ".log"
	Logger = log.New(&lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    100, // 单个日志文件最大100MB
		MaxBackups: 0,   // 不限制备份数量
		MaxAge:     7,   // 日志保留7天
		Compress:   false,
	}, "", log.LstdFlags|log.Lshortfile)
	log.SetOutput(Logger.Writer())
}

func InfoLog(msg string, args ...interface{}) {
	Logger.Printf("[INFO] "+msg, args...)
}

func ErrorLog(msg string, args ...interface{}) {
	Logger.Printf("[ERROR] "+msg, args...)
}
