// Package znlib
/******************************************************************************
  作者: dmzn@163.com 2022-06-08 14:59:33
  描述: znlib库常量定义
******************************************************************************/
package znlib

// DefaultEncryptKey 默认加密秘钥
const DefaultEncryptKey = "znlib-go"

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
	SQLDB_mssql      SqlDbType = "SQL_Server"
	SQLDB_mysql      SqlDbType = "MySQL"
	SQLDB_db2        SqlDbType = "DB2"
	SQLDB_oracle     SqlDbType = "Oracle"
	SQLDB_postgreSQL SqlDbType = "PostgreSQL"
	SQLDB_sqlite     SqlDbType = "SQL_Lite"
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

//-----------------------------------------------------------------------------

// Key define
type Key = uint16

// Key constants
const (
	KeyF1 Key = 0xFFFF - iota
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
	KeyInsert
	KeyDelete
	KeyHome
	KeyEnd
	KeyPgup
	KeyPgdn
	KeyArrowUp
	KeyArrowDown
	KeyArrowLeft
	KeyArrowRight
	key_min // see terminfo
	MouseLeft
	MouseMiddle
	MouseRight
	MouseRelease
	MouseWheelUp
	MouseWheelDown
)

const (
	KeyCtrlTilde      Key = 0x00
	KeyCtrl2          Key = 0x00
	KeyCtrlSpace      Key = 0x00
	KeyCtrlA          Key = 0x01
	KeyCtrlB          Key = 0x02
	KeyCtrlC          Key = 0x03
	KeyCtrlD          Key = 0x04
	KeyCtrlE          Key = 0x05
	KeyCtrlF          Key = 0x06
	KeyCtrlG          Key = 0x07
	KeyBackspace      Key = 0x08
	KeyCtrlH          Key = 0x08
	KeyTab            Key = 0x09
	KeyCtrlI          Key = 0x09
	KeyCtrlJ          Key = 0x0A
	KeyCtrlK          Key = 0x0B
	KeyCtrlL          Key = 0x0C
	KeyEnter          Key = 0x0D
	KeyCtrlM          Key = 0x0D
	KeyCtrlN          Key = 0x0E
	KeyCtrlO          Key = 0x0F
	KeyCtrlP          Key = 0x10
	KeyCtrlQ          Key = 0x11
	KeyCtrlR          Key = 0x12
	KeyCtrlS          Key = 0x13
	KeyCtrlT          Key = 0x14
	KeyCtrlU          Key = 0x15
	KeyCtrlV          Key = 0x16
	KeyCtrlW          Key = 0x17
	KeyCtrlX          Key = 0x18
	KeyCtrlY          Key = 0x19
	KeyCtrlZ          Key = 0x1A
	KeyEsc            Key = 0x1B
	KeyCtrlLsqBracket Key = 0x1B
	KeyCtrl3          Key = 0x1B
	KeyCtrl4          Key = 0x1C
	KeyCtrlBackslash  Key = 0x1C
	KeyCtrl5          Key = 0x1D
	KeyCtrlRsqBracket Key = 0x1D
	KeyCtrl6          Key = 0x1E
	KeyCtrl7          Key = 0x1F
	KeyCtrlSlash      Key = 0x1F
	KeyCtrlUnderscore Key = 0x1F
	KeySpace          Key = 0x20
	KeyBackspace2     Key = 0x7F
	KeyCtrl8          Key = 0x7F
)

// windows virtual keys
const (
	vk_backspace   = 0x8
	vk_tab         = 0x9
	vk_enter       = 0xd
	vk_esc         = 0x1b
	vk_space       = 0x20
	vk_pgup        = 0x21
	vk_pgdn        = 0x22
	vk_end         = 0x23
	vk_home        = 0x24
	vk_arrow_left  = 0x25
	vk_arrow_up    = 0x26
	vk_arrow_right = 0x27
	vk_arrow_down  = 0x28
	vk_insert      = 0x2d
	vk_delete      = 0x2e

	vk_f1  = 0x70
	vk_f2  = 0x71
	vk_f3  = 0x72
	vk_f4  = 0x73
	vk_f5  = 0x74
	vk_f6  = 0x75
	vk_f7  = 0x76
	vk_f8  = 0x77
	vk_f9  = 0x78
	vk_f10 = 0x79
	vk_f11 = 0x7a
	vk_f12 = 0x7b

	right_alt_pressed  = 0x1
	left_alt_pressed   = 0x2
	right_ctrl_pressed = 0x4
	left_ctrl_pressed  = 0x8
	shift_pressed      = 0x10

	k32_keyEvent = 0x1
)
