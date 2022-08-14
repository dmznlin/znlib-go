/*Package znlib ***************************************************************
  作者: dmzn@163.com 2022-08-10 11:33:54
  描述: 唯一标识(ID)

  备注:
  *.雪花算法:最多使用69年
	41bit timestamp | 10 bit machineID : 5bit workerID 5bit dataCenterID ｜ 12 bit sequenceBits
******************************************************************************/
package znlib

import (
	"encoding/binary"
	"errors"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	workerIDBits     = uint64(5) // 10bit 工作机器ID中的 5bit workerID
	dataCenterIDBits = uint64(5) // 10 bit 工作机器ID中的 5bit dataCenterID
	sequenceBits     = uint64(12)

	maxWorkerID     = int64(-1) ^ (int64(-1) << workerIDBits) //节点ID的最大值 用于防止溢出
	maxDataCenterID = int64(-1) ^ (int64(-1) << dataCenterIDBits)
	maxSequence     = int64(-1) ^ (int64(-1) << sequenceBits)

	timeLeft = uint8(22)            // timeLeft = workerIDBits + sequenceBits // 时间戳向左偏移量
	dataLeft = uint8(17)            // dataLeft = dataCenterIDBits + sequenceBits
	workLeft = uint8(12)            // workLeft = sequenceBits // 节点IDx向左偏移量
	twepoch  = int64(1589923200000) //常量时间戳(毫秒): 2020-05-20 08:00:00 +0800 CST
)

type snowflakeWorker struct {
	mu           sync.Mutex
	LastStamp    int64 // 记录上一次ID的时间戳
	WorkerID     int64 // 该节点的ID
	DataCenterID int64 // 该节点的 数据中心ID
	Sequence     int64 // 当前毫秒已经生成的ID序列号(从0 开始累加) 1毫秒内最多生成4096个ID
}

//SnowflakeID 全局雪花算法对象
var SnowflakeID *snowflakeWorker = nil

//snowflakeConfig 配置参数
var snowflakeConfig = struct {
	workerID     int64
	datacenterID int64
}{
	workerID:     1,
	datacenterID: 0,
}

/*init_snowflake 2022-08-11 19:03:40
  描述: 初始化对象
*/
func init_snowflake() {
	SnowflakeID = NewSnowflake(snowflakeConfig.workerID, snowflakeConfig.datacenterID)
}

/*NewSnowflake 2022-08-10 11:42:26
  参数: workerID,节点标识
  参数: dataCenterID,数据中心标识
  描述: 分布式情况下,应通过外部配置文件或其他方式为每台机器分配独立的id
*/
func NewSnowflake(workerID, dataCenterID int64) *snowflakeWorker {
	return &snowflakeWorker{
		WorkerID:     workerID,
		LastStamp:    0,
		Sequence:     0,
		DataCenterID: dataCenterID,
	}
}

func (w *snowflakeWorker) getMilliSeconds() int64 {
	return time.Now().UnixNano() / 1e6
}

/*NextID 2022-08-10 12:30:51
  描述: 生成序列号
*/
func (w *snowflakeWorker) NextID() (uint64, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	timeStamp := w.getMilliSeconds()
	if timeStamp < w.LastStamp {
		return 0, errors.New("Snowflake.NextID: time is moving backwards")
	}

	if w.LastStamp == timeStamp {
		w.Sequence = (w.Sequence + 1) & maxSequence
		if w.Sequence == 0 {
			for timeStamp <= w.LastStamp {
				timeStamp = w.getMilliSeconds()
			}
		}
	} else {
		w.Sequence = 0
	}

	w.LastStamp = timeStamp
	id := ((timeStamp - twepoch) << timeLeft) | (w.DataCenterID << dataLeft) | (w.WorkerID << workLeft) | w.Sequence
	return uint64(id), nil
}

/*NextStr 2022-08-10 12:37:35
  参数: encode,是否编码
  描述: 生成字符串序列号
*/
func (w *snowflakeWorker) NextStr(encode ...bool) (string, error) {
	id, err := w.NextID()
	if err == nil {
		return SerialID.ToString(id, encode...)
	} else {
		return "", err
	}
}

