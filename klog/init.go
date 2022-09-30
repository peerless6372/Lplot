package klog

import (
	"Lplot/env"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 对用户暴露的log配置
type Rotate struct {
	Switch bool   `yaml:"switch"`
	Unit   string `yaml:"unit"`
	Count  int    `yaml:"count"`
}

type Buffer struct {
	Switch        bool          `yaml:"switch"`
	Size          int           `yaml:"size"`
	FlushInterval time.Duration `yaml:"flushInterval"`
}

// 日志切分相关的log配置,仅虚拟机线上支持
type LogConfig struct {
	Level    string `yaml:"level"`
	Stdout   bool   `yaml:"stdout"`
	Log2File bool   `yaml:"log2file"`
	Path     string `yaml:"path"`
	Rotate   Rotate `yaml:"rotate"`
	Buffer   Buffer `yaml:"buffer"`
}

type loggerConfig struct {
	ZapLevel zapcore.Level

	// 以下变量仅对开发环境生效
	Stdout   bool
	Log2File bool
	Path     string

	// 日志切分相关
	RotateSwitch bool
	RotateUnit   string
	RotateCount  int

	// 缓冲区
	BufferSwitch        bool
	BufferSize          int
	BufferFlushInterval time.Duration
}

// 全局配置 仅限Init函数进行变更
var logConfig = loggerConfig{
	ZapLevel: zapcore.InfoLevel,

	Stdout:   false,
	Log2File: true,
	Path:     "./log",

	RotateUnit:   "h",
	RotateCount:  24,
	RotateSwitch: false,

	BufferSwitch:        false,
	BufferSize:          256 * 1024,
	BufferFlushInterval: 5 * time.Second,
}

func InitLog(conf LogConfig) *zap.SugaredLogger {
	if err := RegisterCsseJSONEncoder(); err != nil {
		panic(err)
	}

	logConfig.ZapLevel = getLogLevel(conf.Level)
	if env.IsDockerPlatform() {
		// 容器环境
		logConfig.Log2File = conf.Log2File
		logConfig.Path = conf.Path
		logConfig.Stdout = true
	} else {
		if _, err := os.Stat(logConfig.Path); os.IsNotExist(err) {
			err = os.MkdirAll(logConfig.Path, 0777)
			if err != nil {
				panic(fmt.Errorf("log conf err: create log dir '%s' error: %s", logConfig.Path, err))
			}
		}

		logConfig.BufferSwitch = conf.Buffer.Switch
		logConfig.BufferSize = conf.Buffer.Size
		logConfig.BufferFlushInterval = conf.Buffer.FlushInterval

		//if env.IDC != env.CloudTest {
		// 虚拟机线上
		if conf.Rotate.Switch {
			conf.Rotate.Unit = strings.ToUpper(conf.Rotate.Unit)
			if !rotateUnitValid(conf.Rotate.Unit) {
				panic("rotate unit only support 天(D/d)、小时(H/h)、分钟M(M/m)")
			}
		}

		logConfig.Log2File = conf.Log2File
		logConfig.Path = conf.Path
		logConfig.Stdout = true

		logConfig.RotateSwitch = conf.Rotate.Switch
		logConfig.RotateUnit = conf.Rotate.Unit
		logConfig.RotateCount = conf.Rotate.Count
		setRotateFlag(conf.Rotate.Switch)
	}

	SugaredLogger = GetLogger()
	return SugaredLogger
}

func setRotateFlag(logSwitch bool) {
	flagFile := path.Join(logConfig.Path, ".rotate")
	if logSwitch {
		// 在日志目录增加一个文件表示使用框架切割
		if fd, err := os.Create(flagFile); err != nil {
			panic(".rotate file create error: " + err.Error())
		} else {
			_, _ = fd.WriteString(time.Now().String())
			_ = fd.Close()
		}
	} else {
		if _, err := os.Stat(flagFile); !os.IsNotExist(err) {
			if err := os.Remove(flagFile); err != nil {
				panic(".rotate file remove error: " + err.Error())
			}
		}
	}
}

func rotateUnitValid(when string) bool {
	switch when {
	case "M", "H", "D":
		return true
	default:
		return false
	}
}
