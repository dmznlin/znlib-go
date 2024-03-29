// Package znlib
/******************************************************************************
  作者: dmzn@163.com 2024-01-19 09:09:05
  描述: mqtt辅助类,提供安全验证、消息缓存、异步处理等
******************************************************************************/
package znlib

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	mt "github.com/eclipse/paho.mqtt.golang"
	"reflect"
	"strings"
	"sync"
	"time"
)

const (
	mqttZipTail      = "!z.n_z-i+p"     //压缩数据尾部标识
	mqttZipTagLength = len(mqttZipTail) //压缩标识长度
)

// MqttCode 指令代码
type MqttCode = uint8

// MqttCommand 命令结构
type MqttCommand struct {
	Serial string   `json:"no"` //业务流水
	Sender string   `json:"sd"` //发送方
	Cmd    MqttCode `json:"cd"` //指令代码
	Ext    MqttCode `json:"et"` //指令扩展
	Data   string   `json:"dt"` //指令数据
	Verify string   `json:"vr"` //校验码

	VerifyUse bool          `json:"-"` //无需验证码
	Topic     string        `json:"-"` //主题
	Timeout   time.Duration `json:"-"` //超时时长
}

// MqttHandler 消息处理句柄
type MqttHandler = func(cmd *MqttCommand) error

// MqttUtils 辅助类
type mqttUtils struct {
	enabled     bool          //辅助类已启用
	msgKey      string        //消息加密密钥
	msgVerify   bool          //消息需要验证
	msgZip      bool          //数据需要压缩
	msgZipLen   int           //数据长度:超过该长度才会压缩
	msgFun      []MqttHandler //消息处理链
	msgDone     chan struct{} //消息待处理信号
	msgDelay    time.Duration //消息处理延迟
	workerNum   int           //工作对象个数
	workerGroup *RoutineGroup //工作对象组

	msgRecv *CircularQueue[*MqttCommand] //已接收消息队列
	msgIdle *CircularQueue[*MqttCommand] //空闲数据队列
}

// MqttUtils 封装协议辅助函数
var MqttUtils = &mqttUtils{
	enabled:     false,
	msgKey:      "mqtt.key",
	msgVerify:   false,
	msgZip:      false,
	msgZipLen:   100,
	msgFun:      make([]MqttHandler, 0),
	msgDone:     nil,
	msgDelay:    1 * time.Second,
	workerNum:   2,
	workerGroup: NewRoutineGroup(),
	msgRecv:     NewCircularQueue[*MqttCommand](Circular_FIFO, 0, true),
	msgIdle:     NewCircularQueue[*MqttCommand](Circular_FIFO, 0, true),
}

//  ---------------------------------------------------------------------------

// init 2024-02-06 18:40:20
/*
 描述: 单元初始化
*/
func init() {
	Mqtt.RegisterEventHandler(func(event MqttEvent) {
		switch event {
		case MqttEventServiceStop:
			MqttUtils.stopWorkers()
			//停止辅助类工作对象
		default:
			//do nothing
		}
	})
}

// RegisterHandler 2024-01-16 14:33:49
/*
 参数: fn,句柄
 描述: 注册一个消息处理句柄
*/
func (mu *mqttUtils) RegisterHandler(fn MqttHandler) {
	if IsNil(fn) {
		return
	}

	Application.SyncLock.Lock()
	defer Application.SyncLock.Unlock()

	pFun := reflect.ValueOf(fn)
	for _, v := range mu.msgFun {
		if reflect.ValueOf(v).Pointer() == pFun.Pointer() { //重复注册
			return
		}
	}

	mu.msgFun = append(mu.msgFun, fn)
	//注册
}

// NewCommand 2024-01-14 16:33:15
/*
 描述: 初始化一个命令
*/
func (mu *mqttUtils) NewCommand() (cmd *MqttCommand) {
	cmd = &MqttCommand{}
	cmd.Reset(false)
	return cmd
}

// Reset 2024-02-06 10:01:37
/*
 参数: init,初始化
 描述: 重置指令参数
*/
func (mc *MqttCommand) Reset(init ...bool) {
	if init == nil || init[0] { //清空数据
		*mc = MqttCommand{}
	}

	no, _ := SerialID.NextStr(false)
	//业务序列号

	mc.Serial = fmt.Sprintf("%s_%s", mc.Sender, no)
	mc.Sender = Mqtt.Options.ClientID
	mc.VerifyUse = MqttUtils.msgVerify
}

// GetVerify 2024-01-14 15:48:28
/*
 描述: 计算验证信息
*/
func (mc *MqttCommand) GetVerify() string {
	mc.Verify = MqttUtils.msgKey
	data, err := json.Marshal(mc)
	//占位后生成json字符串

	if err != nil {
		Error(": GetVerify", LogFields{"err": err})
		return ""
	}

	mc.Verify = fmt.Sprintf("%x", md5.Sum(data))
	return mc.Verify
}

