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
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/dmznlin/znlib-go/znlib/copier"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type (
	// initLibUtils 初始化函数集
	initLibUtils = func()

	// LoggerConfig  默认日志配置参数
	LoggerConfig = struct {
		FilePath string        `json:"filePath"` //日志目录
		FileName string        `json:"fileName"` //日志文件名
		Level    string        `json:"logLevel"` //日志级别
		LogLevel logrus.Level  `json:"-"`
		MaxAge   time.Duration `json:"maxAge"`   //日志保存天数
		Colorful bool          `json:"colorful"` //使用彩色终端
	}

	// SnowflakeConfig 雪花算法配置
	SnowflakeConfig struct {
		Enable     bool  `json:"enable"`     //启用
		WorkerID   int64 `json:"worker"`     //节点标识
		Datacenter int64 `json:"datacenter"` //数据中心标识
	}

	RedisTimeout struct {
		Dial  time.Duration `json:"dial"`  //连接建立超时
		Read  time.Duration `json:"read"`  //读超时
		Write time.Duration `json:"write"` //写超时
		Pool  time.Duration `json:"pool"`  //繁忙状态时等待
	}

	// RedisConfig redis配置
	RedisConfig = struct {
		Enable    bool         `json:"enable"`    //启用
		Cluster   bool         `json:"cluster"`   //是否集群
		Servers   []string     `json:"servers"`   //服务器列表
		Password  string       `json:"password"`  //服务密码
		PoolSize  int          `json:"poolSize"`  //最大连接数
		DefaultDB int          `json:"defaultDB"` //默认数据库索引
		Timeout   RedisTimeout `json:"timeout"`   //超时配置
	}

	// DbConn 数据库连接
	DbConn struct {
		Name   string    `json:"name"`   //数据库名称
		Type   SqlDbType `json:"type"`   //数据库类型
		Drive  string    `json:"drive"`  //驱动名称
		User   string    `json:"user"`   //登录用户
		Passwd string    `json:"passwd"` //登录密码
		Host   string    `json:"host"`   //主机地址
		DSN    string    `json:"dsn"`    //连接配置项

		MaxOpen int      `json:"maxOpen"` //同时打开的连接数(使用中+空闲)
		MaxIdle int      `json:"maxIdle"` //最大并发空闲链接数
		DB      *sqlx.DB `json:"-"`       //数据库对象
	}

	// DbConfig 数据库配置
	DbConfig struct {
		Enable      bool      `json:"enable"`     //启用
		EncryptKey  string    `json:"encryptKey"` //加密秘钥
		DefaultName string    `json:"defaultDB"`  //默认数据库名称
		DbConn      []*DbConn `json:"dbConn"`     //连接列表
	}

	MqttTopic = struct {
		Qos   byte   `json:"qos"`   //控制
		Topic string `json:"topic"` //主题
	}

	MqttTLS = struct {
		Used bool   `json:"use"`  //启用 tls
		CA   string `json:"ca"`   //ca 证书
		Key  string `json:"key"`  //客户端秘钥
		Cert string `json:"cert"` //客户端证书
	}

	// MqttConfig mqtt参数
	MqttConfig struct {
		Enable   bool         `json:"enable"` //启用
		Broker   []string     `json:"broker"` //服务器(集群)
		ClientID string       `json:"client"` //客户端标识
		IDAuto   int          `json:"auto"`   //以ClientID为前缀,自动增加n位随机id
		User     string       `json:"user"`   //用户名
		Password string       `json:"pwd"`    //登录密码(des)
		Tls      MqttTLS      `json:"tls"`    //接入认证
		TopicSub []*MqttTopic `json:"sub"`    //命令传输通道
		TopicPub []*MqttTopic `json:"pub"`    //数据传输通道
	}

	// LibConfig 配置文件结构体
	LibConfig struct {
		Logger LoggerConfig    `json:"logger"`        //日志配置
		App    any             `json:"app,omitempty"` //应用配置
		Snow   SnowflakeConfig `json:"snow"`          //雪花算法
		Redis  RedisConfig     `json:"redis"`         //redis
		DB     DbConfig        `json:"db"`            //数据库
		Mqtt   MqttConfig      `json:"mqtt"`          //mqtt
	}
)

