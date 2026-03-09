// Package mqtt
/******************************************************************************
  作者: dmzn@163.com 2024-01-09 22:33:40
  描述: 支持tls的mqtt客户端

备注:
*.使用方法
  //1.启动
  Client.Start(func(cli mt.Client, msg mt.Message) {
    Info(string(msg.Topic()) + string(msg.Payload()))
  })
  //2.发布
  Client.Publish("", 0, []byte("hello")
  //3.停止
  Client.Stop()
******************************************************************************/
package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"reflect"

	. "github.com/dmznlin/znlib-go/znlib"
	mt "github.com/eclipse/paho.mqtt.golang"
)

// Qos qos定义
type Qos = byte

const (
	Qos0    Qos = 0  //最多交付一次
	Qos1    Qos = 1  //至少交付一次
	Qos2    Qos = 2  //只交付一次
	QosNone Qos = 27 //使用配置文件中的 qos
)

type (
	Event        = byte              //event 代码
	EventHandler = func(event Event) //event 事件
)

const (
	EventConnected    Event = iota //连接broker 成功
	EventDisconnect                //broker 连接断开
	EventServiceStart              //mqtt 服务启动
	EventServiceStop               //mqtt 服务停止
)

// Utils 辅助类
type Utils struct {
	Client    mt.Client
	Options   *mt.ClientOptions //选项
	SubTopics map[string]Qos    //订阅主题
	PubTopics map[string]Qos    //发布主题
	events    []EventHandler    //事件处理列表
}

// Client 客户端
var Client = &Utils{
	Options:   mt.NewClientOptions(),
	SubTopics: make(map[string]Qos),
	PubTopics: make(map[string]Qos),
	events:    make([]EventHandler, 0),
}

// initMqtt 2024-01-09 17:03:09
/*
 描述: 初始化mqtt连接
*/
func init() {
	Application.RegisterExitHandler(func() {
		Client.Stop()
		//退出时停止
	})

	Application.RegisterInitHandler(func(cfg *LibConfig) {
		Client.ApplyConfig(&cfg.Mqtt)
	})
}

