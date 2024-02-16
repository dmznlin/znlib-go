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
	"fmt"
	"github.com/beevik/etree"
	"github.com/dmznlin/znlib-go/znlib/cast"
	"github.com/sirupsen/logrus"
	"os"
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

	// configStatus 由配置文件设置的状态
	configStatus = struct {
		dbManager bool
		snowFlake bool
		redis     bool
		mqtt      bool
	}{
		dbManager: false,
		snowFlake: false,
		redis:     false,
		mqtt:      false,
	}
)

// InitLib 2022-08-16 20:49:11
/*
 描述: 由调用者执行初始化
*/
func InitLib(before, after initLibUtils) (result struct{}) {
	initBeforeUtil = before
	initLibOnce.Do(initLibrary)

	if after != nil {
		after()
	}
	return struct{}{}
}

// initLibrary 2022-05-30 13:47:33
/*
 描述: 根据先后依赖调用各源文件初始化
*/
func initLibrary() {
	//默认配置: -------------------------------------------------------------------
	initApp()
	//application.go

	if initBeforeUtil != nil { //执行外部设定
		initBeforeUtil()
	}

	loadLogConfig(nil)
	loadDBConfig(nil, nil)
	loadRedisConfig(nil)
	loadSnowflakeConfig(nil)
	loadMqttConfig(nil)

	//外部配置: -------------------------------------------------------------------
	if FileExists(Application.ConfigFile, false) {
		xml := etree.NewDocument()
		if err := xml.ReadFromFile(Application.ConfigFile); err != nil {
			ErrorCaller(err, "znlib.config.initLibrary")
			return
		}

		root := xml.SelectElement("znlib")
		loadLogConfig(root.SelectElement("logger"))
		init_logger() //logger.go

		loadDBConfig(root.SelectElement("dbmanager"), DBManager)
		loadSnowflakeConfig(root.SelectElement("snowflake"))
		loadRedisConfig(root.SelectElement("redis"))

		loadMqttConfig(root.SelectElement("mqtt"))
		loadMqttSSHConfig(root.SelectElement("mqttSSH"))
	} else {
		init_logger()
		//logger.go
	}

	//启用配置: -------------------------------------------------------------------
	if configStatus.snowFlake {
		init_snowflake()
		//idgen.go
	}

	if configStatus.dbManager {
		init_db()
		//dbhelper.go
	}

	if configStatus.redis {
		init_redis()
		//redis.go
	}

	if configStatus.mqtt {
		init_mqtt()
		//mqtt.go
		init_mqttSSH()
		//mqttssh.go
	}
}

// loadLogConfig 2022-08-11 19:22:24
/*
 参数: root,log根节点
 描述: 载入日志外部配置
*/
func loadLogConfig(root *etree.Element) {
	if root == nil {
		logConfig.filePath = Application.LogPath
		return
	}

	val := StrTrim(root.SelectElement("filePath").Text())
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

	val = StrTrim(root.SelectElement("filename").Text())
	if val != "" {
		logConfig.fileName = val
	}

	levels := []string{"trace", "debug", "info", "warning", "error", "fatal", "panic"}
	val = root.SelectElement("loglevel").Text()
	if !StrIn(val, levels...) {
		val = "info"
	}

	logConfig.logLevel, _ = logrus.ParseLevel(val)
	Application.IsDebug = logConfig.logLevel == logrus.DebugLevel
	//全局debug开发

	if Application.IsWindows { //win下支持彩色
		logConfig.colors = root.SelectElement("colorful").Text() == "true"
	}

	days := cast.ToDuration(root.SelectElement("max_age").Text())
	logConfig.maxAge = days * 24 * time.Hour
	//以天计时
}

// loadRedisConfig 2022-08-11 19:40:53
/*
 参数: root,snowflake配置根节点
 描述: 载入redis外部配置
*/
func loadSnowflakeConfig(root *etree.Element) {
	if root == nil {
		return
	}

	configStatus.snowFlake = root.SelectAttr("enable").Value == StrTrue
	//status

	val, err := cast.ToInt64E(root.SelectElement("workerID").Text())
	if err != nil {
		snowflakeConfig.workerID = 1
	} else {
		snowflakeConfig.workerID = val
	}

	val, err = cast.ToInt64E(root.SelectElement("dataCenterID").Text())
	if err != nil {
		snowflakeConfig.datacenterID = 0
	} else {
		snowflakeConfig.datacenterID = val
	}
}

