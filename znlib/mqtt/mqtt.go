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
	"time"

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
	Event        = byte                         //event 代码
	EventHandler = func(mc *Utils, event Event) //event 事件
)

const (
	EventConnected    Event = iota //连接broker 成功
	EventDisconnect                //broker 连接断开
	EventReConnect                 //broker 自动重连
	EventServiceStart              //mqtt 服务启动
	EventServiceStop               //mqtt 服务停止
)

// Utils 辅助类
type Utils struct {
	Client    mt.Client
	Options   *mt.ClientOptions //选项
	SubTopics map[string]Qos    //订阅主题
	PubTopics map[string]Qos    //发布主题

	HintInfo     bool           //打印提示信息
	KeyEncrypted bool           //密码已加密
	events       []EventHandler //事件处理列表
	waitePub     *Waiter[bool]  //等待注册完成
}

// Client 客户端
var Client = &Utils{
	Options:   mt.NewClientOptions(),
	SubTopics: make(map[string]Qos),
	PubTopics: make(map[string]Qos),

	events:       nil,
	waitePub:     nil,
	HintInfo:     true,
	KeyEncrypted: true,
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
		err := Client.ApplyConfig(&cfg.Mqtt)
		if err != nil {
			ErrorCaller(err, "znlib.mqtt.init")
		}
	})
}

// hintMsg 2026-03-31 18:37:59
/*
 参数: msg,提示信息
 描述: 按需求打印提示信息
*/
func (mc *Utils) hintMsg(msg string, caller ...string) {
	if mc.HintInfo {
		if len(caller) < 1 {
			Info(msg)
		} else {
			ErrorCaller(msg, caller[0])
		}
	}
}

// ApplyConfig 2026-03-09 15:04:52
/*
 参数: cfg,mqtt配置
 描述:
*/
func (mc *Utils) ApplyConfig(cfg *MqttConfig) error {
	if !cfg.Enable {
		return nil
	}

	if len(cfg.Broker) < 1 {
		return fmt.Errorf("mqtt.broker is empty")
	}

	if cfg.Tls.Used {
		cfg.Tls.CA = FixPathVar(cfg.Tls.CA)
		if !FileExists(cfg.Tls.CA, false) {
			return fmt.Errorf("mqtt.Tls.ca is miss")
		}

		cfg.Tls.Key = FixPathVar(cfg.Tls.Key)
		if !FileExists(cfg.Tls.Key, false) {
			return fmt.Errorf("mqtt.Tls.key is miss")
		}

		cfg.Tls.Cert = FixPathVar(cfg.Tls.Cert)
		if !FileExists(cfg.Tls.Cert, false) {
			return fmt.Errorf("mqtt.Tls.cert is miss")
		}

		rootCA, err := os.ReadFile(cfg.Tls.CA)
		if err != nil {
			return fmt.Errorf("mqtt.ReadFile: %v", err)
		}

		cp := x509.NewCertPool()
		if !cp.AppendCertsFromPEM(rootCA) {
			return fmt.Errorf("mqtt.Tls.ca load error")
		}

		cert, err := tls.LoadX509KeyPair(cfg.Tls.Cert, cfg.Tls.Key)
		if err != nil {
			return fmt.Errorf("mqtt.LoadX509KeyPair: %v", err)
		}

		mc.Options.SetTLSConfig(&tls.Config{
			RootCAs:            cp,
			ClientAuth:         tls.NoClientCert,
			ClientCAs:          nil,
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{cert},
		})
	}

	if cfg.Password != "" && mc.KeyEncrypted { // broker 密码
		buf, err := NewEncrypter(EncryptDesEcb, []byte(DefaultEncryptKey)).Decrypt([]byte(cfg.Password), true)
		if err != nil {
			return fmt.Errorf("mqtt.pwd is invalid: %v", err)
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
		if len(v.Topic) > 0 {
			v.Topic = StrReplace(v.Topic, cfg.ClientID, "$id")
			//更新数据通道标识
			mc.SubTopics[v.Topic] = v.Qos
			//订阅主题
		}
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

	mc.Options.Servers = mc.Options.Servers[:0]
	//clear first
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

			mc.hintMsg("znlib.mqtt.connected: " + host)
			_ = mc.SubscribeMultiple(client)
			//连接成功后,重新订阅主题

			if mc.waitePub != nil {
				mc.waitePub.Wakeup(new(bool))
			}

			mc.eventAction(EventConnected)
			//触发已连接事件
		})
	}

	if mc.Options.OnConnectionLost == nil {
		mc.Options.SetConnectionLostHandler(func(client mt.Client, err error) {
			mc.hintMsg(err.Error(), "znlib.mqtt.lostconnect")
			mc.eventAction(EventDisconnect) //触发断开事件
		})
	}

	if mc.Options.OnReconnecting == nil {
		mc.Options.SetReconnectingHandler(func(client mt.Client, options *mt.ClientOptions) {
			mc.hintMsg("znlib.mqtt: reconnect broker")
			mc.eventAction(EventReConnect) //触发重连事件
		})
	}

	return nil
}