// SendCommand 2024-01-14 16:19:13
/*
 参数: topic,主题
 参数: qos,发送级别
 描述: 将mc发送至topic
*/
func (mc *MqttCommand) SendCommand(topic string, qos MqttQos) *MqttCommand {
	if mc.VerifyUse {
		if mc.Verify == "" {
			mc.GetVerify()
		}
	} else {
		mc.Verify = ""
	}

	caller := "znlib.mqttutils.SendCommand"
	data, err := json.Marshal(mc)
	if err != nil {
		ErrorCaller(err, caller)
		return nil
	}

	if MqttUtils.msgZip && len(data) >= MqttUtils.msgZipLen { //压缩数据
		data, err = NewZipper().ZipData(data)
		if err != nil {
			ErrorCaller(err, caller)
			return nil
		}

		data = append(data, []byte(mqttZipTail)...)
		//添加压缩标识
	}

	var waiter *mqttWaiter
	if mc.Timeout > 0 {
		waiter = mc.waiter()
		defer waiter.Clear()
	}

	Mqtt.Publish(topic, qos, [][]byte{data})
	//发送数据
	if waiter != nil {
		return waiter.WaitFor(false)
		//等待命令返回
	}

	return nil
}

// SwitchUpDown 2024-01-19 16:13:45
/*
 参数: up,上行
 描述: 切换mc的上、下行主题
*/
func (mc *MqttCommand) SwitchUpDown(up bool) {
	mc.Sender = Mqtt.Options.ClientID
	//更新发送者

	if up {
		mc.Topic = strings.ReplaceAll(mc.Topic, "/down/", "/up/")
	} else {
		mc.Topic = strings.ReplaceAll(mc.Topic, "/up/", "/down/")
	}
}

//  ---------------------------------------------------------------------------

// mqttWaiter 异步转同步等待对象
type mqttWaiter struct {
	used bool             //是否使用
	sign chan MqttCommand //信号
	data *MqttCommand     //数据
}

var (
	// 读写同步锁定
	waitSync = &sync.RWMutex{}

	// 等待对象池
	waiterBuf = make([]*mqttWaiter, 0)
	// 等待对象索引

	waiters = make(map[string]*mqttWaiter)
)

// newWaiter 2024-01-16 13:36:09
/*
 参数: mc,mqtt指令
 描述: 生成mc的异步等待对象
*/
func (mc *MqttCommand) waiter() (mw *mqttWaiter) {
	waitSync.Lock()
	defer waitSync.Unlock()

	mw = nil
	for _, v := range waiterBuf {
		if !v.used { //空闲
			mw = v
			break
		}
	}

	if mw == nil { //新增
		mw = &mqttWaiter{
			sign: make(chan MqttCommand, 1),
			//缓冲为1,避免超时退出后,消息达到被阻塞
		}

		waiterBuf = append(waiterBuf, mw)
		Info(fmt.Sprintf("znlib.mqttutils.waiter: buffer size %d", len(waiterBuf)))
	} else {
	loop:
		for { //清空原有信号
			select {
			case <-mw.sign: //测试是否有信号
				//do nothing
			default:
				break loop
			}
		}
	}

	mw.used = true
	mw.data = mc
	waiters[mc.Serial] = mw
	return
}

// WaitFor 2024-01-16 13:50:35
/*
 参数: clear,自动清理
 描述: true,数据正常;false,等待超时
*/
func (mw *mqttWaiter) WaitFor(clear bool) (mc *MqttCommand) {
	if clear { //等待后执行清理
		defer mw.Clear()
	}

	tk := time.Tick(mw.data.Timeout)
	select {
	case <-tk: //超时
		return nil
	case cmd, ok := <-mw.sign: //数据到达
		if ok {
			return &cmd
		}
	}

	return nil
}

// Clear 2024-01-17 10:44:47
/*
 描述: 清理资源
*/
func (mw *mqttWaiter) Clear() {
	waitSync.Lock()
	defer waitSync.Unlock()

	if mw.used {
		delete(waiters, mw.data.Serial) //删除索引
		mw.used = false                 //置为空闲
		mw.data = nil                   //置空数据
	}
}

//  ---------------------------------------------------------------------------

