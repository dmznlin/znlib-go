### 名称: terminal
> 1.基于 xterm.js,在浏览器内提供访问 ssh 的能力 \
> 2.配合 **znlib/mqttssh** 可实现基于 mqtt 的 ssh

使用方法:
> 1.在 terminal 目录下运行 npm install,获取依赖包 \
> 2.在 terminal 目录下运行 build.bat,有效文件在 **dist/terminal** 目录中. \
> 3.将 **dist/terminal** 复制到项目中.

### go后端处理代码:
```go
package main

import (
	. "github.com/dmznlin/znlib-go/znlib"
	"github.com/olahol/melody"
	"net/http"
)

// 初始化znlib-go基础库
var _ = InitLib(func() {
	Application.IsDebug = true
	Application.ConfigFile = AppPath + "manager.xml"
	//开发模式
}, nil)

func main() {
	ws := melody.New() //用于WebSocket
	msg := make(chan string, 10)

	Mqtt.StartWithUtils(func(cmd *MqttCommand) error { //接收远程数据
		msg <- cmd.Data
		return nil
	})

	go func() { // 处理来自虚拟终端的消息
		for {
			select {
			case <-Application.Ctx.Done():
				return
			case data := <-msg:
				ws.Broadcast([]byte(data)) // 将数据发送给网页
				Info("remote:" + data)
			}
		}
	}()

	ws.HandleMessage(func(s *melody.Session, msg []byte) { // 处理来自WebSocket的消息
		nlen := len(msg)
		if nlen > 5 && string(msg[0:6]) == string([]byte{0, 1, 2, 3, 4, 5}) { //检查命令前缀(6字节)
			switch {
			case nlen >= 10 && string(msg[6:10]) == "conn": //连接指令
				MqttSSH.Connect("kt001")
			case nlen >= 12 && string(msg[6:12]) == "resize": //调整大小
				MqttSSH.SendData("kt001", MqttSSHResize, msg[12:])
			default:
				//do nothing
			}

			return
		}

		MqttSSH.SendData("kt001", MqttSSHCommon, msg)
		//普通数据
	})

	http.HandleFunc("/terminal", func(w http.ResponseWriter, r *http.Request) {
		ws.HandleRequest(w, r) //转交给melody处理
	})

	termFS := FixPathVar("$path/dist/terminal/")
	http.Handle("/ssh/", http.StripPrefix("/ssh", http.FileServer(http.Dir(termFS)))) // 设置静态文件服务

	go http.ListenAndServe("0.0.0.0:22333", nil)

	WaitSystemExit(func() error {
		Mqtt.Stop()
		return nil
	})
}

```