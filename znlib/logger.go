// Package znlib
/******************************************************************************
  作者: dmzn@163.com 2022-05-30 13:46:20
  描述: 提供写入文件的日志
******************************************************************************/
package znlib

import (
	"bufio"
	"fmt"
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/shiena/ansicolor"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type logType = int8

const (
	logInfo logType = iota
	logWarn
	logEror
)

// Logger 全局日志对象
var Logger *logrus.Logger = nil

// LogTraceCaller 跟踪日志调用
var LogTraceCaller = false

// LogFields 日志附加字段
type LogFields = logrus.Fields

// logConfig 默认日志配置参数
var logConfig = struct {
	filePath string        //日志目录
	fileName string        //日志文件名
	logLevel logrus.Level  //日志级别
	maxAge   time.Duration //日志保存天数
	colors   bool          //使用彩色终端
}{
	fileName: "sys.log",
	logLevel: logrus.InfoLevel,
	maxAge:   24 * time.Hour,
}

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
		case logEror:
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
		case logEror:
			Logger.WithFields(all).Error(log)
		}
	}

	if trace { //附加调用路径跟踪
		var (
			idx = 2
			str strings.Builder
		)

		for true {
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
	AddLog(logEror, error, LogTraceCaller, fields...)
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
		AddLog(logEror, error, LogTraceCaller)
	} else {
		AddLog(logEror, error, LogTraceCaller, LogFields{"caller": caller})
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

	hwnd, err := os.OpenFile(logConfig.filePath+"log_def.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return
	}
	defer hwnd.Close() //close file

	buf := bufio.NewWriterSize(hwnd, 200)
	defer buf.Flush() //将缓冲中的数据写入

	buf.WriteString(DateTime2Str(time.Now(), LayoutDateTimeMilli) + string(KeyTab) + data)
	if StrCopyRight(data, 1) != string(KeyEnter) {
		buf.WriteString(string(KeyEnter))
	}
}

//-----------------------------------------------------------------------------

func init_logger() {
	Logger = logrus.New()
	//new logger

	if !FileExists(logConfig.filePath, true) {
		MakeDir(logConfig.filePath) //创建日志目录
	}

	logfile := logConfig.filePath + logConfig.fileName
	opt := []rotatelogs.Option{
		// 设置最大保存时间(7天)
		rotatelogs.WithMaxAge(logConfig.maxAge),
		// 设置日志切割时间间隔(1天)
		rotatelogs.WithRotationTime(24 * time.Hour),
	}

	if Application.IsLinux {
		// 生成软链，指向最新日志文件
		opt = append(opt, rotatelogs.WithLinkName(logfile))
	}

	writer, err := rotatelogs.New(logConfig.filePath+logConfig.fileName+"%Y%m%d.log", opt...)
	if err != nil {
		WriteDefaultLog("znlib.rotatelogs.New: " + err.Error())
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

	if logConfig.colors {
		nFormatter.ForceColors = true
		// then wrap the log output with it
		Logger.SetOutput(ansicolor.NewAnsiColorWriter(os.Stdout))
	}

	Logger.SetFormatter(&nFormatter)
	//输出格式化
	Logger.SetLevel(logConfig.logLevel)
	//输出级别控制
}
