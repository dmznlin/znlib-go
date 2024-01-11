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
			if len(Mqtt.subTopics) > 0 {
				token := client.SubscribeMultiple(Mqtt.subTopics, nil)
				if token.Wait() && token.Error() == nil {
					for k, v := range Mqtt.subTopics {
						Info(fmt.Sprintf("znlib.mqtt.subscribe: %s,%d", k, v))
					}
				} else {
					Error("znlib.mqtt.subscribe: ", LogFields{"err": token.Error()})
				}
			}
		})
	}

	if Mqtt.Options.OnConnectionLost == nil {
		Mqtt.Options.SetConnectionLostHandler(func(client mt.Client, err error) {
			Error("znlib.mqtt.lostconnect: " + err.Error())
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
		Error("znlib.mqtt.connect_broker", LogFields{"err": token.Error()})
	}

	return token.Error()
}

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
			Error("znlib.mqtt.unsubscribe", LogFields{"err": token.Error()})
		} else {
			Info(fmt.Sprintf("znlib.mqtt.unsubscribe: %v", topics))
		}
	}

	mc.client.Disconnect(500)
	//断开链路
}

// Publish 2024-01-10 15:32:26
/*
 参数: topic,主题
 参数: msg,消息内容
 描述: 向topic发布msg消息
*/
func (mc *mqttClient) Publish(topic string, msg []string) {
	pub := func(tp string) {
		qos, ok := mc.pubTopics[tp]
		if !ok {
			Info(fmt.Sprintf("znlib.mqtt.publish: topic %s isn't exists", topic))
			return
		}

		for _, v := range msg {
			token := mc.client.Publish(tp, qos, false, v)
			if token.Wait() && token.Error() != nil {
				Error("znlib.mqtt.publish", LogFields{"err": token.Error()})
			}
		}
	}

	topic = StrTrim(topic)
	if topic == "" {
		for k := range mc.pubTopics {
			pub(k)
		}
	} else {
		pub(topic)
		//自定义主题
	}
}