// ApplyConfig 2026-03-09 15:04:52
/*
 参数: cfg,mqtt配置
 描述:
*/
func (mc *Utils) ApplyConfig(cfg *MqttConfig) {
	if !cfg.Enable {
		return
	}

	caller := "znlib.mqtt.ApplyConfig"
	if len(cfg.Broker) < 1 {
		ErrorCaller("mqtt.broker is empty", caller)
		return
	}

	if cfg.Tls.Used {
		cfg.Tls.CA = FixPathVar(cfg.Tls.CA)
		if !FileExists(cfg.Tls.CA, false) {
			ErrorCaller("mqtt.Tls.ca is miss", caller)
			return
		}

		cfg.Tls.Key = FixPathVar(cfg.Tls.Key)
		if !FileExists(cfg.Tls.Key, false) {
			ErrorCaller("mqtt.Tls.key is miss", caller)
			return
		}

		cfg.Tls.Cert = FixPathVar(cfg.Tls.Cert)
		if !FileExists(cfg.Tls.Cert, false) {
			ErrorCaller("mqtt.Tls.cert is miss", caller)
			return
		}

		rootCA, err := os.ReadFile(cfg.Tls.CA)
		if err != nil {
			ErrorCaller(err, caller)
			return
		}

		cp := x509.NewCertPool()
		if !cp.AppendCertsFromPEM(rootCA) {
			ErrorCaller("mqtt.Tls.ca load error", caller)
			return
		}

		cert, err := tls.LoadX509KeyPair(cfg.Tls.Cert, cfg.Tls.Key)
		if err != nil {
			ErrorCaller(err, caller)
			return
		}

		mc.Options.SetTLSConfig(&tls.Config{
			RootCAs:            cp,
			ClientAuth:         tls.NoClientCert,
			ClientCAs:          nil,
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{cert},
		})
	}

	if cfg.Password != "" { // broker 密码
		buf, err := NewEncrypter(EncryptDesEcb, []byte(DefaultEncryptKey)).Decrypt([]byte(cfg.Password), true)
		if err != nil {
			ErrorCaller("mqtt.pwd is invalid", caller)
			return
		}

		cfg.Password = string(buf)
	}

	if cfg.IDAuto > 0 { //自动生成 Client-id
		idLen := 23 - len(cfg.ClientID) //mqtt id长度限制
		if cfg.IDAuto > idLen {         //取最大可用长度
			cfg.IDAuto = idLen
		}

		//new id
		cfg.ClientID = cfg.ClientID + SerialID.MakeID(cfg.IDAuto)
	}

	for _, v := range cfg.TopicSub {
		v.Topic = StrReplace(v.Topic, cfg.ClientID, "$id")
		//更新数据通道标识
		mc.SubTopics[v.Topic] = v.Qos
		//订阅主题
	}

	for _, v := range cfg.TopicPub {
		v.Topic = StrReplace(v.Topic, cfg.ClientID, "$id")
		//更新数据通道标识
	}

	mc.Options.SetClientID(cfg.ClientID)
	//更新 Client id

	if cfg.User != "" {
		mc.Options.SetUsername(cfg.User)
		//user-name
	}

	if cfg.Password != "" {
		mc.Options.SetPassword(cfg.Password)
		//user-password
	}

	for _, v := range cfg.Broker { //多服务器支持
		mc.Options.AddBroker(v)
	}

	if mc.Options.OnConnect == nil {
		mc.Options.SetOnConnectHandler(func(client mt.Client) {
			var host string
			for _, v := range mc.Options.Servers {
				if host == "" {
					host = v.String()
				} else {
					host = host + "," + v.String()
				}
			}

			Info("znlib.mqtt.connected: " + host)
			_ = mc.subscribeMultiple(client) //连接成功后,重新订阅主题
			mc.eventAction(EventConnected)   //触发已连接事件
		})
	}

	if mc.Options.OnConnectionLost == nil {
		mc.Options.SetConnectionLostHandler(func(client mt.Client, err error) {
			ErrorCaller(err, "znlib.mqtt.lostconnect")
			mc.eventAction(EventDisconnect) //触发断开事件
		})
	}

	if mc.Options.OnReconnecting == nil {
		mc.Options.SetReconnectingHandler(func(client mt.Client, options *mt.ClientOptions) {
			Info("znlib.mqtt: reconnect broker")
		})
	}
}

// Start 2024-01-11 08:24:20
/*
 参数: msgHandler,消息处理函数
 描述: 启动mqtt服务
*/
func (mc *Utils) Start(msgHandler mt.MessageHandler) error {
	if mc.Client != nil {
		return nil
	}

	if msgHandler != nil {
		mc.Options.SetDefaultPublishHandler(msgHandler)
		//默认消息处理句柄
	}

	mc.Client = mt.NewClient(mc.Options)
	//创建链路
	token := mc.Client.Connect()
	//连接 broker

	if token.Wait() && token.Error() != nil {
		ErrorCaller(token.Error(), "znlib.mqtt.Start")
	}

	mc.eventAction(EventServiceStart)
	//触发服务启动事件
	return token.Error()
}

// Stop 2024-01-14 15:23:20
/*
 描述: 停止mqtt服务
*/
func (mc *Utils) Stop() {
	if mc.Client != nil {
		_ = mc.Unsubscribe(mc.Client)
		//退订主题
		mc.Client.Disconnect(500)
		//断开链路
		mc.Client = nil
	}

	mc.eventAction(EventServiceStop)
	//触发服务停止事件
}