// loadRedisConfig 2022-08-11 19:40:53
/*
 参数: root,redis配置根节点
 描述: 载入redis外部配置
*/
func loadRedisConfig(root *etree.Element) {
	if root == nil {
		return
	}

	configStatus.redis = root.SelectAttr("enable").Value == StrTrue
	//status

	var str string
	redisConfig.cluster = root.SelectElement("cluster").Text() == StrTrue
	str = StrTrim(root.SelectElement("server").Text())

	if str != "" {
		hosts := strings.Split(str, ",")
		redisConfig.servers = append(redisConfig.servers, hosts...)
	}

	str = StrTrim(root.SelectElement("password").Text())
	if str != "" {
		buf, err := NewEncrypter(EncryptDES_ECB, []byte(DefaultEncryptKey)).Decrypt([]byte(str), true)
		if err == nil {
			redisConfig.password = string(buf)
		} else {
			ErrorCaller(err, "znlib.config.loadRedisConfig")
			return
		}
	}

	v1, e1 := cast.ToIntE(root.SelectElement("poolSize").Text())
	if e1 == nil && v1 > 0 {
		redisConfig.poolSize = v1
	}

	v1, e1 = cast.ToIntE(root.SelectElement("defaultDB").Text())
	if e1 == nil && v1 > 0 {
		redisConfig.defaultDB = v1
	}

	node := root.SelectElement("timeout")
	val, err := cast.ToDurationE(node.SelectElement("dial").Text())
	if err == nil && val > 0 {
		redisConfig.dialTimeout = val * time.Second
	}

	val, err = cast.ToDurationE(node.SelectElement("read").Text())
	if err == nil && val > 0 {
		redisConfig.readTimeout = val * time.Second
	}

	val, err = cast.ToDurationE(node.SelectElement("write").Text())
	if err == nil && val > 0 {
		redisConfig.writeTimeout = val * time.Second
	}

	val, err = cast.ToDurationE(node.SelectElement("pool").Text())
	if err == nil && val > 0 {
		redisConfig.poolTimeout = val * time.Second
	}
}

// loadDBConfig 2024-02-05 18:22:05
/*
 参数: root,db配置根节点
 描述: 载入数据库外部配置
*/
func loadDBConfig(root *etree.Element, dm *DbUtils) {
	if root == nil || dm == nil {
		return
	}

	configStatus.dbManager = root.SelectAttr("enable").Value == StrTrue
	//status

	caller := "znlib.config.loadDBConfig"
	str := StrTrim(root.SelectElement("encryptKey").Text()) //秘钥
	if str != "" {
		buf, err := NewEncrypter(EncryptDES_ECB, []byte(DefaultEncryptKey)).Decrypt([]byte(str), true)
		if err == nil {
			if len(buf) == 8 {
				dm.EncryptKey = string(buf) //new key
			} else {
				ErrorCaller("EncryptKey length!=8", caller)
				return
			}
		} else {
			ErrorCaller(ErrorMsg(err, "EncryptKey wrong"), caller)
			return
		}
	}

	str = StrTrim(root.SelectElement("defaultDB").Text()) //默认数据库
	if str != "" {
		dm.DefaultName = str
	}

	if dm.EncryptKey == "" { //default key
		dm.EncryptKey = DefaultEncryptKey
	}

	if dm.DBList == nil { //empty list
		dm.DBList = make(map[string]*DBConfig)
	}
	//-------------------------------------------------------------------------

	for _, conn := range root.SelectElement("conn").ChildElements() {
		dbname := conn.SelectAttr("name").Value
		str = conn.SelectElement("type").Text()
		if !StrIn(str, SQLDB_Types...) { //no match db-type
			ErrorCaller(fmt.Sprintf(`"%s" invalid db-type.`, dbname), caller)
			continue
		}

		db := &DBConfig{
			Name:    dbname,
			Type:    str,
			DSN:     conn.SelectElement("dsn").Text(),
			Drive:   conn.SelectElement("driver").Text(),
			User:    conn.SelectElement("user").Text(),
			Passwd:  conn.SelectElement("password").Text(),
			Host:    conn.SelectElement("host").Text(),
			MaxIdle: cast.ToInt(conn.SelectElement("maxIdle").Text()),
			MaxOpen: cast.ToInt(conn.SelectElement("maxOpen").Text()),
		}

		buf, err := NewEncrypter(EncryptDES_ECB, []byte(dm.EncryptKey)).Decrypt([]byte(db.Passwd), true)
		if err != nil {
			ErrorCaller(ErrorMsg(err, fmt.Sprintf(`"%s.passwd" wrong`, dbname)), caller)
			continue
		}

		dsnPart := strings.Split(db.DSN, "\n")
		if len(dsnPart) > 1 {
			db.DSN = ""
			for _, v := range dsnPart { //多行dsn合并
				db.DSN = db.DSN + StrTrim(v)
			}
		}

		db.Passwd = string(buf)
		db.ApplyDSN() //update value
		dm.DBList[db.Name] = db

		if db.MaxOpen < 1 {
			db.MaxOpen = 5
		}
		if db.MaxIdle < 1 {
			db.MaxIdle = 2
		}

		if dm.DefaultName == "" { //first is default
			dm.DefaultName = db.Name
			dm.DefaultType = db.Type
		} else if strings.EqualFold(db.Name, dm.DefaultName) { //match default type
			dm.DefaultType = db.Type
		}
	}

	if len(dm.DBList) < 1 {
		ErrorCaller("db-list is empty", caller)
	}
}

