/*Package znlib ***************************************************************
作者: dmzn@163.com 2022-05-30 13:45:17
描述: 配置lib库

描述:
1.依据配置文件初始化各单元
2.依据依赖先后顺序初始化各单元
******************************************************************************/
package znlib

import iniFile "github.com/go-ini/ini"

/*init 2022-05-30 13:47:33
  描述: 根据先后依赖调用各源文件初始化
*/
func init() {
	initApp()
	//application.go

	cfg := struct {
		logger    bool
		dbmanager bool
		snowflake bool

		workerID     int64
		datacenterID int64
	}{
		logger:    true,
		dbmanager: false,
		snowflake: true,

		workerID:     1,
		datacenterID: 0,
	}

	if FileExists(Application.ConfigFile, false) {
		ini, err := iniFile.Load(Application.ConfigFile)
		if err == nil {
			sec := ini.Section("logger")
			cfg.logger = sec.Key("enable").In("true", []string{"true", "false"}) == "true"

			sec = ini.Section("dbmanager")
			cfg.dbmanager = sec.Key("enable").In("true", []string{"true", "false"}) == "true"

			sec = ini.Section("snowflake")
			cfg.snowflake = sec.Key("enable").In("true", []string{"true", "false"}) == "true"
			cfg.workerID = sec.Key("workerID").MustInt64(1)
			cfg.datacenterID = sec.Key("dataCenterID").MustInt64(0)
		}
	}

	if cfg.logger {
		initLogger()
		//logger.go
	}

	if cfg.snowflake {
		SnowflakeID = NewSnowflake(cfg.workerID, cfg.datacenterID)
		//idgen.go
	}

	if cfg.dbmanager {
		db_init()
		//dbhelper.go
	}
}
