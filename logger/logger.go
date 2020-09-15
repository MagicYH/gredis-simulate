package logger

import (
	"context"
	"errors"
	"gredissimulate/config"
	"gredissimulate/helper"
	"log"
	"os"
	"path/filepath"
)

// var logger *log.Logger
// var file *os.File
// var (
// 	logger     *log.Logger
// 	file       *os.File
// 	logCh      chan logMsg
// 	logCtx     context.Context
// 	logCanncel context.CancelFunc
// )

var loggerInstance *Log

// Log : logger struct
type Log struct {
	logger     *log.Logger
	file       *os.File
	logCh      chan logMsg
	logCtx     context.Context
	logCanncel context.CancelFunc
}

// LogConf : Config infomation of logger
type LogConf struct {
	LogPath   string // Log path
	CacheSize int    // Channel cache size
}

type logMsg struct {
	msgType string        // Message type
	content []interface{} // Message content
}

// NewServer : create logger
func NewServer(conf LogConf) (*Log, error) {
	if nil != loggerInstance {
		return loggerInstance, nil
	}

	logPath := conf.LogPath
	if "" == logPath {
		logPath = config.GetAppPath() + "/log/run.log"
	}

	logDir := filepath.Dir(logPath)
	ret, err := helper.PathExists(logDir)
	if (false == ret) && (nil == err) {
		err := os.Mkdir(logDir, os.ModeDir|os.ModePerm)
		if nil != err {
			return nil, errors.New("Create log dir fail")
		}
	}

	file, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, errors.New("Fail to open log file " + err.Error())
	}

	instance := log.New(file, "", log.LstdFlags|log.Lshortfile)
	log.Println("Create log success")

	logCh := make(chan logMsg, conf.CacheSize)

	loggerInstance = &Log{
		logger: instance,
		file:   file,
		logCh:  logCh,
	}

	return loggerInstance, nil
}

// Start : Start log service
func (logger *Log) Start(ctx context.Context) error {
	logCtx, logCanncel := context.WithCancel(ctx)
	logger.logCtx = logCtx
	logger.logCanncel = logCanncel

	for {
		select {
		case msg := <-logger.logCh:
			logger.doLog(msg)

		case <-logCtx.Done():
			return nil
		}
	}
}

// Close : Close logger
func (logger *Log) Close() error {
	logger.logCanncel()
	logger.file.Close()
	return nil
}

// LogInfo : write data to log
func LogInfo(content ...interface{}) {
	if nil == loggerInstance {
		return
	}

	msg := logMsg{
		msgType: "info",
		content: content,
	}

	loggerInstance.logCh <- msg
}

// LogError : write data to log level error
func LogError(content ...interface{}) {
	if nil == loggerInstance {
		return
	}

	msg := logMsg{
		msgType: "error",
		content: content,
	}

	loggerInstance.logCh <- msg
}

func (logger *Log) doLog(msg logMsg) {
	switch msg.msgType {
	case "info":
		logger.logger.Println(msg.content...)
	case "error":
		logger.logger.Fatalln(msg.content...)
	default:

	}
}
