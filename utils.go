package main

import (
	"io"
	"log"
	"os"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *log.Logger

func InitLogger(logDir string) {
	// 设置北京时间
	loc, _ := time.LoadLocation("Asia/Shanghai")
	time.Local = loc
	
	// 设置日志的默认时区格式
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		os.MkdirAll(logDir, 0755)
	}
	logPath := logDir + "/starring-qr-service-" + time.Now().Format("2006-01-02") + ".log"
	
	// 配置日志输出
	fileLogger := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    100,    // 单个日志文件最大100MB
		MaxBackups: 0,      // 不限制备份数量
		MaxAge:     7,      // 日志保留7天
		Compress:   false,  // 不压缩旧日志
	}
	
	// 同时输出到文件和控制台
	mw := io.MultiWriter(os.Stdout, fileLogger)
	Logger = log.New(mw, "", log.LstdFlags|log.Lshortfile)
	log.SetOutput(Logger.Writer())
}