// onMessge 2024-01-14 13:59:47
/*
 参数: cli,mqtt链路
 参数: msg,mqtt消息内容
 描述: 接收broker下发的消息
*/
func (mu *mqttUtils) onMessge(cli mt.Client, msg mt.Message) {
	if !mu.enabled { //mqtt has stopped
		return
	}

	newMsg := msg.Payload()
	newLen := len(newMsg)
	if newLen < 2 { //空json长度: {}
		return
	}

	caller := "znlib.mqttutils.OnMessge"
	defer DeferHandle(false, caller)
	//捕捉异常

	if newMsg[newLen-1] != '}' { //非json结构
		tagStart := newLen - mqttZipTagLength
		if tagStart > 0 && string(newMsg[tagStart:]) == mqttZipTail { //判断是否压缩
			var err error
			newMsg = newMsg[0:tagStart] //去掉标识
			newMsg, err = NewZipper().UnzipData(newMsg)

			if err != nil {
				ErrorCaller(err, caller)
				return
			}

			if Application.IsDebug {
				Info(fmt.Sprintf(caller+".Unzip: %d -> %d", newLen, len(newMsg)))
				//打印压缩信息
			}
		}
	}

	//  ---------------------------------------------------------------------------
	if Application.IsDebug {
		Info(fmt.Sprintf(caller+": %s,%s", msg.Topic(), newMsg))
		//msg content

		rNum, rCap := mu.msgRecv.Size()
		iNum, iCap := mu.msgIdle.Size()
		Info(fmt.Sprintf(caller+": recv[%d,%d],idle[%d,%d]", rNum, rCap, iNum, iCap))
	}

	cmd, ok := mu.msgIdle.Pop(nil)
	if !ok { //无空闲则新增
		cmd = &MqttCommand{}
	}

	err := json.Unmarshal(newMsg, cmd)
	if err != nil {
		pe := mu.msgIdle.Push(cmd) //回收
		if pe != nil {
			err = ErrorMsg(err, pe.Error())
		}

		ErrorCaller(err, caller+".json.Unmarshal")
		return
	}

	if cmd.Sender == Mqtt.Options.ClientID { //收到自己发送的消息,直接抛弃
		pe := mu.msgIdle.Push(cmd) //回收
		if pe != nil {
			ErrorCaller(pe, caller)
		}

		Info(caller + ": receive message from self")
		return
	}

	if mu.msgVerify { //需验证
		str := cmd.Verify
		if str == "" || cmd.GetVerify() != str { //验证失败
			pe := mu.msgIdle.Push(cmd) //回收
			if pe != nil {
				ErrorCaller(pe, caller)
			}

			ErrorCaller("mqtt message verify failure", caller)
			return
		}
	}

	cmd.Topic = msg.Topic()
	cmd.Timeout = 0
	//整理数据

	pe := mu.msgRecv.Push(cmd)
	//放入缓冲等待处理
	if pe != nil {
		ErrorCaller(pe, caller)
	}

	func() { //唤醒等待返回的消息
		waitSync.RLock()
		waitSync.RUnlock()

		waiter, ok := waiters[cmd.Serial]
		if ok {
			waiter.sign <- *cmd
			//返回数据
		}
	}()

	start := time.Now()
	mu.msgDone <- struct{}{}
	//发送信号,让work处理消息
	delay := time.Now().Sub(start)

	if delay > mu.msgDelay { //延迟超过上限
		Warn(fmt.Sprintf(caller+": worker delay %dms", delay/time.Millisecond))
	}
}

// addWorkers 2024-01-18 10:40:57
/*
 描述: 添加工作对象
*/
func (mu *mqttUtils) addWorkers() {
	if mu.enabled { //已添加
		return
	}

	worker := func(args ...interface{}) {
		workerID, _ := args[0].(int)
		caller := fmt.Sprintf("znlib.mqttutils.worker[%d]", workerID)
		//对象标识
		var isOK bool
		var cmd *MqttCommand

		for {
			select {
			case <-Application.Ctx.Done(): //主程序退出
				Info(caller + ": exit")
				return
			case _, ok := <-mu.msgDone: //跳出等待,处理消息
				if !ok { //通道关闭
					Info(caller + ": exit self")
					return
				}
			}

			hand := func() { //发布消息
				defer func() {
					_ = mu.msgIdle.Push(cmd)
					//放回空闲队列

					err := recover()
					if err != nil {
						ErrorCaller(err, caller)
						return
					}
				}()

				for _, v := range mu.msgFun { //将消息发布到处理链上
					err := v(cmd)
					if err != nil {
						ErrorCaller(err, caller)
					}
				}
			}

			cmd, isOK = mu.msgRecv.Pop(nil)
			//取最早一条消息
			for isOK {
				hand()
				cmd, isOK = mu.msgRecv.Pop(nil)
				//下一条消息
			}
		}
	}

	mu.enabled = true
	//启用辅助类
	mu.msgDone = make(chan struct{}, mu.workerNum*3)
	//按通道分配信号

	var id = 0
	for mu.workerNum > id { //创建工作routine
		id++ //从1开始
		mu.workerGroup.Run(worker, id)
	}
}

// stopWorkers 2024-01-29 18:55:39
/*
 描述: 停止工作对象
*/
func (mu *mqttUtils) stopWorkers() {
	if !mu.enabled { //未启用
		return
	}

	mu.enabled = false
	Info("znlib.mqttutils.stopWorkers: wait worker exit")

	close(mu.msgDone)
	mu.workerGroup.Wait()
	Info("znlib.mqttutils.stopWorkers: all worker has exit")
}