var (
	// initLibOnce 确保一次初始化
	initLibOnce sync.Once

	// initBeforeUtil 初始化开始前执行操作
	initBeforeUtil initLibUtils

	// GlobalConfig lib 库全局参数配置
	GlobalConfig = LibConfig{
		Logger: LoggerConfig{
			FilePath: "$path/logs",
			FileName: "app_",
			Level:    "info",
			LogLevel: logrus.InfoLevel,
			MaxAge:   7,
			Colorful: false,
		},
		App: nil,
		Snow: SnowflakeConfig{
			Enable:     false,
			WorkerID:   1,
			Datacenter: 0,
		},
		Redis: RedisConfig{
			Enable:    false,
			Cluster:   false,
			Servers:   []string{"tcp://:1883"},
			Password:  "",
			PoolSize:  0,
			DefaultDB: 0,
			Timeout: RedisTimeout{
				Dial:  0, //连接建立超时时间,默认5秒
				Read:  0, //读超时,默认3秒,-1表示取消读超时
				Write: 0, //写超时,默认等于读超时,-1表示取消写超时
				Pool:  0, //当所有连接都处在繁忙状态时,客户端等待可用连接的最大等待时长,默认为读超时+1秒
			},
		},
		DB: DbConfig{
			Enable:      false,
			EncryptKey:  "",
			DefaultName: "mssql_main",
			DbConn: []*DbConn{
				{
					Name:    "mssql_main",
					Type:    "SqlServer",
					Drive:   "adodb",
					User:    "sa",
					Passwd:  "",
					Host:    "127.0.0.1",
					DSN:     "user=$user;pwd=$pwd;host=$host",
					MaxOpen: 5,
					MaxIdle: 2,
				},
			},
		},
		Mqtt: MqttConfig{
			Enable:   false,
			Broker:   []string{"tcp://broker.hivemq.com:1883"},
			ClientID: "mt-",
			IDAuto:   7,
			User:     "",
			Password: "",
			Tls: MqttTLS{
				Used: false,
				CA:   "$path/cert/ca.crt",
				Key:  "$path/cert/mqtt.key",
				Cert: "$path/cert/mqtt.crt",
			},
			TopicSub: []*MqttTopic{
				{
					Qos:   0,
					Topic: "",
				},
			},
			TopicPub: []*MqttTopic{
				{
					Qos:   0,
					Topic: "",
				},
			},
		},
	}
)

// LoadConfig 2026-03-08 10:37:00
/*
 参数: cfg,配置文件
 参数: val,配置变量
 描述: 载入 cfg 配置文件
*/
func LoadConfig(cfg string, val any) error {
	df, err := os.ReadFile(cfg)
	if err != nil {
		return fmt.Errorf("load config(%s): %w", cfg, err)
	}

	if err = json.Unmarshal(df, val); err != nil {
		return fmt.Errorf("unmarshal config(%s): %w", cfg, err)
	}

	return nil
}

// SaveConfig 2026-01-26 17:49:53
/*
 参数: cfg, 配置文件
 参数: val,配置变量
 描述: 保存通道信息到cfg中
*/
func SaveConfig(cfg string, val any) error {
	dt, err := json.MarshalIndent(val, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config(%s): %w", cfg, err)
	}

	if err := os.WriteFile(cfg, dt, 0644); err != nil {
		return fmt.Errorf("save config(%s): %w", cfg, err)
	}

	return nil
}

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
	initApp()
	//application.go

	if initBeforeUtil != nil { //执行外部设定
		initBeforeUtil()
	}

	var err error
	if FileExists(Application.ConfigFile, false) {
		var cfg LibConfig
		err = LoadConfig(Application.ConfigFile, &cfg) //加载配置
		if err == nil {
			_ = copier.Copy(&GlobalConfig, &cfg)
			//配置生效
		}
	} else {
		err = SaveConfig(Application.ConfigFile, &GlobalConfig)
		//生成默认配置
	}

	if err != nil {
		ErrorCaller(err, "znlib.initLibrary")
	}

	loadLogConfig(&GlobalConfig.Logger)
	initLogger(&GlobalConfig.Logger)
	//初始化日志

	for _, caller := range initCallers {
		caller(&GlobalConfig)
		//初始化各模块
	}
}

// loadLogConfig 2022-08-11 19:22:24
/*
 参数: cfg,日志参数
 描述: 载入日志外部配置
*/
func loadLogConfig(cfg *LoggerConfig) {
	if len(cfg.FilePath) < 1 {
		cfg.FilePath = Application.LogPath
	} else {
		cfg.FilePath = FixPathVar(cfg.FilePath)
		//替换路径中的变量
	}

	cfg.FilePath = FixPath(cfg.FilePath)
	//添加路径分隔符

	levels := []string{"trace", "debug", "info", "warning", "error", "fatal", "panic"}
	if !StrIn(cfg.Level, levels...) {
		cfg.Level = "info"
	}

	cfg.LogLevel, _ = logrus.ParseLevel(cfg.Level)
	Application.IsDebug = cfg.LogLevel == logrus.DebugLevel
	//全局 debug 开发
}
