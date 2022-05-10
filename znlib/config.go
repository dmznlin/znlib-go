package znlib

/******************************************************************************
作者: dmzn@163.com 2022-05-10
描述: 配置lib库

描述:
1.依据配置文件初始化各单元
2.依据依赖先后顺序初始化各单元
******************************************************************************/
import iniFile "github.com/go-ini/ini"

//Date: 2022-05-09
//Desc: 根据先后依赖调用各源文件初始化
func init() {
	initApp()
	//application.go

	cfg := struct {
		logger bool
	}{
		logger: true,
	}

	if FileExists(Application.ConfigFile, false) {
		ini, err := iniFile.Load(Application.ConfigFile)
		if err == nil {
			sec := ini.Section("znlib")
			cfg.logger = sec.Key("EnableLogger").In("true", []string{"true", "false"}) == "true"
		}
	}

	if cfg.logger {
		initLogger()
		//logger.go
	}

}
