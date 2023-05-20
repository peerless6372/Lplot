package env

import (
	"github.com/wyywawj1991/goframework/utils"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

const DefaultRootPath = "."

const (
	// 容器中的环境变量
	CsseClusterType = "CSSE_CLUSTER_TYPE"
	DockAppName     = "APP_NAME"
	DockerRunEnv    = "RUN_ENV"
)

// RUN_ENV： (prod，tips，test)
const (
	RunEnvTest   = 0
	RunEnvTips   = 1
	RunEnvOnline = 2
)

var (
	LocalIP string
	AppName string
	RunMode string

	runEnv int

	rootPath       string
	dockerPlatform bool

	Hostname string
)

func init() {
	LocalIP = utils.GetLocalIp()
	dockerPlatform = false
	if r := os.Getenv(CsseClusterType); r != "" {
		dockerPlatform = true
		// 容器里，appName在编排的时候决定
		if n := os.Getenv(DockAppName); n != "" {
			AppName = n
			println("docker env, APP_NAME=", n)
		} else {
			println("docker env, lack APP_NAME!!!")
		}
	} else {
		var err error
		if Hostname, err = os.Hostname(); err != nil || Hostname == "" {
			Hostname = os.Getenv("HOSTNAME")
		}
	}

	// 运行环境
	RunMode = gin.ReleaseMode
	r := os.Getenv(DockerRunEnv)
	switch r {
	case "prod":
		runEnv = RunEnvOnline
	case "tips":
		runEnv = RunEnvTips
	default:
		runEnv = RunEnvTest
		RunMode = gin.DebugMode
	}

	gin.SetMode(RunMode)

	initDBSecret()
}

// 判断项目运行平台：容器 vs 开发环境
func IsDockerPlatform() bool {
	return dockerPlatform
}

// 开发环境可手动指定SetAppName
func SetAppName(appName string) {
	if !dockerPlatform {
		AppName = appName
	}
}

func GetAppName() string {
	return AppName
}

// SetRootPath 设置应用的根目录
func SetRootPath(r string) {
	if !dockerPlatform {
		rootPath = r
	}
}

// RootPath 返回应用的根目录
func GetRootPath() string {
	if rootPath != "" {
		return rootPath
	} else {
		return DefaultRootPath
	}
}

// GetConfDirPath 返回配置文件目录绝对地址
func GetConfDirPath() string {
	return filepath.Join(GetRootPath(), "conf")
}

// LogRootPath 返回log目录的绝对地址
func GetLogDirPath() string {
	return filepath.Join(GetRootPath(), "log")
}

func GetRunEnv() int {
	return runEnv
}
