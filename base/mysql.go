package base

import (
	"context"
	"fmt"
	"github.com/peerless6372/Lplot/env"
	"github.com/peerless6372/Lplot/klog"
	"github.com/peerless6372/Lplot/utils"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	ormUtil "gorm.io/gorm/utils"
	"lzh/gin-gonic/gin"
)

const prefix = "@@mysql."

type MysqlConf struct {
	Service         string        `yaml:"service"`
	DataBase        string        `yaml:"database"`
	Addr            string        `yaml:"addr"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	Charset         string        `yaml:"charset"`
	MaxIdleConns    int           `yaml:"maxidleconns"`
	MaxOpenConns    int           `yaml:"maxopenconns"`
	ConnMaxLifeTime time.Duration `yaml:"connMaxLifeTime"`
	ConnTimeOut     time.Duration `yaml:"connTimeOut"`
	WriteTimeOut    time.Duration `yaml:"writeTimeOut"`
	ReadTimeOut     time.Duration `yaml:"readTimeOut"`
}

func (conf *MysqlConf) checkConf() {
	env.CommonSecretChange(prefix, *conf, conf)

	if conf.MaxIdleConns == 0 {
		conf.MaxIdleConns = 10
	}
	if conf.MaxOpenConns == 0 {
		conf.MaxOpenConns = 1000
	}
	if conf.ConnMaxLifeTime == 0 {
		conf.ConnMaxLifeTime = 3600 * time.Second
	}
	if conf.ConnTimeOut == 0 {
		conf.ConnTimeOut = 3 * time.Second
	}
	if conf.WriteTimeOut == 0 {
		conf.WriteTimeOut = 1 * time.Second
	}
	if conf.ReadTimeOut == 0 {
		conf.ReadTimeOut = 1 * time.Second
	}
}

func InitMysqlClient(conf MysqlConf) (client *gorm.DB, err error) {
	conf.checkConf()

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?timeout=%s&readTimeout=%s&writeTimeout=%s&parseTime=True&loc=Asia%%2FShanghai",
		conf.User,
		conf.Password,
		conf.Addr,
		conf.DataBase,
		conf.ConnTimeOut,
		conf.ReadTimeOut,
		conf.WriteTimeOut)

	if conf.Charset != "" {
		dsn = dsn + "&charset=" + conf.Charset
	}

	c := &gorm.Config{
		SkipDefaultTransaction:                   true,
		NamingStrategy:                           nil,
		FullSaveAssociations:                     false,
		Logger:                                   newLogger(&conf),
		NowFunc:                                  nil,
		DryRun:                                   false,
		PrepareStmt:                              false,
		DisableAutomaticPing:                     false,
		DisableForeignKeyConstraintWhenMigrating: false,
		AllowGlobalUpdate:                        false,
		ClauseBuilders:                           nil,
		ConnPool:                                 nil,
		Dialector:                                nil,
		Plugins:                                  nil,
	}

	client, err = gorm.Open(mysql.Open(dsn), c)
	if err != nil {
		return client, err
	}

	sqlDB, err := client.DB()
	if err != nil {
		return client, err
	}

	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(conf.MaxIdleConns)

	// SetMaxOpenConns 设置打开数据库连接的最大数量
	sqlDB.SetMaxOpenConns(conf.MaxOpenConns)

	// SetConnMaxLifetime 设置了连接可复用的最大时间
	sqlDB.SetConnMaxLifetime(conf.ConnMaxLifeTime)

	return client, nil
}

type ormLogger struct {
	Service  string
	Addr     string
	Database string
}

func newLogger(conf *MysqlConf) logger.Interface {
	s := conf.Service
	if conf.Service == "" {
		s = conf.DataBase
	}

	return &ormLogger{
		Service:  s,
		Addr:     conf.Addr,
		Database: conf.DataBase,
	}
}

// LogMode log mode
func (l *ormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newlogger := *l
	return &newlogger
}

// Info print info
func (l ormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	m := fmt.Sprintf(msg, append([]interface{}{ormUtil.FileWithLineNum()}, data...)...)
	// 非trace日志改为debug级别输出
	klog.DebugLogger(nil, m, l.commonFields(ctx)...)
}

// Warn print warn messages
func (l ormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	m := fmt.Sprintf(msg, append([]interface{}{ormUtil.FileWithLineNum()}, data...)...)
	klog.WarnLogger(nil, m, l.commonFields(ctx)...)
}

// Error print error messages
func (l ormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	m := fmt.Sprintf(msg, append([]interface{}{ormUtil.FileWithLineNum()}, data...)...)
	klog.ErrorLogger(nil, m, l.commonFields(ctx)...)
}

func (l ormLogger) commonFields(ctx context.Context) []klog.Field {
	var logID, requestID string
	if c, ok := ctx.(*gin.Context); ok && c != nil {
		logID, _ = ctx.Value(klog.ContextKeyLogID).(string)
		requestID, _ = ctx.Value(klog.ContextKeyRequestID).(string)
	}

	fields := []klog.Field{
		klog.String(klog.TopicType, klog.LogNameModule),
		klog.String("logId", logID),
		klog.String("requestId", requestID),
		klog.String("prot", "mysql"),
		klog.String("module", env.GetAppName()),
		klog.String("service", l.Service),
		klog.String("addr", l.Addr),
		klog.String("db", l.Database),
	}
	return fields
}

// Trace print sql message
func (l ormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	end := time.Now()
	elapsed := end.Sub(begin)
	cost := float64(elapsed.Nanoseconds()/1e4) / 100.0

	// 请求是否成功
	msg := "mysql do success"
	ralCode := -0
	if err != nil {
		msg = err.Error()
		ralCode = -1
	}

	sql, rows := fc()
	fileLineNum := ormUtil.FileWithLineNum()

	fields := l.commonFields(ctx)
	fields = append(fields,
		klog.String("sql", sql),
		klog.Int64("affectedrow", rows),
		klog.String("requestEndTime", utils.GetFormatRequestTime(end)),
		klog.String("requestStartTime", utils.GetFormatRequestTime(begin)),
		klog.String("fileLine", fileLineNum),
		klog.Float64("cost", cost),
		klog.Int("ralCode", ralCode),
	)

	klog.InfoLogger(nil, msg, fields...)
}
