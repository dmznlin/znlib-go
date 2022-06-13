/*Package znlib ***************************************************************
作者: dmzn@163.com 2022-05-30 13:46:20
描述: 提供写入文件的日志
******************************************************************************/
package znlib

import (
	"bufio"
	iniFile "github.com/go-ini/ini"
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"os"
	"sync"
	"time"
)

const (
	logInfo int8 = iota
	logWarn
	logEror
)

type LogConfig struct {
	FilePath string        `ini:"filePath"`
	FileName string        `ini:"filename"`
	LogLevel logrus.Level  `ini:"loglevel"`
	MaxAge   time.Duration `ini:"max_age"`
}

//logcfg 默认日志配置参数
var logcfg *LogConfig = nil

//logcfg_init 初始化日志配置
var logcfg_init sync.Once

/*initLogConfig 2022-06-13 15:26:57
  描述: 初始化默认日志配置
*/
func initLogConfig() {
	logcfg_init.Do(func() {
		logcfg = NewLogConfig()
		logcfg.LoadConfig()
	})
}

/*NewLogConfig 2022-06-13 15:27:40
  描述: 日志配置
*/
func NewLogConfig() *LogConfig {
	return &LogConfig{
		FilePath: Application.LogPath,
		FileName: "sys.log",
		LogLevel: logrus.InfoLevel,
		MaxAge:   24 * time.Hour,
	}
}

/*LoadConfig 2022-06-05 14:49:34
  参数: cFile,日志配置文件
  对象: cfg,日志配置
  描述: 从confFile中载入cfg的配置
*/
func (cfg *LogConfig) LoadConfig(cFile ...string) {
	var cf string
	if cFile == nil {
		cf = Application.ConfigFile
	} else {
		cf = cFile[0]
	}

	if !FileExists(cf, false) {
		return
	}

	ini, err := iniFile.Load(cf)
	if err == nil {
		sec := ini.Section("logger")
		val := StrTrim(sec.Key("filePath").String())
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

		val = StrTrim(sec.Key("filename").String())
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

//--------------------------------------------------------------------------------

//Logger 全局日志对象
var Logger *logrus.Logger

//LogFields 日志附加字段
type LogFields = logrus.Fields

/*addLog 2022-05-30 13:47:50
  参数: logType,日志类型
  参数: log,日志内容
  参数: fields,附加字段
  描述: 新增一条类型为logType的日志
*/
func addLog(logType int8, log string, fields ...LogFields) {
	if Logger == nil {
		WriteDefaultLog("msg:" + log)
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

/*Info 2022-05-30 13:48:29
  参数: info,日志内容
  参数: fields,附加字段
  描述: 新增一条info信息
*/
func Info(info string, fields ...LogFields) {
	addLog(logInfo, info, fields...)
}

/*Warn 2022-05-30 13:48:44
  参数: warn,日志内容
  参数: fields,附加字段
  描述: 新增一条警告信息
*/
func Warn(warn string, fields ...logrus.Fields) {
	addLog(logWarn, warn, fields...)
}

/*Error 2022-05-30 13:48:58
  参数: error,日志内容
  参数: fields,附加字段
  描述: 新增一条错误信息
*/
func Error(error string, fields ...logrus.Fields) {
	addLog(logEror, error, fields...)
}

/*WriteDefaultLog 2022-06-05 14:46:04
  参数: data,日志数据
  描述: 将data写入默认日志文件
*/
func WriteDefaultLog(data string) {
	if len(data) == 0 {
		return
	}

	if logcfg == nil {
		initLogConfig()
	}

	hwnd, err := os.OpenFile(logcfg.FilePath+"log_def.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
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

func initLogger() {
	Logger = logrus.New()
	//new logger
	initLogConfig()

	if !FileExists(logcfg.FilePath, true) {
		MakeDir(logcfg.FilePath) //创建日志目录
	}

	logfile := logcfg.FilePath + logcfg.FileName
	opt := []rotatelogs.Option{
		// 设置最大保存时间(7天)
		rotatelogs.WithMaxAge(logcfg.MaxAge),
		// 设置日志切割时间间隔(1天)
		rotatelogs.WithRotationTime(24 * time.Hour),
	}

	if Application.IsLinux {
		// 生成软链，指向最新日志文件
		opt = append(opt, rotatelogs.WithLinkName(logfile))
	}

	writer, err := rotatelogs.New(logcfg.FilePath+logcfg.FileName+"%Y%m%d.log", opt...)
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

	/* You could set this to any `io.Writer` such as a file
	file, err := os.OpenFile(logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		WriteDefaultLog("znlib.os.OpenFile: " + err.Error())
		return
	}

	Logger.SetOutput(io.MultiWriter(file, os.Stdout))
	//设置双输出 */
	Logger.SetLevel(logcfg.LogLevel)

	Logger.SetFormatter(&logrus.TextFormatter{
		ForceQuote:      true,                //键值对加引号
		FullTimestamp:   true,                //完整时间戳
		TimestampFormat: LayoutDateTimeMilli, //时间格式
	})
}
