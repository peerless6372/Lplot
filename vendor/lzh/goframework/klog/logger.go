package klog

import (
	"github.com/wyywawj1991/goframework/env"
	"io"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Field = zap.Field

var (
	Binary = zap.Binary
	Bool   = zap.Bool

	ByteString = zap.ByteString
	String     = zap.String
	Strings    = zap.Strings

	Float64 = zap.Float64
	Float32 = zap.Float32

	Int   = zap.Int
	Int64 = zap.Int64
	Int32 = zap.Int32
	Int16 = zap.Int16
	Int8  = zap.Int8

	Uint   = zap.Uint
	Uint64 = zap.Uint64
	Uint32 = zap.Uint32

	Reflect   = zap.Reflect
	Namespace = zap.Namespace
	Duration  = zap.Duration
	Object    = zap.Object
	Any       = zap.Any
	Skip      = zap.Skip()
)
var (
	SugaredLogger *zap.SugaredLogger
	ZapLogger     *zap.Logger
)

// log文件后缀类型
const (
	txtLogNormal    = "normal"
	txtLogWarnFatal = "warnfatal"
	txtLogStdout    = "stdout"
)

// NewLogger 新建Logger，每一次新建会同时创建x.log与x.log.wf (access.log 不会生成wf)
func newLogger() *zap.Logger {
	var infoLevel = zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= logConfig.ZapLevel && lvl <= zapcore.InfoLevel
	})

	var errorLevel = zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= logConfig.ZapLevel && lvl >= zapcore.WarnLevel
	})

	var stdLevel = zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= logConfig.ZapLevel && lvl >= zapcore.DebugLevel
	})

	name := env.AppName
	if name == "" {
		name = "server"
	}
	var zapCore []zapcore.Core
	if logConfig.Stdout {
		c := zapcore.NewCore(
			getEncoder(),
			zapcore.AddSync(getLogWriter(name, txtLogStdout)),
			stdLevel)
		zapCore = append(zapCore, c)
	}

	// 仅开发环境有效，便于开发调试
	if logConfig.Log2File {
		zapCore = append(zapCore,
			zapcore.NewCore(
				getEncoder(),
				zapcore.AddSync(getLogWriter(name, txtLogNormal)),
				infoLevel))

		zapCore = append(zapCore,
			zapcore.NewCore(
				getEncoder(),
				zapcore.AddSync(getLogWriter(name, txtLogWarnFatal)),
				errorLevel))
	}

	// core
	core := zapcore.NewTee(zapCore...)

	// 开启开发模式，堆栈跟踪
	caller := zap.WithCaller(true)

	// 由于之前没有DPanic，同化DPanic和Panic
	development := zap.Development()

	// 设置初始化字段
	filed := zap.Fields()

	// 构造日志
	logger := zap.New(core, filed, caller, development)

	return logger
}

func getLogLevel(lv string) (level zapcore.Level) {
	str := strings.ToUpper(lv)
	switch str {
	case "DEBUG":
		level = zap.DebugLevel
	case "INFO":
		level = zap.InfoLevel
	case "WARN":
		level = zap.WarnLevel
	case "ERROR":
		level = zap.ErrorLevel
	case "FATAL":
		level = zap.FatalLevel
	default:
		level = zap.InfoLevel
	}
	return level
}

func getEncoder() zapcore.Encoder {
	// time字段编码器
	timeEncoder := zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.999999")

	encoderCfg := zapcore.EncoderConfig{
		LevelKey:       "level",
		TimeKey:        "time",
		CallerKey:      "file",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeCaller:   zapcore.ShortCallerEncoder, // 短路径编码器
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     timeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}
	return NewCsseJSONEncoder(encoderCfg)
}

func getLogWriter(name, loggerType string) (ws zapcore.WriteSyncer) {
	var w io.Writer
	if loggerType == txtLogStdout {
		// stdOut
		w = os.Stdout
	} else {
		// 打印到 name.log 中
		w = NewTimeFileLogWriter(appendLogFileTail(name, loggerType))
	}

	ws = zapcore.AddSync(w)

	if logConfig.BufferSwitch {
		// 带缓冲区
		ws = &zapcore.BufferedWriteSyncer{
			WS:            zapcore.AddSync(w),
			Size:          logConfig.BufferSize,
			FlushInterval: logConfig.BufferFlushInterval,
			Clock:         nil,
		}
	}

	return ws
}

// genFilename 拼装完整文件名
func appendLogFileTail(appName, loggerType string) string {
	var tailFixed string
	switch loggerType {
	case txtLogNormal:
		tailFixed = ".log"
	case txtLogWarnFatal:
		tailFixed = ".log.wf"
	default:
		tailFixed = ".log"
	}
	return appName + tailFixed
}

func CloseLogger() {
	if SugaredLogger != nil {
		_ = SugaredLogger.Sync()
	}

	if ZapLogger != nil {
		_ = ZapLogger.Sync()
	}
}

// -------------避免用户改动过大，以下为封装的之前的Entry打印field的方法----------
type Fields map[string]interface{}
type entry struct {
	s *zap.SugaredLogger
}

func NewEntry(s *zap.SugaredLogger) *entry {
	x := s.Desugar().WithOptions(zap.AddCallerSkip(+1)).Sugar()
	return &entry{s: x}
}

// 注意这种使用方式固定头的顺序会变
func (e entry) WithFields(f Fields) *zap.SugaredLogger {
	var fields []interface{}
	for k, v := range f {
		fields = append(fields, k, v)
	}

	return e.s.With(fields...)
}
