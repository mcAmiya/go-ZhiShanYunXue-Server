package util

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

/*
此内容修改自CSDN
*/

// 颜色常量定义
const (
	Red    = 31
	Yellow = 33
	Blue   = 36
	Gray   = 37
)

// LogFormatter 结构体实现自定义日志格式化器
type LogFormatter struct{}

// Format 实现logrus.Formatter接口
func (lf *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var levelColor int
	switch entry.Level {
	case logrus.DebugLevel, logrus.TraceLevel:
		levelColor = Gray
	case logrus.WarnLevel:
		levelColor = Yellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = Red
	default:
		levelColor = Blue
	}

	// 创建或复用Buffer
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	// 自定义日期格式
	timestamp := entry.Time.Format("2006-01-02 15:04:05")
	filePath := ""
	funcName := ""

	// 如果开启了调用者追踪，添加文件名和函数名信息
	if entry.HasCaller() {
		funcName = filepath.Base(entry.Caller.Function)
		filePath = fmt.Sprintf("%s:%d", filepath.Base(entry.Caller.File), entry.Caller.Line)
		_, err := fmt.Fprintf(b, "[%s] \x1b[%dm[%s]\x1b[0m [%s] [%s] %s\n", timestamp, levelColor, entry.Level, filePath, funcName, entry.Message)
		if err != nil {
			return nil, err
		}
	} else {
		_, err := fmt.Fprintf(b, "[%s] \x1b[%dm[%s]\x1b[0m %s\n", timestamp, levelColor, entry.Level, entry.Message)
		if err != nil {
			return nil, err
		}
	}

	return b.Bytes(), nil
}

// NewLogger 创建并配置好自定义日志实例
func NewLogger() (*logrus.Logger, error) {
	logger := logrus.New()

	// 设置输出到标准输出（默认也是这样，这里仅作演示）
	logger.SetOutput(os.Stdout)

	// 设置自定义日志格式化器
	logger.SetFormatter(&LogFormatter{})

	// 开启调用者追踪
	logger.SetReportCaller(true)

	// 设置日志级别
	logger.SetLevel(logrus.DebugLevel)

	return logger, nil
}
