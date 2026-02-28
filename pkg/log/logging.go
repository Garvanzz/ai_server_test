package log

import (
	"fmt"
	"github.com/charmbracelet/log"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path/filepath"
	"strings"
	"xfx/pkg/env"
)

var logger *Logger

var _ io.Writer = (*lumberjack.Logger)(nil)

type Logger struct {
	*log.Logger
	conf       *env.Log
	fileWriter *lumberjack.Logger
}

func Init(conf *env.Log) {
	if conf.MaxSize == 0 {
		conf.MaxSize = 100
	}
	if conf.MaxBackups == 0 {
		conf.MaxBackups = 10
	}
	if conf.MaxAge == 0 {
		conf.MaxAge = 30
	}

	logger = &Logger{
		conf: conf,
	}

	output, fileWriter := logger.getOutput()
	logger.fileWriter = fileWriter
	logger.Logger = log.NewWithOptions(output, logger.buildOptions())
}

func DefaultInit() {
	Init(&env.Log{
		WriteFile: false,
		Level:     "DEBUG",
	})
}

func (l *Logger) getOutput() (io.Writer, *lumberjack.Logger) {
	if !l.conf.WriteFile {
		return os.Stdout, nil
	}

	if l.conf.FilePath == "" {
		return os.Stdout, nil
	} else {
		dir := filepath.Dir(l.conf.FilePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("创建日志目录失败: %v", err)
			return os.Stdout, nil
		}

		fileWriter := &lumberjack.Logger{
			Filename:   l.conf.FilePath,
			MaxSize:    l.conf.MaxSize,    // MB
			MaxBackups: l.conf.MaxBackups, // 最大备份数
			MaxAge:     l.conf.MaxAge,     // 天数
			Compress:   l.conf.Compress,   // 是否压缩
		}

		var writers []io.Writer
		writers = append(writers, os.Stdout, fileWriter)

		return io.MultiWriter(writers...), fileWriter
	}
}

func (l *Logger) buildOptions() log.Options {
	options := log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      "2006-01-02 15:04:05",
		Prefix:          l.conf.Prefix,
		CallerOffset:    1,
	}

	l.conf.Level = strings.ToUpper(l.conf.Level)

	switch l.conf.Level {
	case "DEBUG":
		options.Level = log.DebugLevel
	case "INFO":
		options.Level = log.InfoLevel
	case "WARN":
		options.Level = log.WarnLevel
	case "ERROR":
		options.Level = log.ErrorLevel
	case "FATAL":
		options.Level = log.FatalLevel
	default:
		options.Level = log.InfoLevel
	}

	switch l.conf.Format {
	case "json":
		options.Formatter = log.JSONFormatter
	case "logfmt":
		options.Formatter = log.LogfmtFormatter
	default:
		options.Formatter = log.TextFormatter
	}

	return options
}

// GetLogFileSize 获取当前日志文件大小
func (l *Logger) GetLogFileSize() int64 {
	if l.fileWriter != nil {
		if info, err := os.Stat(l.fileWriter.Filename); err == nil {
			return info.Size()
		}
	}
	return 0
}

func Printf(format string, args ...interface{}) {
	logger.Printf(format, args...)
}

func Debug(format string, v ...interface{}) {
	logger.Debugf(format, v...)
}

func Info(format string, v ...interface{}) {
	logger.Infof(format, v...)
}

func Warn(format string, v ...interface{}) {
	logger.Warnf(format, v...)
}

func Error(format string, v ...interface{}) {
	logger.Errorf(format, v...)
}

func Fatal(format string, v ...interface{}) {
	logger.Fatalf(format, v...)
}
