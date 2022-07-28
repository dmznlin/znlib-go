/*Package znlib ***************************************************************
  作者: dmzn@163.com 2022-07-26 16:04:21
  描述: 多数据库连接池、嵌套事务

------------------------ db.ini示例 ------------------------
[config]
#数据库类型
MySQL=mysql_local
SQL_Server=mssql_local

#变量列表
#$user: 用户
#$pwd:密码
#$host:主机
#$path: 路径

[mysql_local]
#驱动名称
driver=mysql
#登录用户
user=root
#用户密码(DES)
passwd=eR6jbw4QNo4=
#主机地址
host=127.0.0.1
#连接配置(base64)
dsn=JHVzZXI6JHB3ZEB0Y3AoJGhvc3Q6MzMwNikvdGVzdA==

[mssql_local]
#驱动名称
driver=adodb
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
	"errors"
	"fmt"
	iniFile "github.com/go-ini/ini"
	"github.com/jmoiron/sqlx"
	"strings"
	"sync"
)

//DBEncryptKey 数据库加密秘钥
const DBEncryptKey = "libdbkey"

//DBConfig 数据库配置项
type DBConfig struct {
	Name   string    //数据库名称
	Type   SqlDbType //数据库类型
	Drive  string    //驱动名称
	User   string    //登录用户
	Passwd string    //登录密码
	Host   string    //主机地址
	DSN    string    //连接配置项
	DB     *sqlx.DB  //数据库对象
}

//DBList 多数据库配置,k:数据库名称
var DBList = make(map[string]*DBConfig)

/*db_init 2022-07-26 16:33:53
  描述: 初始化数据库配置
*/
func db_init() {
	ini, err := iniFile.Load(Application.ConfigDB)
	if err != nil {
		Error(err)
		return
	}

	var (
		str    string
		strs   []string
		buf    []byte
		sec    *iniFile.Section
		key    *iniFile.Key
		dbtype = make(map[string]SqlDbType)
	)

	sec, err = ini.GetSection("config")
	if err != nil {
		Error("db-config file has no [config] section.")
		return
	}

	for _, key = range sec.Keys() {
		if !StrIn(key.Name(), SQLDB_Types...) { //invalid dbtype
			continue
		}

		str = StrTrim(key.String())
		if str == "" { //no db
			continue
		}

		strs = strings.Split(str, ",")
		for _, str = range strs {
			str = StrTrim(str)
			if str != "" {
				dbtype[str] = key.Name()
			}
		}
	}

	for _, sec = range ini.Sections() {
		key, err = sec.GetKey("dsn")
		if err != nil { //invalid section
			continue
		}

		typ, ok := dbtype[sec.Name()]
		if !ok { //no match db-type
			Error(fmt.Sprintf(`db_init:"%s" invalid db-type.`, sec.Name()))
			continue
		}

		db := &DBConfig{
			Name:   sec.Name(),
			Type:   typ,
			DSN:    key.String(),
			Drive:  sec.Key("driver").String(),
			User:   sec.Key("user").String(),
			Passwd: sec.Key("passwd").String(),
			Host:   sec.Key("host").String(),
		}

		buf, err = NewEncrypter(EncryptBase64_STD, nil).DecodeBase64([]byte(db.DSN))
		if err != nil {
			Error(fmt.Sprintf(`db_init:"%s.dsn" invalid base64 coding.`, sec.Name()))
			continue
		}
		db.DSN = string(buf)

		buf, err = NewEncrypter(EncryptDES_ECB, []byte(DBEncryptKey)).Decrypt([]byte(db.Passwd), true)
		if err != nil {
			Error(fmt.Sprintf(`db_init:"%s.passwd" wrong: %s.`, sec.Name(), err.Error()))
			continue
		}
		db.Passwd = string(buf)

		db.DSN = StrReplace(db.DSN, db.User, "$user")
		db.DSN = StrReplace(db.DSN, db.Passwd, "$pwd")
		db.DSN = StrReplace(db.DSN, db.Host, "$host")
		db.DSN = StrReplace(db.DSN, Application.ExePath, "$path\\", "$path/", "$path")
		DBList[db.Name] = db
	}
}

//dbsync 数据库同步锁定
var dbsync sync.Mutex

type DBManager struct{}

/*GetDB 2022-07-28 18:18:44
  参数: dbname,数据库名称
  描述: 获取指定数据库连接对象
*/
func (dm DBManager) GetDB(dbname string) (db *sqlx.DB, err error) {
	cfg, ok := DBList[dbname]
	if !ok {
		return nil, errors.New(fmt.Sprintf(`znlib.GetDB: "%s" not invalid.`, dbname))
	}

	dbsync.Lock()
	defer dbsync.Unlock()

	if cfg.DB != nil {
		return cfg.DB, nil
	}

	db, err = sqlx.Open(cfg.Drive, cfg.DSN)
	if err == nil {
		cfg.DB = db
	}
	return
}

type DBTrans struct {
	db               *sqlx.DB
	tx               *sqlx.Tx
	savePointID      string
	savePointEnabled bool
	nested           bool
}
