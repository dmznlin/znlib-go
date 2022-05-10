package znlib

/******************************************************************************
作者: dmzn@163.com 2022-05-09
描述: 系统常量、变量、函数等
******************************************************************************/

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

//application full name
var AppFile = os.Args[0]

//application所在路径
var AppPath = filepath.Dir(AppFile)

//路径分隔符: / or \\
var PathSeparator = "/"

//application相关属性
type application struct {
	ExeName    string //exe full
	ExePath    string //exe所在路径
	LogPath    string //日志目录
	ConfigFile string //主配置文件
	ConfigDB   string //数据库配置
	PathSymbol string //路径分隔符
	IsWindows  bool   //win
	IsLinux    bool   //linux
}

//全局application对象
var Application application

//Date: 2022-05-09
//Desc: 获取当前系统名称
func OSName() string {
	return runtime.GOOS
}

//Date: 2022-05-09
//Parm: 路径;是否文件夹
//Desc: 判断file是否能存在
func FileExists(file string, isDir bool) bool {
	info, err := os.Stat(file)
	switch {
	case err == nil:
		if isDir {
			return info.IsDir()
		}
		return true
	case os.IsNotExist(err):
		return false
	default:
		return false
	}
}

//Date: 2022-05-09
//Parm: 文件夹路径34
//Desc: 若dir末尾没有分隔符,则添加
func FixPath(dir string) string {
	len := len(dir) - 1
	if len < 0 {
		return dir
	}

	if os.IsPathSeparator(dir[len]) {
		return dir
	} else {
		return dir + PathSeparator
	}
}

//Date: 2022-05-09
//Parm: 文件夹路径
//Desc: 创建dir目录
func MakeDir(dir string) {
	os.MkdirAll(dir, 755)
}

//Date: 2022-05-09
//Desc: 初始化
func initApp() {
	os := OSName()
	if strings.EqualFold(os, "windows") {
		PathSeparator = "\\"
	} else {
		PathSeparator = "/"
	}

	AppPath = FixPath(AppPath)
	//尾部添加路径分隔符

	Application = application{
		ExeName:    AppFile,
		ExePath:    AppPath,
		LogPath:    AppPath + "logs" + PathSeparator,
		ConfigFile: AppPath + "config.ini",
		ConfigDB:   AppPath + "db.ini",
		PathSymbol: PathSeparator,
		IsLinux:    strings.EqualFold(os, "linux"),
		IsWindows:  strings.EqualFold(os, "windows"),
	}
}
