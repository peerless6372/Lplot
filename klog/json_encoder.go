package klog

import (
	"github.com/peerless6372/Lplot/env"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

const (
	// trace 日志前缀标识（放在[]zap.Field的第一个位置提高效率）
	TopicType = "_tp"
	// 业务日志名字
	LogNameServer = "server"
	// access 日志文件名字
	LogNameAccess = "access"
	// module 日志文件名字
	LogNameModule = "module"
)

// RegisterCsseJSONEncoder registers a special jsonEncoder under "csse-json" name.
func RegisterCsseJSONEncoder() error {
	return zap.RegisterEncoder("csse-json", func(cfg zapcore.EncoderConfig) (zapcore.Encoder, error) {
		return NewCsseJSONEncoder(cfg), nil
	})
}

type jsonHexEncoder struct {
	zapcore.Encoder
}

func NewCsseJSONEncoder(cfg zapcore.EncoderConfig) zapcore.Encoder {
	jsonEncoder := zapcore.NewJSONEncoder(cfg)
	return &jsonHexEncoder{
		Encoder: jsonEncoder,
	}
}
func (enc *jsonHexEncoder) Clone() zapcore.Encoder {
	encoderClone := enc.Encoder.Clone()
	return &jsonHexEncoder{Encoder: encoderClone}
}
func (enc *jsonHexEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	// 增加 trace 日志前缀，ex： tt= tp=module.log
	fName := LogNameServer
	if len(fields) > 0 && fields[0].Key == TopicType {
		fName = fields[0].String // 确保一定是string类型的
		fields = fields[1:]
	}

	switch fName {
	case LogNameAccess:
	case LogNameModule:
	case LogNameServer:
	default:
		// 不识别的tp修改为 server
		fName = LogNameServer
	}

	buf, err := enc.Encoder.EncodeEntry(ent, fields)
	if !env.IsDockerPlatform() || buf == nil {
		return buf, err
	}

	tt := ""
	if fName == "server" {
		tt = "-notice.new"
	}
	tp := appendLogFileTail(fName, getLevelType(ent.Level))
	prefix := "tt=" + tt + " tp=" + tp + " "
	n := append([]byte(prefix), buf.Bytes()...)
	buf.Reset()
	_, _ = buf.Write(n)
	return buf, err
}

func getLevelType(lel zapcore.Level) string {
	if lel <= zapcore.InfoLevel {
		return txtLogNormal
	}
	return txtLogWarnFatal
}
