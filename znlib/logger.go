package znlib

/******************************************************************************
作者: dmzn@163.com 2022-05-09
描述: 提供写入文件的日志
******************************************************************************/
import (
	iniFile "github.com/go-ini/ini"
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"time"
)

const (
	logInfo = iota
	logWarn
	logEror
)

type logCfg struct {
	FilePath string        `ini:"filePath"`
	FileName string        `ini:"filename"`
	LogLevel logrus.Level  `ini:"loglevel"`
	MaxAge   time.Duration `ini:"max_age"`
}

//全局日志对象
var Logger *logrus.Logger

//日志附加字段
type LogFields = logrus.Fields

//-----------------------------------------------------------------------------
//Date: 2022-05-10
//Parm: 日志类型;日志内容;附加字段
//Desc: 新增一条类型为logType的日志
func addLog(logType int8, log string, fields ...LogFields) {
	if Logger == nil {
		logrus.Warn("znlib.Logger is nil(not init)")
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
}

//Date: 2022-05-10
//Parm: 日志内容;附加字段
//Desc: 新增一条info信息
func Info(info string, fields ...LogFields) {
	addLog(logInfo, info, fields...)
}

//Date: 2022-05-10
//Parm: 日志内容;附加字段
//Desc: 新增一条警告信息
func Warn(warn string, fields ...logrus.Fields) {
	addLog(logWarn, warn, fields...)
}

//Date: 2022-05-10
//Parm: 日志内容;附加字段
//Desc: 新增一条错误信息
func Error(error string, fields ...logrus.Fields) {
	addLog(logEror, error, fields...)
}

//-----------------------------------------------------------------------------
func initLogger() {
	Logger = logrus.New()
	//new logger

	cfg := logCfg{
		FilePath: Application.LogPath,
		FileName: "sys.log",
		LogLevel: logrus.InfoLevel,
		MaxAge:   30 * 24 * time.Hour,
	} //default config

	if FileExists(Application.ConfigFile, false) {
		ini, err := iniFile.Load(Application.ConfigFile)
		if err == nil {
			sec := ini.Section("logger")
			val := Trim(sec.Key("filePath").String())
			if val != "" {
				if StrPos(val, "$path") < 0 {
					cfg.FilePath = val
				} else {
					cfg.FilePath = StrReplace(val, Application.ExePath, "$path\\", "$path/", "$path")
					//替换路径中的变量
				}

				cfg.FilePath = FixPath(cfg.FilePath)
				//添加路径分隔符
			}

			val = Trim(sec.Key("filename").String())
			if val != "" {
				cfg.FileName = val
			}

			levels := []string{"trace", "debug", "info", "warning", "error", "fatal", "panic"}
			val = sec.Key("loglevel").In("info", levels)
			cfg.LogLevel, _ = logrus.ParseLevel(val)

			days := sec.Key("max_age").MustInt(30)
			cfg.MaxAge = time.Duration(days) * 24 * time.Hour
			//以天计时
		}
	}

	if !FileExists(cfg.FilePath, true) {
		MakeDir(cfg.FilePath) //创建日志目录
	}

	logfile := cfg.FilePath + cfg.FileName
	opt := []rotatelogs.Option{
		// 设置最大保存时间(7天)
		rotatelogs.WithMaxAge(cfg.MaxAge),
		// 设置日志切割时间间隔(1天)
		rotatelogs.WithRotationTime(24 * time.Hour),
	}

	if Application.IsLinux {
		// 生成软链，指向最新日志文件
		opt = append(opt, rotatelogs.WithLinkName(logfile))
	}

	writer, err := rotatelogs.New(cfg.FilePath+"%Y%m%d.log", opt...)
	writeMap := lfshook.WriterMap{
		logrus.InfoLevel:  writer,
		logrus.FatalLevel: writer,
		logrus.DebugLevel: writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
		logrus.PanicLevel: writer,
	}

	lfHook := lfshook.NewHook(writeMap, &logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000000",
	})
	Logger.AddHook(lfHook)

	// You could set this to any `io.Writer` such as a file
	file, err := os.OpenFile(logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return
	}

	Logger.SetOutput(io.MultiWriter(file, os.Stdout))
	//设置双输出
	Logger.SetLevel(cfg.LogLevel)

	Logger.SetFormatter(&logrus.TextFormatter{
		ForceQuote:      true,                         //键值对加引号
		FullTimestamp:   true,                         //完整时间戳
		TimestampFormat: "2006-01-02 15:04:05.000000", //时间格式
	})
}
