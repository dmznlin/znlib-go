// Package znlib
/******************************************************************************
  作者: dmzn@163.com 2022-05-30 13:46:20
  描述: 提供写入文件的日志
******************************************************************************/
package znlib

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	rl "github.com/lestrrat/go-file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/shiena/ansicolor"
	"github.com/sirupsen/logrus"
)

type logType = int8

const (
	logInfo logType = iota
	logWarn
	logError
)

// LogFields 日志附加字段
type LogFields = logrus.Fields

var (
	// Logger 全局日志对象
	Logger *logrus.Logger = nil

	// LogTraceCaller 跟踪日志调用
	LogTraceCaller = false
)

// AddLog 2022-05-30 13:47:50
/*
 参数: logType,日志类型
 参数: log,日志内容
 参数: trace,是否跟踪调用
 参数: fields,附加字段
 描述: 新增一条类型为logType的日志
*/
func AddLog(logType logType, log interface{}, trace bool, fields ...LogFields) {
	if Logger == nil {
		WriteDefaultLog(fmt.Sprintf("msg: %v", log))
		return
	}

	if fields == nil { //无附加字段
		switch logType {
		case logInfo:
			Logger.Info(log)
		case logWarn:
			Logger.Warn(log)
		case logError:
			Logger.Error(log)
		}
	} else {
		all := make(LogFields, 10)
		for _, fs := range fields {
			for k, v := range fs {
				all[k] = v
			}
		}

		switch logType {
		case logInfo:
			Logger.WithFields(all).Info(log)
		case logWarn:
			Logger.WithFields(all).Warn(log)
		case logError:
			Logger.WithFields(all).Error(log)
		}
	}

	if trace { //附加调用路径跟踪
		var (
			idx = 2
			str strings.Builder
		)

		for {
			pc, file, line, ok := runtime.Caller(idx)
			if !ok {
				break
			}

			file = filepath.Base(file)
			if StrIn(file, "proc.go") {
				break
			}

			idx++
			fn := runtime.FuncForPC(pc)
			if fn != nil {
				str.WriteString(fmt.Sprintf(" %s,%d", file, line))
			}
		}

		if str.Len() > 0 {
			AddLog(logInfo, "trace:"+str.String(), false)
		}
	}
}

// Info 2022-05-30 13:48:29
/*
 参数: info,日志内容
 参数: fields,附加字段
 描述: 新增一条info信息
*/
func Info(info interface{}, fields ...LogFields) {
	AddLog(logInfo, info, LogTraceCaller, fields...)
}

// Warn 2022-05-30 13:48:44
/*
 参数: warn,日志内容
 参数: fields,附加字段
 描述: 新增一条警告信息
*/
func Warn(warn interface{}, fields ...logrus.Fields) {
	AddLog(logWarn, warn, LogTraceCaller, fields...)
}

// Error 2022-05-30 13:48:58
/*
 参数: error,日志内容
 参数: fields,附加字段
 描述: 新增一条错误信息
*/
func Error(error interface{}, fields ...logrus.Fields) {
	AddLog(logError, error, LogTraceCaller, fields...)
}

// ErrorCaller 2024-01-19 09:28:13
/*
 参数: error,日志内容
 参数: caller,调用者
 描述: 新增一条错误信息
*/
func ErrorCaller(error interface{}, caller string) {
	caller = StrTrim(caller)
	if caller == "" {
		AddLog(logError, error, LogTraceCaller)
	} else {
		AddLog(logError, error, LogTraceCaller, LogFields{"caller": caller})
	}
}

// WriteDefaultLog 2022-06-05 14:46:04
/*
 参数: data,日志数据
 描述: 将data写入默认日志文件
*/
func WriteDefaultLog(data string) {
	if len(data) == 0 {
		return
	}

	lf, err := os.OpenFile(GlobalConfig.Logger.FilePath+"log_def.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return
	}
	defer func(lf *os.File) {
		err = lf.Close()
		if err != nil {
			fmt.Println("znlib.WriteDefaultLog: ", err)
		}
	}(lf) //close file

	buf := bufio.NewWriterSize(lf, 200)
	defer func(buf *bufio.Writer) {
		err = buf.Flush()
		if err != nil {
			fmt.Println("znlib.WriteDefaultLog: ", err)
		}
	}(buf) //将缓冲中的数据写入

	if StrCopyRight(data, 1) != "\n" {
		data = data + "\n"
	}
	_, _ = buf.WriteString(DateTime2Str(time.Now(), LayoutDateTimeMilli) + "\t" + data)
}

//-----------------------------------------------------------------------------

func initLogger(cfg *LoggerConfig) {
	Logger = logrus.New()
	//new logger

	if !FileExists(cfg.FilePath, true) {
		MakeDir(cfg.FilePath) //创建日志目录
	}

	logfile := cfg.FilePath + cfg.FileName
	opt := []rl.Option{
		// 设置最大保存时间(7天)
		rl.WithMaxAge(cfg.MaxAge * 24 * time.Hour),
		// 设置日志切割时间间隔(1天)
		rl.WithRotationTime(24 * time.Hour),
	}

	if Application.IsLinux {
		// 生成软链，指向最新日志文件
		opt = append(opt, rl.WithLinkName(logfile))
	}

	writer, err := rl.New(cfg.FilePath+cfg.FileName+"%Y%m%d.log", opt...)
	if err != nil {
		WriteDefaultLog("znlib.rl.New: " + err.Error())
		return
	}

	writeMap := lfshook.WriterMap{
		logrus.InfoLevel:  writer,
		logrus.FatalLevel: writer,
		logrus.DebugLevel: writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
		logrus.PanicLevel: writer,
	}

	lfHook := lfshook.NewHook(writeMap, &logrus.TextFormatter{
		ForceQuote:      true,                //键值对加引号
		FullTimestamp:   true,                //完整时间戳
		TimestampFormat: LayoutDateTimeMilli, //时间格式
	})
	Logger.AddHook(lfHook)

	var nFormatter = logrus.TextFormatter{
		ForceQuote:      true,                //键值对加引号
		FullTimestamp:   true,                //完整时间戳
		TimestampFormat: LayoutDateTimeMilli, //时间格式
		ForceColors:     false,
	}

	if cfg.Colorful {
		nFormatter.ForceColors = true
		// then wrap the log output with it
		Logger.SetOutput(ansicolor.NewAnsiColorWriter(os.Stdout))
	}

	Logger.SetFormatter(&nFormatter)
	//输出格式化
	Logger.SetLevel(cfg.LogLevel)
	//输出级别控制
}
