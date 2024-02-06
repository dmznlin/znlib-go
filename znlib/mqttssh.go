// Package znlib
/******************************************************************************
  作者: dmzn@163.com 2024-02-01 18:30:04
  描述: 基于mqtt的ssh,用于远程管理

备注:
  *.ssh先决条件:
	1.使用 Mqtt.StartWithUtils 启动mqtt服务,这样mqtt被格式话为MqttCommand.
	2.配置 config.ini => mqttSSH 小节.
	3.channel中需要有一个带 $id 变量的通道,用于 远程<=>本地 传递数据.
  *.MqttCommand.Data需要以"\n"结尾,命令才会被执行.
******************************************************************************/
package znlib

import (
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"os"
	"strings"
	"time"
)

const (
	mqttSSHConn = "conn" //连接
	mqttSSHExit = "exit" //断开
)

// sshClient ssh配置
type sshClient struct {
	enabled  bool   //ssh启用
	host     string //ip:port
	user     string //用户
	password string //密码

	connTimeout time.Duration      //连接超时(毫秒)
	exitTimeout time.Duration      //超时退出(秒)
	caller      string             //远程标识
	sshCmd      uint8              //ssh指令代码
	channel     map[string]MqttQos //返回时的通道列表

	client      *ssh.Client    //客户端
	stdinPipe   io.WriteCloser //输入
	session     *ssh.Session   //会话
	workerGroup *RoutineGroup  //工作对象组

	fromMqtt chan struct{}          //有来自mqtt的数据
	dataRecv *CircularQueue[string] //已接收消息队列
}

// MqttSSH ssh客户端
var MqttSSH = &sshClient{
	enabled:     false,
	host:        "127.0.0.1:22",
	user:        "root",
	password:    "",
	connTimeout: 3000,
	exitTimeout: 60,
	caller:      "",
	sshCmd:      0,
	channel:     make(map[string]MqttQos),
	client:      nil,
	stdinPipe:   nil,
	session:     nil,
	workerGroup: NewRoutineGroup(),
	fromMqtt:    nil,
	dataRecv:    NewCircularQueue[string](Circular_FIFO, 0, true),
}

// init_mqttSSH 2024-02-01 20:15:37
/*
 描述: 初始化ssh
*/
func init_mqttSSH() {
	if MqttSSH.enabled {
		Mqtt.RegisterEventHandler(func(event MqttEvent) {
			switch event {
			case MqttEventServiceStop:
				MqttSSH.closeSSHConn()
				//注册关闭ssh
			}
		})

		MqttUtils.RegisterHandler(doMqttSSHCommand)
		//注册ssh消息通道

		Application.RegisterExitHandler(func() {
			MqttSSH.closeSSHConn()
			//注册关闭ssh
		})
	}
}

// doMqttSSHCommand 2024-02-01 20:20:44
/*
 参数: cmd,ssh指令
 描述: 处理远程发送的ssh指令
*/
func doMqttSSHCommand(cmd *MqttCommand) (err error) {
	if cmd.Cmd != MqttSSH.sshCmd { //非ssh指令
		if Application.IsDebug {
			Info("znlib.mqttssh.doMqttSSHCommand: invalid ssh command")
		}

		return nil
	}

	if Application.IsDebug { //debug log
		Info("znlib.mqttssh.doMqttSSHCommand: " + cmd.Data)
	}

	switch cmd.Data {
	case mqttSSHExit: //退出
		err = MqttSSH.closeSSHConn()
		if err != nil {
			MqttSSH.sendToMqtt(cmd.Sender, []byte(err.Error()))
		}
		return err
	case mqttSSHConn: //连接
		if MqttSSH.client == nil { //创建ssh
			err = MqttSSH.newSSHConn()
			if err != nil {
				MqttSSH.sendToMqtt(cmd.Sender, []byte(err.Error()))
				return err
			}

			MqttSSH.caller = cmd.Sender
			//关联远程标识
		}
	default: //数据
		if MqttSSH.client != nil {
			MqttSSH.dataRecv.Push(cmd.Data)
			MqttSSH.fromMqtt <- struct{}{}
			//数据到达,告知worker处理
		}
	}

	return nil
}

// sendToMqtt 2024-02-04 13:39:17
/*
 参数: receiver,接收者标识
 参数: data,数据
 描述: 将data发送至receiver
*/
func (sc *sshClient) sendToMqtt(receiver string, data []byte) {
	cmd := MqttUtils.NewCommand()
	cmd.Cmd = sc.sshCmd
	cmd.Data = string(data)
	cmd.GetVerify()

	for k, v := range sc.channel {
		k = strings.ReplaceAll(k, "$id", receiver)
		cmd.SendCommand(k, v)
	}
}

// Connect 2024-02-06 19:10:17
/*
 参数: client,mqtt.clientID
 描述: 向client发送连接请求
*/
func (sc *sshClient) Connect(client string) {
	sc.sendToMqtt(client, []byte(mqttSSHConn))
}

