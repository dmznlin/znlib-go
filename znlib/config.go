// Package znlib
/******************************************************************************
作者: dmzn@163.com 2022-05-30 13:45:17
描述: 配置lib库

描述:
1.依据配置文件初始化各单元
2.依据依赖先后顺序初始化各单元
******************************************************************************/
package znlib

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/dmznlin/znlib-go/znlib/cast"
	iniFile "github.com/go-ini/ini"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// initLibUtils 初始化函数集
type initLibUtils = func()

var (
	// initBeforeUtil 初始化开始前执行操作
	initBeforeUtil initLibUtils

	// initLibOnce 确保一次初始化
	initLibOnce sync.Once
)

// InitLib 2022-08-16 20:49:11
/*
 描述: 由调用者执行初始化
*/
func InitLib(before, after initLibUtils) (result struct{}) {
	initBeforeUtil = before
	initLibOnce.Do(init_lib)

	if after != nil {
		after()
	}
	return struct{}{}
}

// init_lib 2022-05-30 13:47:33
/*
 描述: 根据先后依赖调用各源文件初始化
*/
func init_lib() {
	//默认配置: -------------------------------------------------------------------
	initApp()
	//application.go

	if initBeforeUtil != nil { //执行外部设定
		initBeforeUtil()
	}

	cfg := struct {
		dbmanager bool
		snowflake bool
		redis     bool
		mqtt      bool
	}{
		dbmanager: false,
		snowflake: false,
		redis:     false,
		mqtt:      false,
	}

	load_logConfig(nil, nil)
	load_redisConfig(nil, nil)
	load_snowflakeConfig(nil, nil)
	load_mqttConfig(nil, nil)

	//外部配置: -------------------------------------------------------------------
	if FileExists(Application.ConfigFile, false) {
		ini, err := iniFile.Load(Application.ConfigFile)
		if err != nil {
			ErrorCaller(err, "znlib.init_lib")
			return
		}

		strBool := []string{"true", "false"}
		//bool array

		sec := ini.Section("logger")
		load_logConfig(ini, sec)
		init_logger() //logger.go

		sec = ini.Section("dbmanager")
		cfg.dbmanager = sec.Key("enable").In("true", strBool) == "true"

		sec = ini.Section("snowflake")
		cfg.snowflake = sec.Key("enable").In("true", strBool) == "true"
		load_snowflakeConfig(ini, sec)

		sec = ini.Section("redis")
		cfg.redis = sec.Key("enable").In("true", strBool) == "true"
		load_redisConfig(ini, sec)

		sec = ini.Section("mqtt")
		cfg.mqtt = sec.Key("enable").In("true", strBool) == "true"

		load_mqttConfig(ini, sec)
		sec = ini.Section("mqttSSH")
		loadMqttSSHConfig(ini, sec)
	} else {
		init_logger()
		//logger.go
	}

	//启用配置: -------------------------------------------------------------------
	if cfg.snowflake {
		init_snowflake()
		//idgen.go
	}

	if cfg.dbmanager {
		init_db()
		//dbhelper.go
	}

	if cfg.redis {
		init_redis()
		//redis.go
	}

	if cfg.mqtt {
		init_mqtt()
		//mqtt.go
		init_mqttSSH()
		//mqttssh.go
	}
}

