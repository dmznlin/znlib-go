// Package mqtt
/******************************************************************************
  作者: dmzn@163.com 2024-01-09 22:33:40
  描述: 支持tls的mqtt客户端

备注:
*.使用方法
  //1.启动
  Client.Start(func(client mt.mqttClient, msg mt.Message) {
    Info(string(msg.Topic()) + string(msg.Payload()))
  })
  //2.发布
  Client.Publish("", []string{"aa", "bb", "cc"})
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

// mqttClient 客户端参数
type mqttClient struct {
	client    mt.Client
	Options   *mt.ClientOptions
	subTopics map[string]Qos
	pubTopics map[string]Qos
	events    []EventHandler
}

// Client 客户端
var Client = &mqttClient{
	Options:   mt.NewClientOptions(),
	subTopics: make(map[string]Qos),
	pubTopics: make(map[string]Qos),
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
		caller := "znlib.mqtt.init"
		if len(cfg.Mqtt.Broker) < 1 {
			ErrorCaller("mqtt.broker is empty", caller)
			return
		}

		if cfg.Mqtt.Tls.Used {
			cfg.Mqtt.Tls.CA = FixPathVar(cfg.Mqtt.Tls.CA)
			if !FileExists(cfg.Mqtt.Tls.CA, false) {
				ErrorCaller("mqtt.Tls.ca is miss", caller)
				return
			}

			cfg.Mqtt.Tls.Key = FixPathVar(cfg.Mqtt.Tls.Key)
			if !FileExists(cfg.Mqtt.Tls.Key, false) {
				ErrorCaller("mqtt.Tls.key is miss", caller)
				return
			}

			cfg.Mqtt.Tls.Cert = FixPathVar(cfg.Mqtt.Tls.Cert)
			if !FileExists(cfg.Mqtt.Tls.Cert, false) {
				ErrorCaller("mqtt.Tls.cert is miss", caller)
				return
			}

			rootCA, err := os.ReadFile(cfg.Mqtt.Tls.CA)
			if err != nil {
				ErrorCaller(err, caller)
				return
			}

			cp := x509.NewCertPool()
			if !cp.AppendCertsFromPEM(rootCA) {
				ErrorCaller("mqtt.Tls.ca 配置错误: 无法加载", caller)
				return
			}

			cert, err := tls.LoadX509KeyPair(cfg.Mqtt.Tls.Cert, cfg.Mqtt.Tls.Key)
			if err != nil {
				ErrorCaller(err, caller)
				return
			}

			Client.Options.SetTLSConfig(&tls.Config{
				RootCAs:            cp,
				ClientAuth:         tls.NoClientCert,
				ClientCAs:          nil,
				InsecureSkipVerify: true,
				Certificates:       []tls.Certificate{cert},
			})
		}

		if cfg.Mqtt.Password != "" { // broker 密码
			buf, err := NewEncrypter(EncryptDesEcb, []byte(DefaultEncryptKey)).Decrypt([]byte(cfg.Mqtt.Password), true)
			if err != nil {
				ErrorCaller("mqtt.pwd is invalid", caller)
				return
			}

			cfg.Mqtt.Password = string(buf)
		}

		if cfg.Mqtt.IDAuto > 0 { //自动生成 client-id
			idLen := 23 - len(cfg.Mqtt.ClientID) //mqtt id长度限制
			if cfg.Mqtt.IDAuto > idLen {         //取最大可用长度
				cfg.Mqtt.IDAuto = idLen
			}

			//new id
			cfg.Mqtt.ClientID = cfg.Mqtt.ClientID + SerialID.MakeID(cfg.Mqtt.IDAuto)
		}

		for _, v := range cfg.Mqtt.TopicSub {
			v.Topic = StrReplace(v.Topic, cfg.Mqtt.ClientID, "$id")
			//更新数据通道标识
		}

		for _, v := range cfg.Mqtt.TopicPub {
			v.Topic = StrReplace(v.Topic, cfg.Mqtt.ClientID, "$id")
			//更新数据通道标识
		}

		Client.Options.SetClientID(cfg.Mqtt.ClientID)
		//更新 client id

		if cfg.Mqtt.User != "" {
			Client.Options.SetUsername(cfg.Mqtt.User)
			//user-name
		}

		if cfg.Mqtt.Password != "" {
			Client.Options.SetPassword(cfg.Mqtt.Password)
			//user-password
		}

		for _, v := range cfg.Mqtt.Broker { //多服务器支持
			Client.Options.AddBroker(v)
		}

		if Client.Options.OnConnect == nil {
			Client.Options.SetOnConnectHandler(func(client mt.Client) {
				var host string
				for _, v := range Client.Options.Servers {
					if host == "" {
						host = v.String()
					} else {
						host = host + "," + v.String()
					}
				}

				Info("mqtt.connected: " + host)
				_ = Client.subscribeMultiple(client) //连接成功后,重新订阅主题
				Client.eventAction(EventConnected)   //触发已连接事件
			})
		}

		if Client.Options.OnConnectionLost == nil {
			Client.Options.SetConnectionLostHandler(func(client mt.Client, err error) {
				ErrorCaller(err, "mqtt.lostconnect")
				Client.eventAction(EventDisconnect) //触发断开事件
			})
		}

		if Client.Options.OnReconnecting == nil {
			Client.Options.SetReconnectingHandler(func(client mt.Client, options *mt.ClientOptions) {
				Info("mqtt.reconnect_broker.")
			})
		}
	})

}

// RegisterEventHandler 2024-02-06 17:45:36
/*
 参数: fn,事件句柄
 描述: 添加fn处理mqtt事件
*/
func (mc *mqttClient) RegisterEventHandler(fn EventHandler) {
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
func (mc *mqttClient) eventAction(event Event) {
	defer DeferHandle(false, "mqtt.eventAction")
	for _, do := range mc.events {
		do(event)
	}
}

// Start 2024-01-11 08:24:20
/*
 参数: msgHandler,消息处理函数
 描述: 启动mqtt服务
*/
func (mc *mqttClient) Start(msgHandler mt.MessageHandler) error {
	if mc.client != nil {
		return nil
	}

	if msgHandler != nil {
		mc.Options.SetDefaultPublishHandler(msgHandler)
		//默认消息处理句柄
	}

	mc.client = mt.NewClient(Client.Options)
	//创建链路
	token := mc.client.Connect()
	//连接 broker

	if token.Wait() && token.Error() != nil {
		ErrorCaller(token.Error(), "mqtt.connect_broker")
	}

	mc.eventAction(EventServiceStart)
	//触发服务启动事件
	return token.Error()
}

// Stop 2024-01-14 15:23:20
/*
 描述: 停止mqtt服务
*/
func (mc *mqttClient) Stop() {
	if mc.client != nil {
		mc.Unsubscribe(mc.client)
		//退订主题
		mc.client.Disconnect(500)
		//断开链路
		mc.client = nil
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
func (mc *mqttClient) Publish(topic string, qos Qos, msg []byte) {
	pub := func() {
		token := mc.client.Publish(topic, qos, false, msg)
		if token.Wait() && token.Error() != nil {
			ErrorCaller(token.Error(), "znlib.mqtt.publish")
		}
	}

	if topic == "" { //
		var q Qos
		useCfg := qos == QosNone
		//使用配置 qos

		for topic, q = range mc.pubTopics {
			if useCfg {
				qos = q
			}

			pub()
		}
	} else {
		if qos == QosNone {
			q, ok := mc.pubTopics[topic]
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
 描述: 新增订阅topic主题
*/
func (mc *mqttClient) Subscribe(topic string, qos Qos) error {
	mc.subTopics[topic] = qos
	return mc.subscribeMultiple(mc.client)
}

// subscribeMultiple 2024-01-14 15:22:11
/*
 参数: client,链路
 描述: 订阅主题列表
*/
func (mc *mqttClient) subscribeMultiple(client mt.Client) error {
	if len(Client.subTopics) < 1 {
		return nil
	}

	token := client.SubscribeMultiple(Client.subTopics, nil)
	if token.Wait() && token.Error() == nil {
		Info(fmt.Sprintf("znlib.mqtt.subscribe: %v", Client.subTopics))
	} else {
		ErrorCaller(token.Error(), "znlib.mqtt.subscribe")
	}

	return token.Error()
}

// Unsubscribe 2026-03-03 11:30:37
/*
 参数: client,链路
 描述: 退订所有主题
*/
func (mc *mqttClient) Unsubscribe(client mt.Client) error {
	idx := len(mc.subTopics)
	if idx > 0 && client.IsConnected() { //退订所有主题
		topics := make([]string, idx)
		idx = 0
		for k := range mc.subTopics {
			topics[idx] = k
			idx++
		}

		token := client.Unsubscribe(topics...)
		token.Wait()
		if token.Error() != nil {
			ErrorCaller(token.Error(), "mqtt.unsubscribe")
			return token.Error()
		}

		Info(fmt.Sprintf("mqtt.unsubscribe: %v", topics))
	}

	return nil
}
