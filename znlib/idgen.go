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
	"strconv"
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

//SnowflakeWorker 全局雪花算法对象
var SnowflakeWorker *snowflakeWorker = nil

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
	if err != nil {
		return "", err
	}

	if encode != nil && encode[0] == true {
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, id)

		buf, err = NewEncrypter(EncryptBase64_STD, nil).EncodeBase64(buf)
		return string(buf), nil
	} else {
		return strconv.FormatUint(id, 10), nil
	}
}
