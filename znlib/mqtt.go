// Package znlib
/******************************************************************************
  作者: dmzn@163.com 2024-01-09 22:33:40
  描述: 支持tls的mqtt客户端

备注:
*.使用方法
  //1.启动
  Mqtt.Start(func(client mt.Client, msg mt.Message) {
    Info(string(msg.Topic()) + string(msg.Payload()))
  })
  //2.发布
  Mqtt.Publish("", []string{"aa", "bb", "cc"})
  //3.停止
  Mqtt.Stop()
******************************************************************************/
package znlib

import (
	"fmt"
	mt "github.com/eclipse/paho.mqtt.golang"
	"reflect"
)

// MqttQos qos定义
type MqttQos = byte

const (
	MqttQos0    MqttQos = 0  //最多交付一次
	MqttQos1    MqttQos = 1  //至少交付一次
	MqttQos2    MqttQos = 2  //只交付一次
	MqttQosNone MqttQos = 27 //使用配置文件中的qos
)

type (
	MqttEvent        = byte                  //event代码
	MqttEventHandler = func(event MqttEvent) //event事件
)

const (
	MqttEventConnected    MqttEvent = iota //连接broker成功
	MqttEventDisconnect                    //broker连接断开
	MqttEventServiceStart                  //mqtt服务启动
	MqttEventServiceStop                   //mqtt服务停止
)

// mqttClient 客户端参数
type mqttClient struct {
	client    mt.Client
	Options   *mt.ClientOptions
	subTopics map[string]MqttQos
	pubTopics map[string]MqttQos
	events    []MqttEventHandler
}

// Mqtt 客户端
var Mqtt = &mqttClient{
	Options:   mt.NewClientOptions(),
	subTopics: make(map[string]MqttQos),
	pubTopics: make(map[string]MqttQos),
	events:    make([]MqttEventHandler, 0),
}

// initMqtt 2024-01-09 17:03:09
/*
 描述: 初始化mqtt连接
*/
func initMqtt() {
	Application.RegisterExitHandler(func() {
		Mqtt.Stop()
		//退出时停止
	})

	if Mqtt.Options.OnConnect == nil {
		Mqtt.Options.SetOnConnectHandler(func(client mt.Client) {
			var host string
			for _, v := range Mqtt.Options.Servers {
				if host == "" {
					host = v.String()
				} else {
					host = host + "," + v.String()
				}
			}

			Info("znlib.mqtt.connect: " + host)
			//log
			_ = Mqtt.subscribeMultiple(client)
			//连接成功后,重新订阅主题
			Mqtt.eventAction(MqttEventConnected)
			//触发已连接事件
		})
	}

	if Mqtt.Options.OnConnectionLost == nil {
		Mqtt.Options.SetConnectionLostHandler(func(client mt.Client, err error) {
			ErrorCaller(err, "znlib.mqtt.lostconnect")
			//log
			Mqtt.eventAction(MqttEventDisconnect)
			//触发断开事件
		})
	}

	if Mqtt.Options.OnReconnecting == nil {
		Mqtt.Options.SetReconnectingHandler(func(client mt.Client, options *mt.ClientOptions) {
			Info("znlib.mqtt.reconnect_broker.")
			//log
		})
	}
}

// RegisterEventHandler 2024-02-06 17:45:36
/*
 参数: fn,事件句柄
 描述: 添加fn处理mqtt事件
*/
func (mc *mqttClient) RegisterEventHandler(fn MqttEventHandler) {
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
func (mc *mqttClient) eventAction(event MqttEvent) {
	defer DeferHandle(false, "znlib.mqtt.eventAction")
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

	mc.client = mt.NewClient(Mqtt.Options)
	//创建链路
	token := mc.client.Connect()
	//连接broker

	if token.Wait() && token.Error() != nil {
		ErrorCaller(token.Error(), "znlib.mqtt.connect_broker")
	}

	mc.eventAction(MqttEventServiceStart)
	//触发服务启动事件
	return token.Error()
}

// StartWithUtils 2024-01-19 11:21:32
/*
 参数: msgHandle,消息处理函数
 描述: 使用mqttutils处理消息,支持加密和消息缓存
*/
func (mc *mqttClient) StartWithUtils(msgHandle MqttHandler) error {
	MqttUtils.RegisterHandler(msgHandle)
	//注册外部处理
	MqttUtils.addWorkers()
	//启动工作对象
	return mc.Start(MqttUtils.onMessge)
	//启动mqtt
}

// Stop 2024-01-14 15:23:20
/*
 描述: 停止mqtt服务
*/
func (mc *mqttClient) Stop() {
	if mc.client == nil {
		return
	}

	if mc.client.IsConnected() { //退订所有主题
		idx := len(mc.subTopics)
		topics := make([]string, idx)
		idx = 0

		for k := range mc.subTopics {
			topics[idx] = k
			idx++
		}

		token := mc.client.Unsubscribe(topics...)
		token.Wait()
		if token.Error() != nil {
			ErrorCaller(token.Error(), "znlib.mqtt.unsubscribe")
		} else {
			Info(fmt.Sprintf("znlib.mqtt.unsubscribe: %v", topics))
		}
	}

	mc.client.Disconnect(500)
	//断开链路
	mc.client = nil

	mc.eventAction(MqttEventServiceStop)
	//触发服务停止事件
}

// Publish 2024-01-10 15:32:26
/*
 参数: topic,主题
 参数: qos,送达级别
 参数: msg,消息列表
 描述: 向topic发布msg消息
*/
func (mc *mqttClient) Publish(topic string, qos MqttQos, msg [][]byte) {
	pub := func() {
		for _, v := range msg {
			token := mc.client.Publish(topic, qos, false, v)
			if token.Wait() && token.Error() != nil {
				ErrorCaller(token.Error(), "znlib.mqtt.publish")
			}
		}
	}

	if topic == "" {
		var q MqttQos
		useCfg := qos == MqttQosNone
		//使用配置qos

		for topic, q = range mc.pubTopics {
			if useCfg {
				qos = q
			}
			pub()
		}
	} else {
		pub()
		//自定义主题
	}
}

// Subscribe 2024-01-14 14:52:25
/*
 参数: topics,主题列表
 参数: clear,清空原列表
 描述: 新增订阅topics主题
*/
func (mc *mqttClient) Subscribe(topics map[string]MqttQos, clear bool) error {
	if clear {
		mc.subTopics = make(map[string]MqttQos)
	}

	for k, v := range topics {
		mc.subTopics[k] = v
	}

	return mc.subscribeMultiple(mc.client)
}

// subscribeMultiple 2024-01-14 15:22:11
/*
 参数: client,链路
 描述: 订阅主题列表
*/
func (mc *mqttClient) subscribeMultiple(client mt.Client) error {
	if len(Mqtt.subTopics) < 1 {
		return nil
	}

	token := client.SubscribeMultiple(Mqtt.subTopics, nil)
	if token.Wait() && token.Error() == nil {
		Info(fmt.Sprintf("znlib.mqtt.subscribe: %v", Mqtt.subTopics))
	} else {
		ErrorCaller(token.Error(), "znlib.mqtt.subscribe")
	}

	return token.Error()
}
