// Package redis
/******************************************************************************
  作者: dmzn@163.com 2022-08-11 19:50:59
  描述: 支持集群的redis客户端
******************************************************************************/
package redis

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"strconv"
	"strings"
	"time"

	. "github.com/dmznlin/znlib-go/znlib"
	"github.com/go-redis/redis/v8"
)

type (
	Utils struct {
		redis.Cmdable //redis 操作接口
	}

	Locker struct {
		key string //锁名称
		tag string //加锁标识
		err error  //加锁状态
	}
)

var (
	// Single 单机连接
	Single *redis.Client = nil

	// Cluster 集群模式连接
	Cluster *redis.ClusterClient = nil

	// Client 全局redis统一接口
	Client = &Utils{
		Cmdable: nil,
	}
)

// initRedis 2022-08-12 12:44:13
/*
 描述: 初始化redis客户端
*/
func init() {
	Application.RegisterInitHandler(func(cfg *LibConfig) {
		if len(cfg.Redis.Servers) < 1 {
			Error("redis.initRedis: server list empty.")
			return
		}

		if len(cfg.Redis.Password) > 0 {
			buf, err := NewEncrypter(EncryptDesEcb, []byte(DefaultEncryptKey)).Decrypt([]byte(cfg.Redis.Password), true)
			if err == nil {
				cfg.Redis.Password = string(buf)
			} else {
				ErrorCaller(err, "znlib.redis.init")
				return
			}
		}

		cfg.Redis.Timeout.Dial = cfg.Redis.Timeout.Dial * time.Second
		cfg.Redis.Timeout.Read = cfg.Redis.Timeout.Read * time.Second
		cfg.Redis.Timeout.Write = cfg.Redis.Timeout.Write * time.Second
		cfg.Redis.Timeout.Pool = cfg.Redis.Timeout.Pool * time.Second

		if cfg.Redis.Cluster {
			Cluster = redis.NewClusterClient(&redis.ClusterOptions{
				Addrs:    cfg.Redis.Servers,
				Password: cfg.Redis.Password,
				PoolSize: cfg.Redis.PoolSize,

				//超时设置
				DialTimeout:  cfg.Redis.Timeout.Dial,  //连接建立超时时间，默认5秒。
				ReadTimeout:  cfg.Redis.Timeout.Read,  //读超时，默认3秒， -1表示取消读超时
				WriteTimeout: cfg.Redis.Timeout.Write, //写超时，默认等于读超时，-1表示取消读超时
				PoolTimeout:  cfg.Redis.Timeout.Pool,  //当所有连接都处在繁忙状态时，客户端等待可用连接的最大等待时长，默认为读超时+1秒。
			})

			Client.Cmdable = Cluster
		} else {
			Single = redis.NewClient(&redis.Options{
				Addr:     cfg.Redis.Servers[0],
				Password: cfg.Redis.Password,
				PoolSize: cfg.Redis.PoolSize,
				DB:       cfg.Redis.DefaultDB,

				//超时设置
				DialTimeout:  cfg.Redis.Timeout.Dial,
				ReadTimeout:  cfg.Redis.Timeout.Read,
				WriteTimeout: cfg.Redis.Timeout.Write,
				PoolTimeout:  cfg.Redis.Timeout.Pool,
			})

			Client.Cmdable = Single
		}
	})
}

// Ping 2022-08-12 19:21:09
/*
 描述: 检测服务器是否正常
*/
func (ru *Utils) Ping() (str string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	str, err = ru.Cmdable.Ping(ctx).Result()
	return str, ErrorMsg(err, "redis.Ping")
}

// Lock 2022-08-12 19:21:29
/*
 参数: key,锁名称
 参数: waite,等待时长
 参数: timeout,自动加锁超时
 描述: 创建名为key、时长为timeout的锁,若无法获取则等待waite时长.
*/
func (ru *Utils) Lock(key string, waite, timeout time.Duration) *Locker {
	var lock Locker
	lock.key = key
	if GlobalConfig.Snow.Enable {
		lock.tag, lock.err = SnowflakeID.NextStr()
	} else {
		lock.err = fmt.Errorf("redis.Locker: snowflake not enabled")
		Warn(lock.err)
	}

	if lock.err != nil {
		lock.tag = SerialID.TimeID() + strconv.Itoa(rand.Intn(100))
		//使用本机时钟编号,较慢
	}

	start := time.Now()
	for ru.SetNX(Application.Ctx, key, lock.tag, timeout).Val() == false {
		runtime.Gosched()
		//让出 CPU 时间片

		if time.Since(start) >= waite {
			lock.err = errors.New("redis.Locker: wait timeout")
			break
		}
	}

	return &lock
}

// Unlock 2022-08-12 22:53:19
/*
 描述: 解除锁
*/
func (r *Locker) Unlock() {
	if r.err == nil && r.tag != "" {
		if Client.Get(Application.Ctx, r.key).Val() == r.tag {
			Client.Del(Application.Ctx, r.key)
			r.tag = "" //结束锁定(无效化)
		}
	}
}

// DateID 2022-08-14 20:36:06
/*
 参数: key,id标识
 参数: idLen,总长度
 参数: prefix,前缀
 描述: 生成key标识长度为idLen,每日从1递增的编号.

 格式: 前缀 + 年月日 + 顺序编号,ex: 220814001
 注意: 该函数依赖redis服务,使用相同redis.db生成的id唯一.
*/
func (ru *Utils) DateID(key string, idLen int, prefix ...string) (id string, err error) {
	caller := "idgen.DateID"
	defer DeferHandle(false, caller, func(e error) {
		if e != nil {
			err = ErrorMsg(e, caller)
		}
	})

	lock := ru.Lock(RedisSyncLockDateID, 3*time.Second, 10*time.Second)
	if lock.err != nil {
		return "", ErrorMsg(lock.err, caller)
	}
	defer lock.Unlock()

	const (
		fieldBase = "base"
		fieldDate = "date"
	)

	var (
		vals []interface{}
		base = "1"
		now  = DateTime2Str(time.Now(), "060102")
	)

	key = "serial.dateid:" + key
	//避开 key 冲突

	if ru.Exists(Application.Ctx, key).Val() == 1 {
		vals, err = ru.HMGet(Application.Ctx, key, fieldBase, fieldDate).Result()
		if err != nil {
			return "", ErrorMsg(err, caller)
		}

		date := vals[1].(string)
		if date == now { //
			base = vals[0].(string)
			val, e := strconv.ParseInt(base, 10, 64)
			if e == nil {
				val++
				base = strconv.FormatInt(val, 10)
			}
		}
	}

	ru.HMSet(Application.Ctx, key, map[string]string{fieldBase: base, fieldDate: now})
	//更新编号参数
	lock.Unlock()

	if prefix != nil { //1.prefix
		id = prefix[0]
	}

	id = id + now //2.date
	num := idLen - len(id+base)
	if num > 0 {
		return id + strings.Repeat("0", num) + base, nil
	}

	return id + base, nil
}
