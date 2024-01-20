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
	"math"
)

// mqttClient 客户端参数
type mqttClient struct {
	client    mt.Client
	Options   *mt.ClientOptions
	subTopics map[string]byte
	pubTopics map[string]byte
}

// Mqtt 客户端
var Mqtt = &mqttClient{
	Options:   mt.NewClientOptions(),
	subTopics: make(map[string]byte, 0),
	pubTopics: make(map[string]byte, 0),
}

// init_mqtt 2024-01-09 17:03:09
/*
 描述: 初始化mqtt连接
*/
func init_mqtt() {
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

			//连接成功后,重新订阅主题
			Mqtt.subscribeMultiple(client)
		})
	}

	if Mqtt.Options.OnConnectionLost == nil {
		Mqtt.Options.SetConnectionLostHandler(func(client mt.Client, err error) {
			ErrorCaller(err, "znlib.mqtt.lostconnect")
		})
	}

	if Mqtt.Options.OnReconnecting == nil {
		Mqtt.Options.SetReconnectingHandler(func(client mt.Client, options *mt.ClientOptions) {
			Info("znlib.mqtt.reconnect_broker.")
		})
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

	return token.Error()
}

// StartWithUtils 2024-01-19 11:21:32
/*
 参数: msgHandle,消息处理函数
 描述: 使用mqttutils处理消息,支持加密和消息缓存
*/
func (mc *mqttClient) StartWithUtils(msgHandle MqttHandler) error {
	MqttUtils.enabled = true
	//启用辅助类
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
		topics := make([]string, idx, idx)
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

	if MqttUtils.enabled {
		Info("znlib.mqtt.stop: wait worker exit")
		MqttUtils.workerGroup.Wait()
		Info("znlib.mqtt.stop: all worker has exit")
	}
}

// Publish 2024-01-10 15:32:26
/*
 参数: topic,主题
 参数: qos,送达级别
 参数: msg,消息内容
 描述: 向topic发布msg消息
*/
func (mc *mqttClient) Publish(topic string, qos byte, msg []string) {
	pub := func() {
		for _, v := range msg {
			token := mc.client.Publish(topic, qos, false, v)
			if token.Wait() && token.Error() != nil {
				ErrorCaller(token.Error(), "znlib.mqtt.publish")
			}
		}
	}

	if topic == "" {
		var q byte
		useCfg := qos == math.MaxUint8
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
func (mc *mqttClient) Subscribe(topics map[string]byte, clear bool) error {
	if clear {
		mc.subTopics = make(map[string]byte, 0)
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
func (mc mqttClient) subscribeMultiple(client mt.Client) error {
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
