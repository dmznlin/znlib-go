// Package znlib
/******************************************************************************
  作者: dmzn@163.com 2022-06-08 14:59:33
  描述: znlib库常量定义
******************************************************************************/
package znlib

// DefaultEncryptKey 默认加密秘钥
const DefaultEncryptKey = "znlib-go"

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
	Redis_SyncLock_DateID = "znlib.lockkey.dateid" //生成日期编号的事务锁标识
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
	SQLDB_mssql      SqlDbType = "SqlServer"
	SQLDB_mysql      SqlDbType = "MySQL"
	SQLDB_db2        SqlDbType = "DB2"
	SQLDB_oracle     SqlDbType = "Oracle"
	SQLDB_postgreSQL SqlDbType = "PostgreSQL"
	SQLDB_sqlite     SqlDbType = "Sqlite"
)

// SQLDB_Types 数据库类型列表
var SQLDB_Types = []SqlDbType{SQLDB_mssql, SQLDB_mysql, SQLDB_db2, SQLDB_oracle,
	SQLDB_postgreSQL, SQLDB_sqlite}

type SqlValueQuotes = string

const (
	SqlQuotes_Single SqlValueQuotes = "'"  //单引号
	SqlQuotes_Double SqlValueQuotes = "\"" //双引号
)

type SqlDbFlag = string

// 构建SQL的结构体Tag
const (
	SQLTag_Table SqlDbFlag = "table"
	SQLTag_DB    SqlDbFlag = "db"

	SQLFlag_Insert SqlDbFlag = "i"
	SQLFlag_Update SqlDbFlag = "u"
	SQLFlag_Delete SqlDbFlag = "d"
	SQLFlag_Select SqlDbFlag = "s"
)