// Start 2024-01-11 08:24:20
/*
 参数: msgHandler,消息处理函数
 参数: waitPub,等待订阅完成
 描述: 启动mqtt服务
*/
func (mc *Utils) Start(msgHandler mt.MessageHandler, waitPub ...time.Duration) error {
	if mc.Client != nil {
		return nil
	}

	if msgHandler != nil {
		mc.Options.SetDefaultPublishHandler(msgHandler)
		//默认消息处理句柄
	}

	if len(waitPub) > 0 { //等待订阅
		if mc.waitePub == nil {
			mc.waitePub = NewWaiter[bool](nil)
		}

		mc.waitePub.Reset()
		//清空等待信号
	}

	mc.Client = mt.NewClient(mc.Options)
	//创建链路
	token := mc.Client.Connect()
	//连接 broker

	if token.Wait() && token.Error() != nil {
		return token.Error()
	}

	if len(waitPub) > 0 {
		_, ok := mc.waitePub.WaitFor(waitPub[0])
		if !ok {
			return fmt.Errorf("znlib.mqtt.Start:wait publish timeout")
		}
	}

	mc.eventAction(EventServiceStart)
	//触发服务启动事件
	return nil
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

// isConnected 2026-04-01 10:50:19
/*
 参数: cli,链路
 描述: 检测 cli 是否已连接
*/
func (mc *Utils) isConnected(cli mt.Client) error {
	if cli == nil || !cli.IsConnected() {
		return fmt.Errorf("client is not connected")
	}

	return nil
}

// Publish 2026-02-27 14:43:07
/*
 参数: topic,主题
 参数: qos,送达级别
 参数: msg,消息
 描述: 向topic发布msg消息
*/
func (mc *Utils) Publish(topic string, qos Qos, msg []byte) error {
	defer DeferHandle(false, "znlib.mqtt.publish")
	//捕捉网络异常

	if err := mc.isConnected(mc.Client); err != nil {
		return err
	}

	pub := func() error {
		token := mc.Client.Publish(topic, qos, false, msg)
		if token.Wait() && token.Error() != nil {
			mc.hintMsg(token.Error().Error(), "znlib.mqtt.publish")
			return token.Error()
		}

		return nil
	}

	useCfg := qos == QosNone
	//使用配置 qos

	if topic == "" { //发布至所有主题
		var q Qos
		for topic, q = range mc.PubTopics {
			if useCfg {
				qos = q
			}

			if err := pub(); err != nil {
				return err
			}
		}

		return nil
	}

	if useCfg {
		q, ok := mc.PubTopics[topic]
		if ok {
			qos = q
		} else {
			qos = Qos0
		}
	}

	return pub()
	//自定义主题
}

// Subscribe 2024-01-14 14:52:25
/*
 参数: topic,主题
 参数: qos,模式
 参数: ctl,控制参数(1.true:重置订阅;2.true:立即订阅)
 描述: 新增订阅topic主题
*/
func (mc *Utils) Subscribe(topic string, qos Qos, ctl ...bool) error {
	if err := mc.isConnected(mc.Client); err != nil {
		return err
	}

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
		token := mc.Client.Subscribe(topic, qos, nil) //开始订阅
		if token.Wait() && token.Error() != nil {
			mc.hintMsg(token.Error().Error(), "znlib.mqtt.subscribe")
			return token.Error()
		}
	}

	return nil
}

// SubscribeMultiple 2024-01-14 15:22:11
/*
 参数: Client,链路
 描述: 订阅主题列表
*/
func (mc *Utils) SubscribeMultiple(client mt.Client) error {
	if len(mc.SubTopics) < 1 {
		return nil
	}

	if client == nil {
		client = mc.Client
		//use default
	}

	if err := mc.isConnected(client); err != nil {
		return err
	}

	token := client.SubscribeMultiple(mc.SubTopics, nil)
	if token.Wait() && token.Error() != nil {
		mc.hintMsg(token.Error().Error(), "znlib.mqtt.subMultiple")
		return token.Error()
	}

	mc.hintMsg(fmt.Sprintf("znlib.mqtt.subscribe: %v", mc.SubTopics))
	return nil
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

	if err := mc.isConnected(client); err != nil {
		return err
	}

	idx := len(mc.SubTopics)
	if idx < 1 {
		return nil
	}

	topics := make([]string, idx)
	idx = 0
	for k := range mc.SubTopics { //退订所有主题
		topics[idx] = k
		idx++
	}

	token := client.Unsubscribe(topics...)
	if token.Wait() && token.Error() != nil {
		mc.hintMsg(token.Error().Error(), "znlib.mqtt.unsubscribe")
		return token.Error()
	}

	mc.hintMsg(fmt.Sprintf("znlib.mqtt.unsubscribe: %v", topics))
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

	if IsNil(mc.events) {
		mc.events = make([]EventHandler, 0, 2)
	}

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
	if IsNil(mc.events) {
		return
	}

	defer DeferHandle(false, "znlib.mqtt.eventAction")
	for _, do := range mc.events {
		do(mc, event)
	}
}