// load_logConfig 2022-08-11 19:22:24
/*
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
			logConfig.filePath = FixPathVar(val)
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

	if Application.IsWindows { //win下支持彩色
		logConfig.colors = sec.Key("colorful").String() == "true"
	}

	days := sec.Key("max_age").MustInt(30)
	logConfig.maxAge = time.Duration(days) * 24 * time.Hour
	//以天计时
}

// load_redisConfig 2022-08-11 19:40:53
/*
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

// load_redisConfig 2022-08-11 19:40:53
/*
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
			ErrorCaller(err, "znlib.load_redisConfig")
			return
		}
	}

	var val int
	val = sec.Key("poolSize").MustInt(0)
	if val != 0 {
		redisConfig.poolSize = val
	}

	val = sec.Key("defaultDB").MustInt(0)
	if val != 0 {
		redisConfig.defaultDB = val
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

// load_mqttConfig 2024-01-09 16:54:33
/*
 参数: ini,配置文件对象
 参数: sec,mqtt配置小节
 描述: 载入mqtt外部配置
*/
func load_mqttConfig(ini *iniFile.File, sec *iniFile.Section) {
	if sec == nil {
		return
	}

	caller := "znlib.load_mqttConfig"
	var val int
	var str string

	str = sec.Key("broker").String()
	brokers := strings.Split(str, ",")

	for _, v := range brokers {
		Mqtt.Options.AddBroker(v)
		//多服务器支持
	}

	str = StrTrim(sec.Key("clientID").String())
	if str != "" {
		idTag := "auto^"
		idx := strings.Index(str, idTag)
		if idx >= 0 {
			sufLen, err := strconv.Atoi(str[idx+len(idTag):]) //auto^5
			if err != nil {
				ErrorCaller(err, caller+".autoClientID")
				return
			}

			str = str[0:idx]       //前缀
			idLen := 23 - len(str) //mqtt id长度限制
			if sufLen > idLen {    //取最大可用长度
				sufLen = idLen
			}

			suffix := SerialID.TimeID(true) //后缀
			idLen = len(suffix)
			if idLen > sufLen {
				idLen = idLen - sufLen
			} else {
				sufLen = idLen
				idLen = 0
			}
			str = str + suffix[idLen:idLen+sufLen]
		}
		Mqtt.Options.SetClientID(str)
	}

	str = StrTrim(sec.Key("userName").String())
	if str != "" {
		Mqtt.Options.SetUsername(str)
		//user-name
	}

	str = StrTrim(sec.Key("password").String())
	if str != "" {
		buf, err := NewEncrypter(EncryptDES_ECB, []byte(DefaultEncryptKey)).Decrypt([]byte(str), true)
		if err != nil {
			ErrorCaller(err, caller)
			return
		}

		Mqtt.Options.SetPassword(string(buf))
		//user-password
	}

	str = StrTrim(sec.Key("encryptKey").String())
	if str != "" {
		buf, err := NewEncrypter(EncryptDES_ECB, []byte(DefaultEncryptKey)).Decrypt([]byte(str), true)
		if err != nil {
			ErrorCaller(err, caller)
			return
		}

		MqttUtils.msgKey = string(buf)
		//消息加密密钥
	}

	str = StrTrim(sec.Key("verifyMsg").String())
	if str != "" {
		MqttUtils.msgVerify = str == "true"
		//启用消息有效性验证
	}

	val = sec.Key("workerNum").MustInt(0)
	if val > 0 {
		MqttUtils.workerNum = val
		//消息工作对象个数
	}

	val = sec.Key("delayWarn").MustInt(1)
	if str != "" {
		MqttUtils.msgDelay = time.Duration(val) * time.Second
	}

	str = FixPathVar(sec.Key("fileCA").String())
	if FileExists(str, false) { //ca exists
		rootCA, err := os.ReadFile(str)
		if err != nil {
			ErrorCaller(err, caller)
			return
		}

		cp := x509.NewCertPool()
		if !cp.AppendCertsFromPEM(rootCA) {
			ErrorCaller("could not add root crt", caller)
			return
		}

		str = FixPathVar(sec.Key("fileCRT").String())
		key := FixPathVar(sec.Key("fileKey").String())
		cert, err := tls.LoadX509KeyPair(str, key)
		if err != nil {
			ErrorCaller(err, caller)
		}

		Mqtt.Options.SetTLSConfig(&tls.Config{
			RootCAs:            cp,
			ClientAuth:         tls.NoClientCert,
			ClientCAs:          nil,
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{cert},
		})
	}

	getMqttTopic(sec, "subTopic", true, caller)
	//订阅主题列表
	getMqttTopic(sec, "publish", false, caller)
	//发布主题列表
}

// getMqttTopic 2024-02-01 19:49:37
/*
 参数: sec,ini配置小节
 参数: key,ini键
 参数: isSub,是否订阅主题
 描述: 读取sec.key中的主题配置
*/
func getMqttTopic(sec *iniFile.Section, key string, isSub bool, caller string) (topics map[string]mqttQos) {
	var str string
	var qos mqttQos
	str = StrTrim(sec.Key(key).String())
	if str == "" {
		ErrorCaller("invalid topics key > "+key, caller)
		return nil
	}

	list := strings.Split(str, ",")
	if len(list) > 0 {
		topics = make(map[string]mqttQos, len(list))
	}

	for _, v := range list { //topics^qos
		v = strings.ReplaceAll(v, "*", "#")
		//通配符
		pos := strings.Index(v, "^")
		if pos < 1 {
			ErrorCaller("invalid topics format > "+v, caller)
			continue
		}

		str = v[:pos]
		qos = byte(cast.ToInt8(v[pos+1:]))
		topics[str] = qos //add result

		topic := strings.ReplaceAll(str, "$id", Mqtt.Options.ClientID)
		//使用id配置
		if isSub {
			Mqtt.subTopics[topic] = qos
		} else {
			Mqtt.pubTopics[topic] = qos

		}
	}

	return topics
}

// loadMqttSSHConfig 2024-01-09 16:54:33
/*
 参数: ini,配置文件对象
 参数: sec,mqttSSH配置小节
 描述: 载入mqttSSH外部配置
*/
func loadMqttSSHConfig(ini *iniFile.File, sec *iniFile.Section) {
	if sec == nil {
		return
	}

	caller := "znlib.loadMqttSSHConfig"
	var str string
	MqttSSH.enabled = StrTrim(sec.Key("enable").String()) == "true"

	str = StrTrim(sec.Key("host").String())
	if str != "" {
		MqttSSH.host = str
	}

	str = StrTrim(sec.Key("user").String())
	if str != "" {
		MqttSSH.user = str
	}

	str = StrTrim(sec.Key("password").String())
	if str != "" {
		buf, err := NewEncrypter(EncryptDES_ECB, []byte(DefaultEncryptKey)).Decrypt([]byte(str), true)
		if err != nil {
			ErrorCaller(err, caller)
			return
		}

		MqttSSH.password = string(buf)
		//用户密码
	}

	topics := getMqttTopic(sec, "channel", true, caller)
	if topics != nil {
		for k, v := range topics { //获取通道列表
			MqttSSH.channel[k] = v
		}
	}

	str = StrTrim(sec.Key("connTimeout").String())
	if str != "" {
		MqttSSH.connTimeout = cast.ToDuration(str)
	}

	str = StrTrim(sec.Key("exitTimeout").String())
	if str != "" {
		MqttSSH.exitTimeout = cast.ToDuration(str)
	}

	str = StrTrim(sec.Key("command").String())
	if str != "" {
		MqttSSH.sshCmd = cast.ToUint8(str)
	}
}
