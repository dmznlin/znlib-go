/*Package znlib **************************************************************
作者: dmzn@163.com 2022-05-30 13:35:38
描述: 系统常量、变量、函数等
******************************************************************************/
package znlib

import (
	"fmt"
	"github.com/pkg/errors"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
)

//AppFile application full name
var AppFile = os.Args[0]

//AppPath application所在路径
var AppPath = filepath.Dir(AppFile)

//PathSeparator 路径分隔符: / or \\
var PathSeparator = "/"

//application相关属性
type application struct {
	ExeName    string //exe full
	ExePath    string //exe所在路径
	LogPath    string //日志目录
	ConfigFile string //主配置文件
	ConfigDB   string //数据库配置
	PathSymbol string //路径分隔符

	IsWindows bool     //win
	IsLinux   bool     //linux
	HostName  string   //主机名称
	HostIP    []string //主机IP

	SyncLock sync.RWMutex //全局同步锁
}

//Application 全局对象
var Application application

/*OSName 2022-05-30 13:14:24
  描述: 获取当前系统名称
*/
func OSName() string {
	return runtime.GOOS
}

/*FileExists 2022-05-30 13:24:34
  参数: file,路径
  参数: isDir,是否文件夹
  描述: 判断file是否能存在
*/
func FileExists(file string, isDir bool) bool {
	info, err := os.Stat(file)
	switch {
	case err == nil:
		if isDir {
			return info.IsDir()
		}
		return true
	case os.IsNotExist(err):
		return false
	default:
		return false
	}
}

/*FixPath 2022-05-30 13:08:42
  参数: dir,文件夹路径
  描述: 若dir末尾没有分隔符,则添加
*/
func FixPath(dir string) string {
	l := len(dir) - 1
	if l < 0 {
		return dir
	}

	if os.IsPathSeparator(dir[l]) {
		return dir
	} else {
		return dir + PathSeparator
	}
}

/*MakeDir 2022-05-30 13:09:23
  参数: 件夹路径
  描述: 创建dir目录
*/
func MakeDir(dir string) {
	defer DeferHandle(false, "znlib.MakeDir")
	err := os.MkdirAll(dir, 755)

	if err != nil {
		panic(err)
	}
}

//--------------------------------------------------------------------------------

//DeferHandleCallback 异常处理时的回调函数
type DeferHandleCallback = func(err any)

/*DeferHandle 2022-05-30 13:12:31
  参数: throw,重新抛出异常
  参数: caller,调用者名称
  参数: cb,回调函数
  描述: 用于defer默认调用
*/
func DeferHandle(throw bool, caller string, cb ...DeferHandleCallback) {
	err := recover()
	switch t := err.(type) {
	case nil:
		//no error
	case error:
		Error(caller + ": " + t.Error())
	default:
		Error(caller, LogFields{"data: ": t})
	}

	for _, f := range cb {
		f(err)
	}

	if throw { //re-panic
		panic(err)
	}
}

/*ErrorPanic 2022-07-07 12:40:42
  参数: err,异常
  参数: message,描述信息
  描述: 抛出带描述信息的异常
*/
func ErrorPanic(err error, message ...string) {
	if message == nil {
		panic(err)
	} else {
		for _, msg := range message {
			err = errors.WithMessage(err, msg)
		}

		panic(err)
	}
}

/*TryFinally 2022-07-20 13:03:01
  参数: try,业务函数
  参数: finally,强制执行(一定执行)函数
  参数: except,异常处理函数
  描述: 模拟delhpi的try...finally机制
*/
func TryFinally(try, finally func(), except ...func(err any)) (ok bool) {
	ok = false
	func() { //将try装箱
		defer func() {
			defer finally()
			//一定执行

			if err := recover(); err != nil {
				if except == nil {
					Error(fmt.Sprintf("%v", err))
					//write log
				} else {
					for _, exc := range except {
						exc(err)
					}
				}
			}
		}()

		try()
		//执行业务
		ok = true
	}()

	return
}

//ClearWorkOnExit 程序关闭时的清理工作
type ClearWorkOnExit = func() error

/*WaitSystemExit 2022-06-08 15:24:34
  参数: cw,清理函数
  描述: 捕捉操作系统关闭信号,执行清理后退出
*/
func WaitSystemExit(cw ...ClearWorkOnExit) {
	// 程序无法捕获信号 SIGKILL 和 SIGSTOP （终止和暂停进程），因此 os/signal 包对这两个信号无效。
	signals := []os.Signal{syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT}
	if Application.IsLinux {
		signals = append(signals,
			syscall.Signal(0x10), //syscall.SIGUSR1
			syscall.Signal(0x11), // syscall.SIGUSR2
		)
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signals...)

	s := <-ch //阻塞
	close(ch)
	Info("信号:" + s.String() + ",开始清理工作.")

	for i := range cw {
		if err := cw[i](); err != nil {
			Error(err.Error())
		}
	}
	Info("清理工作完成,系统退出.")
}

//--------------------------------------------------------------------------------

/*initApp 2022-05-30 14:01:55
  描述: 初始化
*/
func initApp() {
	hostName, err := os.Hostname()
	if err != nil {
		hostName = "unknown"
	}

	osName := OSName()
	if strings.EqualFold(osName, "windows") {
		PathSeparator = "\\"
	} else {
		PathSeparator = "/"
	}

	AppPath = FixPath(AppPath)
	//尾部添加路径分隔符

	Application = application{
		ExeName:    AppFile,
		ExePath:    AppPath,
		LogPath:    AppPath + "logs" + PathSeparator,
		ConfigFile: AppPath + "config.ini",
		ConfigDB:   AppPath + "db.ini",
		PathSymbol: PathSeparator,
		IsLinux:    strings.EqualFold(osName, "linux"),
		IsWindows:  strings.EqualFold(osName, "windows"),
		HostName:   hostName,
		HostIP:     make([]string, 0, 2),
	}

	addr, err := net.InterfaceAddrs()
	if err == nil {
		for _, val := range addr {
			if ipnet, ok := val.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					Application.HostIP = append(Application.HostIP, ipnet.IP.String())
				}
			}
		}
	}
}
