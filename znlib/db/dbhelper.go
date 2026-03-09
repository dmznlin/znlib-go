// Package db
/******************************************************************************
  作者: dmzn@163.com 2022-07-26 16:04:21
  描述: 多数据库连接池、嵌套事务
******************************************************************************/
package db

import (
	"fmt"
	"sync"

	. "github.com/dmznlin/znlib-go/znlib"
	"github.com/jmoiron/sqlx"
)

type (
	// Trans 数据库事务
	Trans struct {
		Db               *sqlx.DB
		Tx               *sqlx.Tx
		savePointID      string
		savePointEnabled bool
		nested           bool
	}

	// Utils 数据库辅助
	Utils struct {
		sync        sync.RWMutex       //数据库同步锁定
		DefaultType SqlDbType          //默认数据库类型
		DBList      map[string]*DbConn //多数据库配置,k:数据库名称
	}
)

// Manager 全局数据库管理器
var Manager = &Utils{
	DefaultType: DBMssql,
	DBList:      make(map[string]*DbConn),
}

// initDBManager 2022-07-26 16:33:53
/*
 描述: 初始化数据库配置
*/
func init() {
	Application.RegisterInitHandler(func(cfg *LibConfig) {
		if !cfg.DB.Enable {
			return
		}

		caller := "znlib.dbhelper.init"
		if len(cfg.DB.EncryptKey) > 0 {
			buf, err := NewEncrypter(EncryptDesEcb, []byte(DefaultEncryptKey)).Decrypt([]byte(cfg.DB.EncryptKey), true)
			if err == nil {
				if len(buf) == 8 {
					cfg.DB.EncryptKey = string(buf) //new key
				} else {
					ErrorCaller("EncryptKey length!=8", caller)
					return
				}
			} else {
				ErrorCaller(ErrorMsg(err, "EncryptKey wrong"), caller)
				return
			}
		}

		if len(cfg.DB.EncryptKey) < 1 {
			cfg.DB.EncryptKey = DefaultEncryptKey
		}

		for _, conn := range cfg.DB.DbConn {
			Manager.DBList[conn.Name] = conn
			if conn.MaxOpen < 1 {
				conn.MaxOpen = 5
			}
			if conn.MaxIdle < 1 {
				conn.MaxIdle = 2
			}

			if len(conn.Passwd) > 0 {
				buf, err := NewEncrypter(EncryptDesEcb, []byte(cfg.DB.EncryptKey)).Decrypt([]byte(conn.Passwd), true)
				if err != nil {
					ErrorCaller(ErrorMsg(err, fmt.Sprintf(`"%s.passwd" wrong`, conn.Name)), caller)
					return
				}

				conn.Passwd = string(buf)
				//密码明文
			}

			Manager.ApplyDSN(conn)
			//生成连接 dns
		}

		conn, ok := Manager.DBList[cfg.DB.DefaultName]
		if ok {
			Manager.DefaultType = conn.Type
		} else {
			ErrorCaller("config defaultDB not found", caller)
			return
		}
	})
}

// GetDB 2022-07-28 18:18:44
/*
 参数: dbname,数据库名称
 描述: 获取指定数据库连接对象
*/
func (du *Utils) GetDB(dbname string) (db *sqlx.DB, err error) {
	cfg, ok := du.DBList[dbname]
	if !ok {
		return nil, ErrorMsg(nil, fmt.Sprintf(`znlib.dbhelper.GetDB: "%s" not invalid.`, dbname))
	}

	du.sync.RLock()
	if cfg.DB != nil {
		du.sync.RUnlock()
		return cfg.DB, nil
	}
	du.sync.RUnlock()
	//for write

	du.sync.Lock()
	defer du.sync.Unlock()
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
func (du *Utils) ApplyDSN(conn *DbConn) {
	conn.DSN = StrReplace(conn.DSN, conn.User, "$user")
	conn.DSN = StrReplace(conn.DSN, conn.Passwd, "$pwd")
	conn.DSN = StrReplace(conn.DSN, conn.Host, "$host")
	conn.DSN = FixPathVar(conn.DSN)
}

// UpdateDSN 2022-08-03 13:17:26
/*
 参数: dbname,数据库名称
 参数: dsn,新的连接配置
 描述:
*/
func (du *Utils) UpdateDSN(dbname, dsn string) (err error) {
	cfg, ok := du.DBList[dbname]
	if !ok {
		return ErrorMsg(nil, fmt.Sprintf(`znlib.dbhelper.ApplyDSN: "%s" not invalid.`, dbname))
	}

	du.sync.Lock()
	defer DeferHandle(false, "znlib.dbhelper.ApplyDSN", func(e error) {
		du.sync.Unlock()
		if e != nil {
			err = e
		}
	})

	cfg.DSN = dsn
	du.ApplyDSN(cfg) //update dsn

	if cfg.DB != nil { //try to close
		db := cfg.DB
		cfg.DB = nil
		_ = db.Close()
	}

	return nil
}
