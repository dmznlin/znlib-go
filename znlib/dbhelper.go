// Package znlib
/******************************************************************************
  作者: dmzn@163.com 2022-07-26 16:04:21
  描述: 多数据库连接池、嵌套事务
******************************************************************************/
package znlib

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"sync"
)

// DBConfig 数据库配置项
type DBConfig struct {
	Name   string    //数据库名称
	Type   SqlDbType //数据库类型
	Drive  string    //驱动名称
	User   string    //登录用户
	Passwd string    //登录密码
	Host   string    //主机地址
	DSN    string    //连接配置项

	MaxOpen int      //同时打开的连接数(使用中+空闲)
	MaxIdle int      //最大并发空闲链接数
	DB      *sqlx.DB //数据库对象
}

// DbTrans 数据库事务
type DbTrans struct {
	Db               *sqlx.DB
	Tx               *sqlx.Tx
	savePointID      string
	savePointEnabled bool
	nested           bool
}

// DbUtils 数据库操作集合
type DbUtils struct {
	sync        sync.RWMutex         //数据库同步锁定
	EncryptKey  string               //加密秘钥
	DefaultName string               //默认数据库名称
	DefaultType SqlDbType            //默认数据库类型
	DBList      map[string]*DBConfig //多数据库配置,k:数据库名称
}

// DBManager 全局数据库管理器
var DBManager = &DbUtils{
	EncryptKey:  DefaultEncryptKey,
	DefaultName: "",
	DefaultType: SQLDB_mssql,
	DBList:      make(map[string]*DBConfig),
}

// init_db 2022-07-26 16:33:53
/*
 描述: 初始化数据库配置
*/
func init_db() {

}

// GetDB 2022-07-28 18:18:44
/*
 参数: dbname,数据库名称
 描述: 获取指定数据库连接对象
*/
func (dm *DbUtils) GetDB(dbname string) (db *sqlx.DB, err error) {
	cfg, ok := dm.DBList[dbname]
	if !ok {
		return nil, ErrorMsg(nil, fmt.Sprintf(`znlib.GetDB: "%s" not invalid.`, dbname))
	}

	dm.sync.RLock()
	if cfg.DB != nil {
		dm.sync.RUnlock()
		return cfg.DB, nil
	}
	dm.sync.RUnlock()
	//for write

	dm.sync.Lock()
	defer dm.sync.Unlock()
	if cfg.DB != nil {
		return cfg.DB, nil
	}

	db, err = sqlx.Open(cfg.Drive, cfg.DSN)
	if err == nil {
		cfg.DB = db
		db.SetMaxIdleConns(cfg.MaxIdle)
		db.SetMaxOpenConns(cfg.MaxOpen)
	}
	return
}

// ApplyDSN 2022-08-03 12:57:02
/*
 描述: 更新dsn中的变量值
*/
func (dc *DBConfig) ApplyDSN() {
	dc.DSN = StrReplace(dc.DSN, dc.User, "$user")
	dc.DSN = StrReplace(dc.DSN, dc.Passwd, "$pwd")
	dc.DSN = StrReplace(dc.DSN, dc.Host, "$host")
	dc.DSN = FixPathVar(dc.DSN)
}

// UpdateDSN 2022-08-03 13:17:26
/*
 参数: dbname,数据库名称
 参数: dsn,新的连接配置
 描述:
*/
func (dm *DbUtils) UpdateDSN(dbname, dsn string) (err error) {
	cfg, ok := dm.DBList[dbname]
	if !ok {
		return ErrorMsg(nil, fmt.Sprintf(`znlib.ApplyDSN: "%s" not invalid.`, dbname))
	}

	dm.sync.Lock()
	defer DeferHandle(false, "znlib.ApplyDSN", func(e error) {
		dm.sync.Unlock()
		if e != nil {
			err = e
		}
	})

	cfg.DSN = dsn
	cfg.ApplyDSN() //update dsn

	if cfg.DB != nil { //try to close
		db := cfg.DB
		cfg.DB = nil
		db.Close()
	}

	return nil
}
