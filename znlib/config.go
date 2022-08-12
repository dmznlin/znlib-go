/*Package znlib ***************************************************************
作者: dmzn@163.com 2022-05-30 13:45:17
描述: 配置lib库

描述:
1.依据配置文件初始化各单元
2.依据依赖先后顺序初始化各单元
******************************************************************************/
package znlib

import (
	iniFile "github.com/go-ini/ini"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

/*init 2022-05-30 13:47:33
  描述: 根据先后依赖调用各源文件初始化
*/
func init() {
	//默认配置: -------------------------------------------------------------------
	initApp()
	//application.go

	cfg := struct {
		logger    bool
		dbmanager bool
		snowflake bool
		redis     bool
	}{
		logger:    true,
		dbmanager: false,
		snowflake: true,
		redis:     false,
	}

	load_logConfig(nil, nil)
	load_redisConfig(nil, nil)
	load_snowflakeConfig(nil, nil)

	//外部配置: -------------------------------------------------------------------
	if FileExists(Application.ConfigFile, false) {
		ini, err := iniFile.Load(Application.ConfigFile)
		if err == nil {
			strBool := []string{"true", "false"}
			//bool array

			sec := ini.Section("logger")
			cfg.logger = sec.Key("enable").In("true", strBool) == "true"
			load_logConfig(ini, sec)

			sec = ini.Section("dbmanager")
			cfg.dbmanager = sec.Key("enable").In("true", strBool) == "true"

			sec = ini.Section("snowflake")
			cfg.snowflake = sec.Key("enable").In("true", strBool) == "true"
			load_snowflakeConfig(ini, sec)

			sec = ini.Section("redis")
			cfg.redis = sec.Key("enable").In("true", strBool) == "true"
			load_redisConfig(ini, sec)
		}
	}

	//启用配置: -------------------------------------------------------------------
	if cfg.logger {
		initLogger()
		//logger.go
	}

	if cfg.snowflake {
		init_snowflake()
		//idgen.go
	}

	if cfg.dbmanager {
		db_init()
		//dbhelper.go
	}

	if cfg.redis {
		init_redis()
		//redis.go
	}
}

/*load_logConfig 2022-08-11 19:22:24
  参数: ini,配置文件对象
  参数: sec,日志配置小节
  描述: 载入日志外部配置
*/
func load_logConfig(ini *iniFile.File, sec *iniFile.Section) {
	if sec == nil {
		logConfig.filePath = Application.LogPath
		return
	}

	val := StrTrim(sec.Key("filePath").String())
	if val != "" {
		if StrPos(val, "$path") < 0 {
			logConfig.filePath = val
		} else {
			logConfig.filePath = StrReplace(val, Application.ExePath, "$path\\", "$path/", "$path")
			//替换路径中的变量
		}

		logConfig.filePath = FixPath(logConfig.filePath)
		//添加路径分隔符
	}

	val = StrTrim(sec.Key("filename").String())
	if val != "" {
		logConfig.fileName = val
	}

	levels := []string{"trace", "debug", "info", "warning", "error", "fatal", "panic"}
	val = sec.Key("loglevel").In("info", levels)
	logConfig.logLevel, _ = logrus.ParseLevel(val)

	days := sec.Key("max_age").MustInt(30)
	logConfig.maxAge = time.Duration(days) * 24 * time.Hour
	//以天计时
}

/*load_redisConfig 2022-08-11 19:40:53
  参数: ini,配置文件对象
  参数: sec,snowflake配置小节
  参数: def,载入默认
  描述: 载入redis外部配置
*/
func load_snowflakeConfig(ini *iniFile.File, sec *iniFile.Section) {
	if sec == nil {
		return
	}

	snowflakeConfig.workerID = sec.Key("workerID").MustInt64(1)
	snowflakeConfig.datacenterID = sec.Key("dataCenterID").MustInt64(0)
}

/*load_redisConfig 2022-08-11 19:40:53
  参数: ini,配置文件对象
  参数: sec,redis配置小节
  参数: def,载入默认
  描述: 载入redis外部配置
*/
func load_redisConfig(ini *iniFile.File, sec *iniFile.Section) {
	if sec == nil {
		return
	}

	var str string
	redisConfig.cluster = sec.Key("cluster").String() == "true"
	str = StrTrim(sec.Key("server").String())

	if str != "" {
		hosts := strings.Split(str, ",")
		redisConfig.servers = append(redisConfig.servers, hosts...)
	}

	str = sec.Key("password").String()
	if str != "" {
		buf, err := NewEncrypter(EncryptDES_ECB, []byte(DefaultEncryptKey)).Decrypt([]byte(str), true)
		if err == nil {
			redisConfig.password = string(buf)
		} else {
			Error("znlib.load_redisConfig: " + err.Error())
		}
	}

	var val int
	val = sec.Key("poolSize").MustInt(0)
	if val != 0 {
		redisConfig.poolSize = val
	}

	val = sec.Key("dialTimeout").MustInt(0)
	if val != 0 {
		redisConfig.dialTimeout = time.Duration(val) * time.Second
	}

	val = sec.Key("readTimeout").MustInt(0)
	if val != 0 {
		redisConfig.readTimeout = time.Duration(val) * time.Second
	}

	val = sec.Key("writeTimeout").MustInt(0)
	if val != 0 {
		redisConfig.writeTimeout = time.Duration(val) * time.Second
	}

	val = sec.Key("poolTimeout").MustInt(0)
	if val != 0 {
		redisConfig.poolTimeout = time.Duration(val) * time.Second
	}
}
