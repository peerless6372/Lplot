package base

import (
	"Lplot/env"
	"Lplot/klog"
	"net/http"
	"strings"

	"github.com/olivere/elastic"
	"go.uber.org/zap"
)

const esPrefix = "@@es."

type ElasticClientConfig struct {
	Addr     string `yaml:"addr"`
	Service  string `yaml:"service"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`

	Sniff       bool `yaml:"sniff"`
	HealthCheck bool `yaml:"healthCheck"`
	Gzip        bool `yaml:"gzip"`

	Decoder       elastic.Decoder
	RetryStrategy elastic.Retrier
	HttpClient    *http.Client
	Other         []elastic.ClientOptionFunc
}

func (conf *ElasticClientConfig) checkConfig() {
	env.CommonSecretChange(esPrefix, *conf, conf)
}

func NewESClient(cfg ElasticClientConfig) (*elastic.Client, error) {
	cfg.checkConfig()

	addrs := strings.Split(cfg.Addr, ",")
	options := []elastic.ClientOptionFunc{
		elastic.SetURL(addrs...),
		elastic.SetSniff(cfg.Sniff),
		elastic.SetHealthcheck(cfg.HealthCheck),
		elastic.SetGzip(cfg.Gzip),
	}

	esLogger := newEsLogger()
	options = append(options,
		elastic.SetTraceLog(&elasticDebugLogger{esLogger}),
		elastic.SetInfoLog(&elasticInfoLogger{esLogger}),
		elastic.SetErrorLog(&elasticErrorLogger{esLogger}))

	if cfg.Username != "" || cfg.Password != "" {
		options = append(options, elastic.SetBasicAuth(cfg.Username, cfg.Password))
	}

	if cfg.HttpClient != nil {
		options = append(options, elastic.SetHttpClient(cfg.HttpClient))
	}
	if cfg.Decoder != nil {
		options = append(options, elastic.SetDecoder(cfg.Decoder))
	}

	if cfg.RetryStrategy != nil {
		options = append(options, elastic.SetRetrier(cfg.RetryStrategy))
	}

	// override
	if len(cfg.Other) > 0 {
		options = append(options, cfg.Other...)
	}

	return elastic.NewClient(options...)
}

type elasticLogger struct {
	logger *zap.SugaredLogger
}

func newEsLogger() elasticLogger {
	return elasticLogger{
		logger: klog.GetZapLogger().
			With(
				klog.String(klog.TopicType, klog.LogNameModule),
				klog.String("prot", "es"),
				klog.String("localIp", env.LocalIP),
				klog.String("module", env.GetAppName()),
			).Sugar(),
	}
}

type elasticDebugLogger struct {
	elasticLogger
}
type elasticInfoLogger struct {
	elasticLogger
}
type elasticErrorLogger struct {
	elasticLogger
}

func (l *elasticDebugLogger) Printf(format string, v ...interface{}) {
	l.logger.Debugf(format, v...)
}

func (l *elasticInfoLogger) Printf(format string, v ...interface{}) {
	l.logger.Infof(format, v...)
}

func (l *elasticErrorLogger) Printf(format string, v ...interface{}) {
	l.logger.Errorf(format, v...)
}