// loadMqttConfig 2024-01-09 16:54:33
/*
 参数: root,mqtt配置根节点
 描述: 载入mqtt外部配置
*/
func loadMqttConfig(root *etree.Element) {
	if root == nil {
		return
	}

	configStatus.mqtt = root.SelectAttr("enable").Value == StrTrue
	//status

	caller := "znlib.config.loadMqttConfig"
	var val int
	var err error
	var str string

	str = root.SelectElement("broker").Text()
	brokers := strings.Split(str, ",")

	for _, v := range brokers {
		Mqtt.Options.AddBroker(v)
		//多服务器支持
	}

	node := root.SelectElement("auth")
	tmp := node.SelectElement("clientID")
	str = StrTrim(tmp.Text())

	sufLen := cast.ToInt(tmp.SelectAttr("auto").Value)
	if sufLen > 0 { //随机id
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
	str = StrTrim(node.SelectElement("user").Text())
	if str != "" {
		Mqtt.Options.SetUsername(str)
		//user-name
	}

	str = StrTrim(node.SelectElement("password").Text())
	if str != "" {
		buf, err := NewEncrypter(EncryptDES_ECB, []byte(DefaultEncryptKey)).Decrypt([]byte(str), true)
		if err != nil {
			ErrorCaller(err, caller)
			return
		}

		Mqtt.Options.SetPassword(string(buf))
		//user-password
	}

	node = root.SelectElement("utils")
	str = StrTrim(node.SelectElement("encryptKey").Text())
	if str != "" {
		buf, err := NewEncrypter(EncryptDES_ECB, []byte(DefaultEncryptKey)).Decrypt([]byte(str), true)
		if err != nil {
			ErrorCaller(err, caller)
			return
		}

		MqttUtils.msgKey = string(buf)
		//消息加密密钥
	}

	str = StrTrim(node.SelectElement("verifyMsg").Text())
	if str != "" {
		MqttUtils.msgVerify = str == StrTrue
		//启用消息有效性验证
	}

	val, err = cast.ToIntE(node.SelectElement("workerNum").Text())
	if err == nil && val > 0 {
		MqttUtils.workerNum = val
		//消息工作对象个数
	}

	val, err = cast.ToIntE(node.SelectElement("delayWarn").Text())
	if str != "" {
		MqttUtils.msgDelay = time.Duration(val) * time.Second
	}

	tmp = node.SelectElement("zipData")
	str = StrTrim(tmp.Text())
	if str != "" {
		MqttUtils.msgZip = str == StrTrue
		//启用数据压缩
	}

	str = StrTrim(tmp.SelectAttr("min").Value)
	if str != "" {
		MqttUtils.msgZipLen = cast.ToInt(str)
		//数据超过min长度时开始压缩
	}

	node = root.SelectElement("tls")
	if node.SelectAttr("use").Value == StrTrue {
		str = FixPathVar(node.SelectElement("ca").Text())
		if !FileExists(str, false) { //ca
			ErrorCaller("tls.ca: file not exists", caller)
			return
		}

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

		var certs []tls.Certificate
		func() { //加载用户证书
			str = FixPathVar(node.SelectElement("crt").Text())
			key := FixPathVar(node.SelectElement("key").Text())

			if str == "" || key == "" { //not set
				return
			}

			if !FileExists(str, false) {
				ErrorCaller("user crt file not exists", caller)
				return
			}

			if !FileExists(key, false) {
				ErrorCaller("user key file not exists", caller)
				return
			}

			cert, err := tls.LoadX509KeyPair(str, key)
			if err != nil {
				ErrorCaller(err, caller)
				return
			}

			certs = []tls.Certificate{cert}
		}()

		Mqtt.Options.SetTLSConfig(&tls.Config{
			RootCAs:            cp,
			ClientAuth:         tls.NoClientCert,
			ClientCAs:          nil,
			InsecureSkipVerify: true,
			Certificates:       certs,
		})
	}

	getMqttTopic(root.SelectElement("subTopic"), true, caller)
	//订阅主题列表
	getMqttTopic(root.SelectElement("pubTopics"), false, caller)
	//发布主题列表
}

// getMqttTopic 2024-02-01 19:49:37
/*
 参数: root,主题配置根节点
 参数: isSub,是否订阅主题
 描述: 读取root中的主题配置
*/
func getMqttTopic(root *etree.Element, isSub bool, caller string) (topics map[string]MqttQos) {
	if root == nil {
		return nil
	}

	nodes := root.ChildElements()
	if len(nodes) < 1 {
		return nil
	}

	topics = make(map[string]MqttQos, len(nodes))
	for _, v := range nodes {
		str := v.Text()
		qos := MqttQos(cast.ToInt8(v.SelectAttr("qos").Value))
		if qos != MqttQos0 && qos != MqttQos1 && qos != MqttQos2 {
			ErrorCaller(fmt.Sprintf("%s: invalid qos", str), caller)
			continue
		}

		topics[str] = qos
		str = strings.ReplaceAll(str, "$id", Mqtt.Options.ClientID)
		//替换id标识

		if isSub {
			Mqtt.subTopics[str] = qos
		} else {
			Mqtt.pubTopics[str] = qos

		}
	}

	return topics
}

// loadMqttSSHConfig 2024-01-09 16:54:33
/*
 参数: root,ssh配置根节点
 描述: 载入mqttSSH外部配置
*/
func loadMqttSSHConfig(root *etree.Element) {
	if root == nil {
		return
	}

	caller := "znlib.config.loadMqttSSHConfig"
	var str string
	MqttSSH.enabled = root.SelectAttr("enable").Value == StrTrue

	node := root.SelectElement("auth")
	str = StrTrim(node.SelectElement("host").Text())
	if str != "" {
		MqttSSH.host = str
	}

	str = StrTrim(node.SelectElement("user").Text())
	if str != "" {
		MqttSSH.user = str
	}

	str = StrTrim(node.SelectElement("password").Text())
	if str != "" {
		buf, err := NewEncrypter(EncryptDES_ECB, []byte(DefaultEncryptKey)).Decrypt([]byte(str), true)
		if err != nil {
			ErrorCaller(err, caller)
			return
		}

		MqttSSH.password = string(buf)
		//用户密码
	}

	node = root.SelectElement("mqtt")
	topics := getMqttTopic(node.SelectElement("channel"), true, caller)
	if topics != nil {
		for k, v := range topics { //获取通道列表
			MqttSSH.channel[k] = v
		}
	}

	str = StrTrim(node.SelectElement("command").Text())
	if str != "" {
		MqttSSH.SshCmd = cast.ToUint8(str)
	}

	node = root.SelectElement("timeout")
	val, err := cast.ToDurationE(node.SelectElement("conn").Text())
	if err == nil && val > 0 {
		MqttSSH.connTimeout = val
	}

	val, err = cast.ToDurationE(node.SelectElement("exit").Text())
	if err == nil && val > 0 {
		MqttSSH.exitTimeout = val
	}

}