//--------------------------------------------------------------------------------

//serialIDWorker 递增编号
type serialIDWorker struct {
	mu   sync.Mutex //同步锁定
	base uint64     //编号基数
}

//SerialID 全局串行编号对象
var SerialID *serialIDWorker = &serialIDWorker{
	base: 0,
}

/*NextID 2022-08-10 15:03:59
  描述: 本地顺序编号,从1开始
*/
func (w *serialIDWorker) NextID() uint64 {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.base < math.MaxUint64 {
		w.base++
	} else {
		w.base = 1
	}
	return w.base
}

/*NextStr 2022-08-10 12:37:35
  参数: encode,是否编码
  描述: 生成字符串序列号
*/
func (w *serialIDWorker) NextStr(encode ...bool) (string, error) {
	return w.ToString(w.NextID(), encode...)
}

/*ToString 2022-08-10 15:11:28
  参数: id,编号值
  参数: encode,是否编码
  描述: 将id转为字符串id
*/
func (w *serialIDWorker) ToString(id uint64, encode ...bool) (sid string, err error) {
	if encode != nil && encode[0] == true {
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, id)

		buf, err = NewEncrypter(EncryptBase64_STD, nil).EncodeBase64(buf)
		return string(buf), nil
	} else {
		return strconv.FormatUint(id, 10), nil
	}
}

/*TimeID 2022-08-10 17:06:42
  参数: year,包含年月日
  描述: 使用时分秒+毫秒作为编号
*/
func (w *serialIDWorker) TimeID(year ...bool) string {
	var lay string
	if year != nil && year[0] == true {
		lay = "20060102150405.000"
	} else {
		lay = "150405.000"
	}

	id := time.Now().Format(lay)
	time.Sleep(1 * time.Millisecond)
	//避免毫秒重复

	pos := strings.Index(id, ".")
	buf := []byte(id)
	return string(append(buf[:pos], buf[pos+1:]...))
}

/*DateID 2022-08-14 20:36:06
  参数: key,id标识
  参数: idlen,总长度
  参数: prefix,前缀
  描述: 生成key标识长度为idlen,每日从1递增的编号.

  格式: 前缀 + 年月日 + 顺序编号,ex: 220814001
  注意: 该函数依赖redis服务,使用相同redis.db生成的id唯一.
*/
func (w *serialIDWorker) DateID(key string, idlen int, prefix ...string) (id string, err error) {
	defer DeferHandle(false, "znlib.DateID", func(e any) {
		if e != nil {
			err = errors.New("serialIDWorker.DateID has error.")
			//any error
		}
	})

	lock := RedisClient.Lock(Redis_SyncLock_DateID, 3*time.Second, 10*time.Second)
	if lock.err != nil {
		return "", ErrorMsg(lock.err, "znlib.DateID")
	}
	defer lock.Unlock()

	var (
		key_base = key + ".base:int"
		key_date = key + ".date:string"
		now      = DateTime2Str(time.Now(), "060102")

		val        int
		base, date string
	)

	if RedisClient.Exists(Application.Ctx, key_base).Val() == 1 { //编号基数
		val, err = RedisClient.Get(Application.Ctx, key_base).Int()
		if err != nil {
			return "", ErrorMsg(err, "znlib.DateID")
		}

		val++
		base = strconv.Itoa(val)
	} else {
		base = "1"
	}

	if RedisClient.Exists(Application.Ctx, key_date).Val() == 1 { //编号日期
		date, err = RedisClient.Get(Application.Ctx, key_date).Result()
		if err != nil {
			return "", ErrorMsg(err, "znlib.DateID")
		}

		if date != now {
			base = "1"
		}
	}

	RedisClient.Set(Application.Ctx, key_base, base, 0)
	RedisClient.Set(Application.Ctx, key_date, now, 0)
	//更新编号参数

	if prefix != nil { //1.prefix
		id = prefix[0]
	}

	id = id + now //2.date
	num := idlen - len(id+base)
	if num > 0 {
		return id + strings.Repeat("0", num) + base, nil
	} else {
		return id + base, nil
	}
}
