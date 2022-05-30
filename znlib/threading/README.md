## 代码来源

github.com/zeromicro/go-zero/core/threading

如何理解go-zero对Go中goroutine支持的并发组件
=================================

发布时间：2021-10-13 09:52:17 作者：iii

本篇内容主要讲解“如何理解go-zero对Go中goroutine支持的并发组件 ”，感兴趣的朋友不妨来看看。本文介绍的方法操作简单快捷，实用性强。下面就让小编来带大家学习“如何理解go-zero对Go中goroutine支持的并发组件
”吧!

### threading

虽然 `go func()` 已经很方便，但是有几个问题：

* 如果协程异常退出，无法追踪异常栈

* 某个异常请求触发panic，应该做故障隔离，而不是整个进程退出，容易被攻击

我们看看 `core/threading` 包提供了哪些额外选择：

```
func GoSafe(fn func()) {
	go RunSafe(fn)
}

func RunSafe(fn func()) {
	defer rescue.Recover()
	fn()
}

func Recover(cleanups ...func()) {
	for _, cleanup := range cleanups {
		cleanup()
	}

	if p := recover(); p != nil {
		logx.ErrorStack(p)
	}
}
```

#### GoSafe

`threading.GoSafe()` 就帮你解决了这个问题。开发者可以将自己在协程中需要完成逻辑，以闭包的方式传入，由 `GoSafe()` 内部 `go func()`；

当开发者的函数出现异常退出时，会在 `Recover()` 中打印异常栈，以便让开发者更快确定异常发生点和调用栈。

#### NewWorkerGroup

我们再看第二个：`WaitGroup`。日常开发，其实 `WaitGroup` 没什么好说的，你需要 `N` 个协程协作 ：`wg.Add(N)` ，等待全部协程完成任务：`wg.Wait()`
，同时完成一个任务需要手动 `wg.Done()`。

可以看的出来，在任务开始 -> 结束 -> 等待，整个过程需要开发者关注任务的状态然后手动修改状态。

`NewWorkerGroup` 就帮开发者减轻了负担，开发者只需要关注：

1. 任务逻辑【函数】
2. 任务数【`workers`】

然后启动 `WorkerGroup.Start()`，对应任务数就会启动：

```
func (wg WorkerGroup) Start() {
  // 包装了sync.WaitGroup
	group := NewRoutineGroup()
	for i := 0; i &lt; wg.workers; i++ {
    // 内部维护了 wg.Add(1) wg.Done()
    // 同时也是 goroutine 安全模式下进行的
		group.RunSafe(wg.job)
	}
	group.Wait()
}
worker 的状态会自动管理，可以用来固定数量的 worker 来处理消息队列的任务，用法如下：

func main() {
  group := NewWorkerGroup(func() {
    // process tasks
	}, runtime.NumCPU())
	group.Start()
}
```