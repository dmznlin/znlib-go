/*Package znlib ***************************************************************
  作者: dmzn@163.com 2022-07-26 16:04:21
  描述: 多数据库连接池、嵌套事务
******************************************************************************/
package znlib

import (
	"encoding/base64"
	"fmt"
	iniFile "github.com/go-ini/ini"
	"github.com/jmoiron/sqlx"
	"strings"
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
	File   string    //数据文件
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
			Error(fmt.Sprintf(`db "%s" invalid db-type.`, sec.Name()))
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

		buf, err = base64.StdEncoding.DecodeString(db.DSN)
		if err != nil {
			Error(fmt.Sprintf(`db "%s.dsn" invalid base64 coding.`, sec.Name()))
			continue
		}

		db.DSN = string(buf)

		DBList[db.DSN] = db
	}
}

type DBManager struct{}

func (dm *DBManager) GetDB(db string) *sqlx.DB {
	return nil
}

type DBTrans struct {
	db               *sqlx.DB
	tx               *sqlx.Tx
	savePointID      string
	savePointEnabled bool
	nested           bool
}
