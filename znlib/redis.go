// Package znlib
/******************************************************************************
  作者: dmzn@163.com 2022-08-11 19:50:59
  描述: 支持集群的redis客户端
******************************************************************************/
package znlib

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"math/rand"
	"runtime"
	"strconv"
	"time"
)

// redisConfig 配置参数
var redisConfig = struct {
	cluster      bool          //是否集群
	servers      []string      //服务器列表
	password     string        //服务密码
	poolSize     int           //最大连接数
	defaultDB    int           //默认数据库索引
	dialTimeout  time.Duration //连接建立超时
	readTimeout  time.Duration //读超时
	writeTimeout time.Duration //写超时
	poolTimeout  time.Duration //繁忙状态时等待
}{
	cluster:  false,
	servers:  make([]string, 0),
	password: "",
}

// RedisSingle 单机连接
var RedisSingle *redis.Client = nil

// RedisCluster 集群模式连接
var RedisCluster *redis.ClusterClient = nil

// init_redis 2022-08-12 12:44:13
/*
 描述: 初始化redis客户端
*/
func init_redis() {
	if len(redisConfig.servers) < 1 {
		Error("znlib.redis.init_redis: server list empty.")
		return
	}

	if redisConfig.cluster {
		RedisCluster = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    redisConfig.servers,
			Password: redisConfig.password,
			PoolSize: redisConfig.poolSize,

			//超时设置
			DialTimeout:  redisConfig.dialTimeout,  //连接建立超时时间，默认5秒。
			ReadTimeout:  redisConfig.readTimeout,  //读超时，默认3秒， -1表示取消读超时
			WriteTimeout: redisConfig.writeTimeout, //写超时，默认等于读超时，-1表示取消读超时
			PoolTimeout:  redisConfig.poolTimeout,  //当所有连接都处在繁忙状态时，客户端等待可用连接的最大等待时长，默认为读超时+1秒。
		})

		RedisClient.Cmdable = RedisCluster

	} else {
		RedisSingle = redis.NewClient(&redis.Options{
			Addr:     redisConfig.servers[0],
			Password: redisConfig.password,
			PoolSize: redisConfig.poolSize,
			DB:       redisConfig.defaultDB,

			//超时设置
			DialTimeout:  redisConfig.dialTimeout,
			ReadTimeout:  redisConfig.readTimeout,
			WriteTimeout: redisConfig.writeTimeout,
			PoolTimeout:  redisConfig.poolTimeout,
		})

		RedisClient.Cmdable = RedisSingle
	}
}

//-----------------------------------------------------------------------------

// RedisClient 全局redis统一接口
var RedisClient = &redisUtils{
	Cmdable: nil,
}

// redisClient redis
type redisUtils struct {
	redis.Cmdable //redis操作接口
}

type RedisLock struct {
	key string //锁名称
	tag string //加锁标识
	err error  //加锁状态
}

// Ping 2022-08-12 19:21:09
/*
 描述: 检测服务器是否正常
*/
func (r *redisUtils) Ping() (str string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	str, err = r.Cmdable.Ping(ctx).Result()
	return str, ErrorMsg(err, "znlib.redis.Ping")
}

// Lock 2022-08-12 19:21:29
/*
 参数: key,锁名称
 参数: waitfor,等待时长
 参数: timeout,自动加锁超时
 描述: 创建名为key、时长为timeout的锁,若无法获取则等待waitfor时长.
*/
func (r *redisUtils) Lock(key string, waitfor, timeout time.Duration) *RedisLock {
	var lock RedisLock
	lock.key = key
	lock.tag, lock.err = SnowflakeID.NextStr()

	if lock.err != nil {
		lock.tag = SerialID.TimeID() + strconv.Itoa(rand.Intn(100))
		//使用本机时钟编号,较慢
	}

	start := time.Now()
	for r.SetNX(Application.Ctx, key, lock.tag, timeout).Val() == false {
		runtime.Gosched()
		//让出CPU时间片

		if time.Since(start) >= waitfor {
			lock.err = errors.New("znlib.redis.Lock: wait timeout.")
			break
		}
	}

	return &lock
}

// Unlock 2022-08-12 22:53:19
/*
 描述: 解除锁
*/
func (r *RedisLock) Unlock() {
	if r.err == nil && r.tag != "" {
		if RedisClient.Get(Application.Ctx, r.key).Val() == r.tag {
			RedisClient.Del(Application.Ctx, r.key)
			r.tag = "" //结束锁定(无效化)
		}
	}
}
