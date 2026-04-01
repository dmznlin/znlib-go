// Package znlib
/******************************************************************************
  作者: dmzn@163.com 2022-06-08 14:59:33
  描述: znlib库常量定义
******************************************************************************/
package znlib

// DefaultEncryptKey 默认加密秘钥
var DefaultEncryptKey = "znlib-go"

const (
	StrTrue  = "true"  //true string
	StrFalse = "false" //false string
)

const (
	LayoutTime          = "15:04:05"                //时间
	LayoutTimeMilli     = "15:04:05.000"            //时间 + 毫秒
	LayoutDate          = "2006-01-02"              //日期
	LayoutDateTime      = "2006-01-02 15:04:05"     //日期 + 时间
	LayoutDateTimeMilli = "2006-01-02 15:04:05.000" //日期 + 时间 + 毫秒
)

//--------------------------------------------------------------------------------

const (
	RedisSyncLockDateID = "znlib.lockkey.dateid" //生成日期编号的事务锁标识
)

//-----------------------------------------------------------------------------

// ValueRelation 数值比较关系
type ValueRelation = byte

const (
	ValEqual        ValueRelation = iota // =
	ValGreater                           // >
	ValGreaterEqual                      // >=
	ValLess                              // <
	ValLessEqual                         // <=
)

//-----------------------------------------------------------------------------

type SqlDbType = string

// 数据库类型定义
const (
	DBMssql    SqlDbType = "SqlServer"
	DBMysql    SqlDbType = "MySQL"
	DBDb2      SqlDbType = "DB2"
	DBOracle   SqlDbType = "Oracle"
	DBPostgres SqlDbType = "PostgreSQL"
	DBSqlite   SqlDbType = "Sqlite"
)

// DBTypes 数据库类型列表
var DBTypes = []SqlDbType{DBMssql, DBMysql, DBDb2, DBOracle,
	DBPostgres, DBSqlite}

type SqlValueQuotes = string

const (
	QuotesSingle SqlValueQuotes = "'"  //单引号
	QuotesDouble SqlValueQuotes = "\"" //双引号
)

type SqlDbFlag = string

// 构建 SQL 的结构体Tag
const (
	TagTable SqlDbFlag = "table"
	TagDB    SqlDbFlag = "db"

	FlagInsert SqlDbFlag = "i"
	FlagUpdate SqlDbFlag = "u"
	FlagDelete SqlDbFlag = "d"
	FlagSelect SqlDbFlag = "s"
)