// Disconnect 2024-02-06 19:11:16
/*
 参数: client,mqtt.clientID
 描述: 向client发送断开请求
*/
func (sc *sshClient) Disconnect(client string) {
	sc.sendToMqtt(client, []byte(mqttSSHExit))
}

// newSSHConn 2024-02-01 20:38:05
/*
 描述: 新建ssh连接
*/
func (sc *sshClient) newSSHConn() (err error) {
	caller := "znlib.mqttssh.newSSHConn"
	defer DeferHandle(false, caller, func(e error) {
		var ec error
		if err != nil {
			ec = sc.closeSSHConn()
		}

		err = ErrorJoin(err, e, ec)
	})

	config := &ssh.ClientConfig{
		Timeout:         sc.connTimeout * time.Millisecond,
		User:            sc.user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //忽略know_hosts检查
		Auth:            []ssh.AuthMethod{ssh.Password(sc.password)},
	}

	sc.client, err = ssh.Dial("tcp", sc.host, config)
	if err != nil {
		ErrorCaller(err, caller)
		return err
	}

	sc.fromMqtt = make(chan struct{}, 5)
	//数据接收通道

	sc.workerGroup.Run(func(args ...interface{}) {
		sc.startSSHWorker()
	})

	return nil
}

// closeSSHConn 2024-02-03 13:52:51
/*
 描述: 关闭连接
*/
func (sc *sshClient) closeSSHConn() (err error) {
	if sc.client == nil { //没有连接
		return nil
	}

	if !IsNil(sc.fromMqtt) { //关闭接收通道
		close(sc.fromMqtt)
	}

	Info("znlib.mqttssh.closeSSHConn: wait worker exit")
	sc.workerGroup.Wait()
	//等待工作对象退出
	Info("znlib.mqttssh.closeSSHConn: all worker has exit")

	sc.caller = ""
	sc.client = nil
	sc.fromMqtt = nil
	//重置对象状态
	return nil
}

// sshWorker 2024-02-03 18:00:55
/*
   描述: ssh工作线程
*/
func (sc *sshClient) startSSHWorker() {
	caller := "znlib.mqttssh.sshWorker"
	var err error

	defer func() {
		sc.client.Close()
		//最后关闭链路

		sc.caller = ""
		sc.client = nil
		sc.fromMqtt = nil
	}()

	sc.session, err = sc.client.NewSession()
	if err != nil {
		ErrorCaller(err, caller)
		return
	}

	defer sc.session.Close()
	//关闭此次会话

	sc.stdinPipe, err = sc.session.StdinPipe()
	if err != nil {
		ErrorCaller(err, caller)
		return
	}

	defer sc.stdinPipe.Close()
	//关闭会话输入管道

	sc.session.Stdout = sc
	sc.session.Stderr = sc
	//关联输出

	/*
		inFd := int(os.Stdin.Fd())
		state, err := terminal.MakeRaw(inFd)
		if err != nil {
			return
		}
		defer terminal.Restore(inFd, state)
	*/
	/*
		在终端处于 Cooked 模式时，当你输入一些字符后，默认是被当前终端 cache 住的，在你敲了回车之前这些文本都在 cache 中，这样允许应用程序做
		一些处理，比如捕获 Cntl-D 等按键，这时候就会出现敲回车后本地终端帮你打印了一下，导致出现类似回显的效果；当设置终端为 raw 模式后，所有的
		输入将不被 cache，而是发送到应用程序.
	*/

	termW, termH, err := terminal.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		termH = 24
		termW = 80
	}

	termType := os.Getenv("TERM")
	if termType == "" {
		termType = "xterm-256color"
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // enable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	termH = 24
	termW = 80
	if err = sc.session.RequestPty(termType, termH, termW, modes); err != nil {
		ErrorCaller(err, caller)
		return
	}

	if err = sc.session.Shell(); err != nil {
		ErrorCaller(err, caller)
		return
	}

	sc.workerGroup.Run(func(args ...interface{}) {
		err = sc.session.Wait()
		if err != nil {
			ErrorCaller(err, "znlib.mqttssh.session.Wait")
		}
	})

	for {
		select {
		case <-Application.Ctx.Done(): //主程序退出
			Info(caller + ": exit")
			return
		case _, ok := <-sc.fromMqtt: //远程指令
			if !ok { //通道关闭
				Info(caller + ": exit self")
				return
			}

			data, ok := sc.dataRecv.Pop("")
			if ok {
				if Application.IsDebug {
					Info(caller + ": " + data)
				}

				if _, err := sc.stdinPipe.Write([]byte(data)); err != nil { //写入ssh
					ErrorCaller(err, caller)
				}
			}
		}
	}
}

// Read 2024-02-03 17:03:42
/*
   参数: data,ssh回显数据
   描述: 将 data 发送至 mqtt
*/
func (sc *sshClient) Write(data []byte) (n int, err error) {
	sc.sendToMqtt(sc.caller, data)
	return len(data), nil
}