// Publish 2026-02-27 14:43:07
/*
 参数: topic,主题
 参数: qos,送达级别
 参数: msg,消息
 描述: 向topic发布msg消息
*/
func (mc *Utils) Publish(topic string, qos Qos, msg []byte) {
	pub := func() {
		token := mc.Client.Publish(topic, qos, false, msg)
		if token.Wait() && token.Error() != nil {
			ErrorCaller(token.Error(), "znlib.mqtt.publish")
		}
	}

	if topic == "" { //
		var q Qos
		useCfg := qos == QosNone
		//使用配置 qos

		for topic, q = range mc.PubTopics {
			if useCfg {
				qos = q
			}

			pub()
		}
	} else {
		if qos == QosNone {
			q, ok := mc.PubTopics[topic]
			if ok {
				qos = q
			} else {
				qos = Qos0
			}
		}

		pub()
		//自定义主题
	}
}

// Subscribe 2024-01-14 14:52:25
/*
 参数: topic,主题
 参数: qos,模式
 参数: ctl,控制参数(1.true:重置订阅;2.true:立即订阅)
 描述: 新增订阅topic主题
*/
func (mc *Utils) Subscribe(topic string, qos Qos, ctl ...bool) error {
	reset := false //不重置
	via := true    //立即发送
	cl := len(ctl)

	if cl > 0 {
		reset = ctl[0]
	}

	if cl > 1 {
		via = ctl[1]
	}

	if reset {
		mc.SubTopics = make(map[string]Qos)
		//重置为空
	}

	mc.SubTopics[topic] = qos
	//添加新主题

	if via {
		return mc.subscribeMultiple(mc.Client)
		//开始订阅
	}

	return nil
}

// subscribeMultiple 2024-01-14 15:22:11
/*
 参数: Client,链路
 描述: 订阅主题列表
*/
func (mc *Utils) subscribeMultiple(client mt.Client) error {
	if len(mc.SubTopics) < 1 {
		return nil
	}

	token := client.SubscribeMultiple(mc.SubTopics, nil)
	if token.Wait() && token.Error() == nil {
		Info(fmt.Sprintf("znlib.mqtt.subscribe: %v", mc.SubTopics))
	} else {
		ErrorCaller(token.Error(), "znlib.mqtt.subscribe")
	}

	return token.Error()
}

// Unsubscribe 2026-03-03 11:30:37
/*
 参数: Client,链路
 描述: 退订所有主题
*/
func (mc *Utils) Unsubscribe(client mt.Client) error {
	if client == nil {
		client = mc.Client
		//use default
	}

	idx := len(mc.SubTopics)
	if idx > 0 && client.IsConnected() { //退订所有主题
		topics := make([]string, idx)
		idx = 0
		for k := range mc.SubTopics {
			topics[idx] = k
			idx++
		}

		token := client.Unsubscribe(topics...)
		token.Wait()
		if token.Error() != nil {
			ErrorCaller(token.Error(), "znlib.mqtt.unsubscribe")
			return token.Error()
		}

		Info(fmt.Sprintf("znlib.mqtt.unsubscribe: %v", topics))
	}

	return nil
}

// RegisterEventHandler 2024-02-06 17:45:36
/*
 参数: fn,事件句柄
 描述: 添加fn处理mqtt事件
*/
func (mc *Utils) RegisterEventHandler(fn EventHandler) {
	if IsNil(fn) {
		return
	}

	Application.SyncLock.Lock()
	defer Application.SyncLock.Unlock()
	pFun := reflect.ValueOf(fn)

	for _, v := range mc.events {
		if reflect.ValueOf(v).Pointer() == pFun.Pointer() { //重复注册
			return
		}
	}

	mc.events = append(mc.events, fn)
	//注册
}

// eventAction 2024-02-06 12:22:53
/*
 参数: event,事件代码
 描述: 触发一个event事件
*/
func (mc *Utils) eventAction(event Event) {
	defer DeferHandle(false, "znlib.mqtt.eventAction")
	for _, do := range mc.events {
		do(event)
	}
}
