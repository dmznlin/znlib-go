// Package znlib
/******************************************************************************
	作者: dmzn@163.com 2022-07-26 16:04:21
	描述: 多数据库连接池、嵌套事务

------------------------ db.ini示例 ------------------------
[config]
#加密秘钥
EncryptKey=
#默认数据库
DefaultDB=mssql_main

#变量列表
#$user: 用户
#$pwd:密码
#$host:主机
#$path: 路径

[mysql_main]
#驱动名称
driver=mysql
#数据库类型
dbtype=MySQL
#连接池配置
MaxOpen=5
MaxIdle=2
#登录用户
user=root
#用户密码(DES)
passwd=eR6jbw4QNo4=
#主机地址
host=127.0.0.1
#连接配置(base64)
dsn=JHVzZXI6JHB3ZEB0Y3AoJGhvc3Q6MzMwNikvdGVzdA==

[mssql_main]
#驱动名称
driver=adodb
#数据库类型
dbtype=SQL_Server
#连接池配置
MaxOpen=5
MaxIdle=2
#登录用户
user=sa
#用户密码(DES)
passwd=jKwUUfac8V4=
#主机地址
host=127.0.0.1
#连接配置(base64)
dsn=UHJvdmlkZXI9U1FMT0xFREI7SW5pdGlhbCBDYXRhbG9nPVByaW50U2hvcDt1c2VyIGlkPSR1c2
VyO3Bhc3N3b3JkPSRwd2Q7RGF0YSBTb3VyY2U9JGhvc3Q=
******************************************************************************/
package znlib

import (
	"fmt"
	iniFile "github.com/go-ini/ini"
	"github.com/jmoiron/sqlx"
	"strings"
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
var DBManager = DbUtils{
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
	DBManager.LoadConfig()
}

// LoadConfig 2022-08-02 21:48:53
/*
 参数: file,配置文件
 描述: 读取数据库配置
*/
func (dm *DbUtils) LoadConfig(file ...string) (err error) {
	var ini *iniFile.File
	if file == nil {
		ini, err = iniFile.Load(Application.ConfigDB)
	} else {
		ini, err = iniFile.Load(file[0])
	}

	if err != nil {
		Error(err)
		return
	}

	var (
		str string
		buf []byte
		sec *iniFile.Section
		key *iniFile.Key
	)

	sec, err = ini.GetSection("config")
	if err == nil {
		str = StrTrim(sec.Key("EncryptKey").String()) //秘钥
		if str != "" {
			buf, err = NewEncrypter(EncryptDES_ECB, []byte(DefaultEncryptKey)).Decrypt([]byte(str), true)
			if err == nil {
				if len(buf) == 8 {
					dm.EncryptKey = string(buf) //new key
				} else {
					Error(fmt.Sprintf(`DbUtils:"%s.EncryptKey" length!=8.`, sec.Name()))
				}
			} else {
				Error(fmt.Sprintf(`DbUtils:"%s.EncryptKey" wrong: %s.`, sec.Name(), err))
			}
		}

		str = StrTrim(sec.Key("DefaultDB").String()) //默认数据库
		if str != "" {
			dm.DefaultName = str
		}
	} else {
		Error("DbUtils:db-config file has no [config] section.")
	}

	if dm.EncryptKey == "" { //default key
		dm.EncryptKey = DefaultEncryptKey
	}

	if dm.DBList == nil { //empty list
		dm.DBList = make(map[string]*DBConfig)
	}
	//-------------------------------------------------------------------------

	for _, sec = range ini.Sections() {
		key, err = sec.GetKey("dsn")
		if err != nil { //invalid section
			continue
		}

		str = sec.Key("dbtype").String()
		if !StrIn(str, SQLDB_Types...) { //no match db-type
			Error(fmt.Sprintf(`DbUtils:"%s" invalid db-type.`, sec.Name()))
			continue
		}

		db := &DBConfig{
			Name:   sec.Name(),
			Type:   str,
			DSN:    key.String(),
			Drive:  sec.Key("driver").String(),
			User:   sec.Key("user").String(),
			Passwd: sec.Key("passwd").String(),
			Host:   sec.Key("host").String(),

			MaxIdle: sec.Key("MaxIdle").MustInt(2),
			MaxOpen: sec.Key("MaxOpen").MustInt(5),
		}

		buf, err = NewEncrypter(EncryptBase64_STD, nil).DecodeBase64([]byte(db.DSN))
		if err != nil {
			Error(fmt.Sprintf(`DbUtils:"%s.dsn" invalid base64 coding.`, sec.Name()))
			continue
		}
		db.DSN = string(buf)

		buf, err = NewEncrypter(EncryptDES_ECB, []byte(dm.EncryptKey)).Decrypt([]byte(db.Passwd), true)
		if err != nil {
			Error(fmt.Sprintf(`DbUtils:"%s.passwd" wrong: %s.`, sec.Name(), err))
			continue
		}

		db.Passwd = string(buf)
		db.ApplyDSN() //update value
		dm.DBList[db.Name] = db

		if dm.DefaultName == "" { //first is default
			dm.DefaultName = db.Name
			dm.DefaultType = db.Type
		} else if strings.EqualFold(db.Name, dm.DefaultName) { //match default type
			dm.DefaultType = db.Type
		}
	}

	if len(dm.DBList) > 0 {
		return nil
	} else {
		return ErrorMsg(nil, "DbUtils:db-list is empty.")
	}
}

// GetDB 2022-07-28 18:18:44
/*
 参数: dbname,数据库名称
 描述: 获取指定数据库连接对象
*/
func (dm DbUtils) GetDB(dbname string) (db *sqlx.DB, err error) {
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
	dc.DSN = StrReplace(dc.DSN, Application.ExePath, "$path\\", "$path/", "$path")
}

// UpdateDSN 2022-08-03 13:17:26
/*
 参数: dbname,数据库名称
 参数: dsn,新的连接配置
 描述:
*/
func (dm DbUtils) UpdateDSN(dbname, dsn string) (err error) {
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
