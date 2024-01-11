// Package znlib
/******************************************************************************
作者: dmzn@163.com 2022-05-30 13:35:38
描述: 系统常量、变量、函数等
******************************************************************************/
package znlib

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

// application相关属性
type application struct {
	ExeName    string //exe full
	ExePath    string //exe所在路径
	LogPath    string //日志目录
	ConfigFile string //主配置文件
	ConfigDB   string //数据库配置
	PathSymbol string //路径分隔符

	IsWindows bool   //win
	IsLinux   bool   //linux
	HostName  string //主机名称

	Ctx      context.Context //全局上下文
	SyncLock sync.RWMutex    //全局同步锁
}

var (
	//Application 全局对象
	Application application

	//AppFile application full name
	AppFile = os.Args[0]

	//AppPath application所在路径
	AppPath = filepath.Dir(AppFile)

	//PathSeparator 路径分隔符: / or \\
	PathSeparator = "/"

	//cancelFunc 取消函数
	cancelFunc context.CancelFunc
	//cancelCall 执行取消
	cancelCall sync.Once
	//cancelExtend 退出操作列表
	cancelExtend = make([]func(), 0)
)

// OSName 2022-05-30 13:14:24
/*
 描述: 获取当前系统名称
*/
func OSName() string {
	return runtime.GOOS
}

// FileExists 2022-05-30 13:24:34
/*
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

// FixPath 2022-05-30 13:08:42
/*
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

// FixPathVar 2024-01-10 11:30:42
/*
 参数: dir,目录或文件
 描述: 若dir中有$path变量,则替换为AppPath.并统一路径分隔符
*/
func FixPathVar(dir string) string {
	var str string
	if Application.IsWindows {
		str = StrReplace(dir, `\`, "/", "//")
	} else {
		str = StrReplace(dir, "/", `\`, `\\`)
	}

	return StrReplace(str, Application.ExePath, "$path\\", "$path/", "$path")
}

// MakeDir 2022-05-30 13:09:23
/*
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

// DeferHandle 2022-05-30 13:12:31
/*
 参数: throw,重新抛出异常
 参数: caller,调用者名称
 参数: cb,回调函数
 描述: 用于defer默认调用
*/
func DeferHandle(throw bool, caller string, cb ...func(err any)) {
	err := recover()
	if err != nil {
		if caller == "" {
			Error(err)
		} else {
			Error(err, LogFields{"caller": caller})
		}
	}

	for _, fn := range cb {
		fn(err)
	}

	if throw { //re-panic
		panic(err)
	}
}

// ErrorPanic 2022-07-07 12:40:42
/*
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

// ErrorMsg 2022-08-12 15:53:48
/*
 参数: err,异常
 参数: msg,消息
 参数: newone,是否自动生成
 描述: 生成一个携带err+msg的异常对象
*/
func ErrorMsg(err error, msg string, newone ...bool) error {
	if err == nil {
		if newone != nil && newone[0] {
			return errors.New(msg)
		} else {
			return nil
		}
	} else {
		return errors.WithMessage(err, msg)
	}
}

// TryFinal 模拟delhpi的try...finally机制
type TryFinal struct {
	Try     func() (err error) //业务函数
	Finally func()             //强制执行(一定执行)函数
	Except  func(err error)    //异常处理函数
}

// Run 2022-07-20 13:03:01
/*
 返回: error,异常
 描述: 执行业务,返回false时应该退出业务.
*/
func (tf TryFinal) Run() (err error) {
	if tf.Finally != nil { //run last
		defer tf.Finally()
	}

	defer func() {
		if errAny := recover(); errAny != nil { //run after panic
			e, ok := errAny.(error)
			if ok {
				err = e
			} else {
				err = errors.New(fmt.Sprintf("%s", errAny))
			}
		}

		if err != nil {
			if tf.Except == nil {
				AddLog(logEror, err, true)
				//write log
			} else {
				tf.Except(err)
				//do except
			}
		}
	}()

	err = tf.Try() //执行业务
	return
}

// WaitSystemExit 2022-06-08 15:24:34
/*
 参数: cw,清理函数
 描述: 捕捉操作系统关闭信号,执行清理后退出
*/
func WaitSystemExit(cw ...func() error) {
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

	Application.Exit()
	Info("广播信号,执行routine退出.")

	for _, fn := range cw {
		if err := fn(); err != nil {
			Error(err)
		}
	}
	Info("清理工作完成,系统退出.")
}

// WaitFor 2024-01-11 09:07:22
/*
 参数: d,等待间隔
 参数: canExit,检测是否可以退出
 描述: 等待一段时间,如果canExit返回true则提前退出
*/
func WaitFor(d time.Duration, canExit func() bool) {
	itv := 20 * time.Millisecond
	//interval:最小时间间隔
	if d <= itv {
		time.Sleep(d)
		return
	}

	num := d / itv
	//计时次数:按最小间隔拆分
	if num > 100 {
		itv = d / 100
	}

	end := time.Now().Add(d)
	//结束时间
	for range time.Tick(itv) {
		if canExit != nil && canExit() { //外部退出
			return
		}

		if time.Now().After(end) { //计时结束
			return
		}
	}
}

//--------------------------------------------------------------------------------

// initApp 2022-05-30 14:01:55
/*
 描述: 初始化系统运行环境
*/
func initApp() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	//random

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
	}

	Application.Ctx, cancelFunc = context.WithCancel(context.Background())
	//全局取消context,用于控制routine退出
}

// OnExit 2024-01-11 21:50:25
/*
 参数: fn,函数
 描述: 注册fn函数,在系统退出时执行
*/
func (app *application) OnExit(fn func()) {
	app.SyncLock.Lock()
	defer app.SyncLock.Unlock()
	cancelExtend = append(cancelExtend, fn)
}

// Exit 2024-01-11 21:14:24
/*
 描述: 程序退出,广播消息给所有routine(需支持Ctx)
*/
func (app *application) Exit() {
	cancelCall.Do(func() {
		cancelFunc()
		//设置退出标记

		for _, fn := range cancelExtend { //执行退出操作
			fn()
		}
	})
}
