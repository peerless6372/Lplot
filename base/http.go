package base

import (
	"Lplot/env"
	"Lplot/klog"
	"Lplot/utils"
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	json "github.com/json-iterator/go"
)

const HttpHeaderService = "SERVICE"
const HttpUrlPressureTestKey = "_call_uri"

const (
	EncodeJson = "_json"
	EncodeForm = "_form"
	EncodeRaw  = "_raw"
)

type TransportOption struct {
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	IdleConnTimeout     time.Duration
	CustomTransport     *http.Transport
}

var globalTransport *http.Transport

// 初始化全局的transport
func InitHttp(opts *TransportOption) {
	if opts == nil {
		globalTransport = &http.Transport{
			MaxIdleConns:        500,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     300 * time.Second,
		}
	} else if opts.CustomTransport != nil {
		globalTransport = opts.CustomTransport
	} else {
		globalTransport = &http.Transport{
			MaxIdleConns:        opts.MaxIdleConns,
			MaxIdleConnsPerHost: opts.MaxIdleConnsPerHost,
			IdleConnTimeout:     opts.IdleConnTimeout,
		}
	}
}

type HttpRequestOptions struct {
	// 通用请求体，可通过Encode来对body做编码
	RequestBody interface{}
	// 针对 RequestBody 的编码
	Encode string
	// 老的请求data，httpGet / httPost 仍支持
	Data map[string]string
	// httpPostJson 参数
	JsonBody interface{}
	// 请求头指定
	Headers map[string]string
	// cookie 设定
	Cookies map[string]string

	/*
		httpGet / httPost 默认 application/x-www-form-urlencoded
		httpPostJson 默认 application/json
	*/
	BodyType string
	// 重试策略，可不指定，默认使用`defaultRetryPolicy`(只有在`api.yaml`中指定retry>0 时生效)
	RetryPolicy RetryPolicy
	// 重试间隔机制，可不指定，默认使用`defaultBackOffPolicy`(只有在`api.yaml`中指定retry>0 时生效)
	BackOffPolicy BackOffPolicy
}

func (o *HttpRequestOptions) GetData() (string, error) {
	if len(o.Data) > 0 {
		return o.GetUrlData()
	}

	if o.JsonBody != nil {
		return o.GetJsonData()
	}

	// 以上两种兼容老的使用Data和JsonBody传参的情况，以下是使用RequestBody传参的解析
	return o.GetRequestData()
}
func (o *HttpRequestOptions) GetContentType() (cType string) {
	switch o.Encode {
	case EncodeJson:
		cType = "application/json"
	case EncodeForm: // 默认Form编码方式
		fallthrough
	default:
		cType = "application/x-www-form-urlencoded"
	}
	return cType
}
func (o *HttpRequestOptions) GetRequestData() (encodeData string, err error) {
	if o.RequestBody == nil {
		return encodeData, nil
	}

	switch o.Encode {
	case EncodeJson:
		reqBody, e := json.Marshal(o.RequestBody)
		encodeData, err = string(reqBody), e
	case EncodeRaw:
		ok := true
		encodeData, ok = o.RequestBody.(string)
		if !ok {
			err = errors.New("raw data need string type")
		}
	case EncodeForm: // 由于历史原因，默认Form编码方式
		fallthrough
	default:
		v := url.Values{}
		if data, ok := o.RequestBody.(map[string]string); ok {
			for key, value := range data {
				v.Add(key, value)
			}
		} else if data, ok := o.RequestBody.(map[string]interface{}); ok {
			for key, value := range data {
				var vStr string
				switch value.(type) {
				case string:
					vStr = value.(string)
				default:
					if tmp, err := json.Marshal(value); err != nil {
						return encodeData, err
					} else {
						vStr = string(tmp)
					}
				}
				v.Add(key, vStr)
			}
		} else {
			return encodeData, errors.New("unSupport RequestBody type")
		}
		encodeData, err = v.Encode(), nil
	}
	return encodeData, err
}
func (o *HttpRequestOptions) GetUrlData() (string, error) {
	v := url.Values{}
	if len(o.Data) > 0 {
		for key, value := range o.Data {
			v.Add(key, value)
		}
	}
	return v.Encode(), nil
}
func (o *HttpRequestOptions) GetJsonData() (string, error) {
	reqBody, err := json.Marshal(o.JsonBody)
	return string(reqBody), err
}
func (o *HttpRequestOptions) GetRetryPolicy() RetryPolicy {
	r := defaultRetryPolicy
	if o.RetryPolicy != nil {
		r = o.RetryPolicy
	}
	return r
}
func (o *HttpRequestOptions) GetBackOffPolicy() BackOffPolicy {
	b := defaultBackOffPolicy
	if o.BackOffPolicy != nil {
		b = o.BackOffPolicy
	}

	return b
}

type ApiClient struct {
	Service        string        `yaml:"service"`
	AppKey         string        `yaml:"appkey"`
	AppSecret      string        `yaml:"appsecret"`
	Domain         string        `yaml:"domain"`
	Timeout        time.Duration `yaml:"timeout"`
	ConnectTimeout time.Duration `yaml:"connectTimeout"`
	Retry          int           `yaml:"retry"`
	HttpStat       bool          `yaml:"httpStat"`
	Host           string        `yaml:"host"`
	Proxy          string        `yaml:"proxy"`
	BasicAuth      struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	}

	HTTPClient *http.Client
	clientInit sync.Once
}

func (client *ApiClient) GetTransPort() *http.Transport {
	trans := globalTransport
	if client.Proxy != "" {
		trans.Proxy = func(_ *http.Request) (*url.URL, error) {
			return url.Parse(client.Proxy)
		}
	} else {
		trans.Proxy = nil
	}

	if client.ConnectTimeout != 0 {
		trans.DialContext = (&net.Dialer{
			Timeout: client.ConnectTimeout,
		}).DialContext
	} else {
		trans.DialContext = nil
	}

	return trans
}

func (client *ApiClient) makeRequest(ctx *gin.Context, method, url string, data io.Reader, opts HttpRequestOptions) (*http.Request, error) {
	req, err := http.NewRequest(method, url, data)
	if err != nil {
		return nil, err
	}

	if opts.Headers != nil {
		for k, v := range opts.Headers {
			req.Header.Set(k, v)
		}
	}

	if client.Host != "" {
		req.Host = client.Host
	} else if h := req.Header.Get("host"); h != "" {
		req.Host = h
	}

	for k, v := range opts.Cookies {
		req.AddCookie(&http.Cookie{
			Name:  k,
			Value: v,
		})
	}

	if client.BasicAuth.Username != "" {
		req.SetBasicAuth(client.BasicAuth.Username, client.BasicAuth.Password)
	}

	cType := opts.BodyType
	if cType == "" { // 根据 encode 获得一个默认的类型
		cType = opts.GetContentType()
	}
	req.Header.Set("Content-Type", cType)

	req.Header.Set(HttpHeaderService, env.AppName)
	req.Header.Set(klog.TraceHeaderKey, klog.GetRequestID(ctx))

	return req, nil
}

func (client *ApiClient) HttpGet(ctx *gin.Context, path string, opts HttpRequestOptions) (*ApiResult, error) {
	// http request
	urlData, err := opts.GetData()
	if err != nil {
		klog.WarnLogger(ctx, "http client make data error: "+err.Error(), klog.String(klog.TopicType, klog.LogNameModule))
		return nil, err
	}

	domain := client.Domain

	var u string
	if urlData == "" {
		u = fmt.Sprintf("%s%s", domain, path)
	} else {
		u = fmt.Sprintf("%s%s?%s", domain, path, urlData)
	}

	req, err := client.makeRequest(ctx, "GET", u, nil, opts)
	if err != nil {
		klog.WarnLogger(ctx, "http client makeRequest error: "+err.Error(), klog.String(klog.TopicType, klog.LogNameModule))
		return nil, err
	}

	t := client.beforeHttpStat(ctx, req)
	body, fields, err := client.httpDo(ctx, req, &opts)
	client.afterHttpStat(ctx, req.URL.Scheme, t)

	klog.DebugLogger(ctx, "http get request",
		klog.String(klog.TopicType, klog.LogNameModule),
		klog.String("url", u),
		klog.Int("responseCode", body.HttpCode),
		klog.String("responseBody", string(body.Response)),
	)
	msg := "http request success"
	if err != nil {
		msg = err.Error()
	}

	klog.InfoLogger(ctx, msg, fields...)

	return &body, err
}

func (client *ApiClient) HttpPost(ctx *gin.Context, path string, opts HttpRequestOptions) (*ApiResult, error) {
	// http request
	urlData, err := opts.GetData()
	if err != nil {
		klog.WarnLogger(ctx, "http client make data error: "+err.Error(), klog.String(klog.TopicType, klog.LogNameModule))
		return nil, err
	}

	domain := client.Domain

	u := fmt.Sprintf("%s%s", domain, path)

	req, err := client.makeRequest(ctx, "POST", u, strings.NewReader(urlData), opts)
	if err != nil {
		klog.WarnLogger(ctx, "http client makeRequest error: "+err.Error(), klog.String(klog.TopicType, klog.LogNameModule))
		return nil, err
	}

	t := client.beforeHttpStat(ctx, req)
	body, fields, err := client.httpDo(ctx, req, &opts)
	client.afterHttpStat(ctx, req.URL.Scheme, t)

	klog.DebugLogger(ctx, "http post request",
		klog.String(klog.TopicType, klog.LogNameModule),
		klog.String("url", u),
		klog.String("params", urlData),
		klog.Int("responseCode", body.HttpCode),
		klog.String("responseBody", string(body.Response)),
	)

	msg := "http request success"
	if err != nil {
		msg = err.Error()
	}

	klog.InfoLogger(ctx, msg, fields...)

	return &body, err
}

// deprecated , use HttpPost instead
func (client *ApiClient) HttpPostJson(ctx *gin.Context, path string, opts HttpRequestOptions) (*ApiResult, error) {
	urlData, err := opts.GetJsonData()
	if err != nil {
		klog.WarnLogger(ctx, "http client make data error: "+err.Error(), klog.String(klog.TopicType, klog.LogNameModule))
		return nil, err
	}

	domain := client.Domain

	u := fmt.Sprintf("%s%s", domain, path)

	opts.BodyType = EncodeJson
	req, err := client.makeRequest(ctx, "POST", u, strings.NewReader(urlData), opts)
	if err != nil {
		klog.WarnLogger(ctx, "http client makeRequest error: "+err.Error(), klog.String(klog.TopicType, klog.LogNameModule))
		return nil, err
	}

	t := client.beforeHttpStat(ctx, req)
	body, fields, err := client.httpDo(ctx, req, &opts)
	client.afterHttpStat(ctx, req.URL.Scheme, t)

	klog.DebugLogger(ctx, "HttpPostJson",
		klog.String(klog.TopicType, klog.LogNameModule),
		klog.String("url", u),
		klog.String("params", urlData),
		klog.Int("responseCode", body.HttpCode),
		klog.String("responseBody", string(body.Response)),
	)

	msg := "http request success"
	if err != nil {
		msg = err.Error()
	}
	klog.InfoLogger(ctx, msg, fields...)

	return &body, err
}

type ApiResult struct {
	HttpCode int
	Response []byte
	Ctx      *gin.Context
}

func (client *ApiClient) httpDo(ctx *gin.Context, req *http.Request, opts *HttpRequestOptions) (res ApiResult, field []klog.Field, err error) {
	start := time.Now()
	fields := []klog.Field{
		klog.String(klog.TopicType, klog.LogNameModule),
		klog.String("prot", "http"),
		klog.String("service", client.Service),
		klog.String("method", req.Method),
		klog.String("domain", client.Domain),
		klog.String("requestUri", req.URL.Path),
		klog.String("proxy", client.Proxy),
		klog.Duration("timeout", client.Timeout),
		klog.String("requestStartTime", utils.GetFormatRequestTime(start)),
	}

	client.clientInit.Do(func() {
		if client.HTTPClient == nil {
			timeout := 3 * time.Second
			if client.Timeout > 0 {
				timeout = client.Timeout
			}

			trans := client.GetTransPort()
			client.HTTPClient = &http.Client{
				Timeout:   timeout,
				Transport: trans,
			}
		}
	})

	var (
		resp         *http.Response
		dataBuffer   *bytes.Reader
		maxAttempts  int
		attemptCount int
		doErr        error
		shouldRetry  bool
	)

	attemptCount, maxAttempts = 0, client.Retry

	retryPolicy := opts.GetRetryPolicy()
	backOffPolicy := opts.GetBackOffPolicy()

	for {
		if req.GetBody != nil {
			bodyReadCloser, _ := req.GetBody()
			req.Body = bodyReadCloser
		} else if req.Body != nil {
			if dataBuffer == nil {
				data, err := ioutil.ReadAll(req.Body)
				_ = req.Body.Close()
				if err != nil {
					return res, fields, err
				}
				dataBuffer = bytes.NewReader(data)
				req.ContentLength = int64(dataBuffer.Len())
				req.Body = ioutil.NopCloser(dataBuffer)
			}
			_, _ = dataBuffer.Seek(0, io.SeekStart)
		}

		attemptCount++
		resp, doErr = client.HTTPClient.Do(req)
		if doErr != nil {
			f := []klog.Field{
				klog.String(klog.TopicType, klog.LogNameModule),
				klog.String("prot", "http"),
				klog.String("service", client.Service),
				klog.String("requestUri", req.URL.Path),
				klog.Duration("timeout", client.Timeout),
				klog.Int("attemptCount", attemptCount),
			}
			klog.WarnLogger(ctx, doErr.Error(), f...)
		}

		shouldRetry = retryPolicy(resp, doErr)
		if !shouldRetry {
			break
		}

		if attemptCount > maxAttempts {
			break
		}

		if doErr == nil {
			drainAndCloseBody(resp, 16384)
		}
		wait := backOffPolicy(attemptCount)
		select {
		case <-req.Context().Done():
			return res, fields, req.Context().Err()
		case <-time.After(wait):
		}
	}

	if resp != nil {
		res.HttpCode = resp.StatusCode
		res.Response, err = ioutil.ReadAll(resp.Body)
		_ = resp.Body.Close()
	}

	err = doErr
	if err == nil && shouldRetry {
		err = fmt.Errorf("hit retry policy")
	}

	end := time.Now()
	if err != nil {
		err = fmt.Errorf("giving up after %d attempt(s): %w", attemptCount, err)
	}

	fields = append(fields,
		klog.String("retry", fmt.Sprintf("%d/%d", attemptCount-1, client.Retry)),
		klog.Int("httpCode", res.HttpCode),
		klog.String("requestEndTime", utils.GetFormatRequestTime(end)),
		klog.Float64("cost", utils.GetRequestCost(start, end)),
		klog.Int("ralCode", client.calRalCode(resp, err)),
	)

	return res, fields, err
}

// 本次请求正确性判断
func (client *ApiClient) calRalCode(resp *http.Response, err error) int {
	if err != nil || resp == nil || resp.StatusCode >= 400 || resp.StatusCode == 0 {
		return -1
	}
	return 0
}

type timeTrace struct {
	dnsStartTime,
	dnsDoneTime,
	connectDoneTime,
	gotConnTime,
	gotFirstRespTime,
	tlsHandshakeStartTime,
	tlsHandshakeDoneTime,
	finishTime time.Time
}

func (client *ApiClient) beforeHttpStat(ctx *gin.Context, req *http.Request) *timeTrace {
	if client.HttpStat == false {
		return nil
	}

	var t = &timeTrace{}
	trace := &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) { t.dnsStartTime = time.Now() },
		DNSDone:  func(_ httptrace.DNSDoneInfo) { t.dnsDoneTime = time.Now() },
		ConnectStart: func(_, _ string) {
			if t.dnsDoneTime.IsZero() {
				t.dnsDoneTime = time.Now()
			}
		},
		ConnectDone: func(net, addr string, err error) {
			t.connectDoneTime = time.Now()
		},
		GotConn:              func(_ httptrace.GotConnInfo) { t.gotConnTime = time.Now() },
		GotFirstResponseByte: func() { t.gotFirstRespTime = time.Now() },
		TLSHandshakeStart:    func() { t.tlsHandshakeStartTime = time.Now() },
		TLSHandshakeDone:     func(_ tls.ConnectionState, _ error) { t.tlsHandshakeDoneTime = time.Now() },
	}
	*req = *req.WithContext(httptrace.WithClientTrace(context.Background(), trace))
	return t
}

func (client *ApiClient) afterHttpStat(ctx *gin.Context, scheme string, t *timeTrace) {
	if client.HttpStat == false {
		return
	}
	t.finishTime = time.Now() // after read body

	if t.dnsStartTime.IsZero() {
		t.dnsStartTime = t.dnsDoneTime
	}

	cost := func(d time.Duration) float64 {
		if d < 0 {
			return -1
		}
		return float64(d.Nanoseconds()/1e4) / 100.0
	}

	switch scheme {
	case "https":
		f := []klog.Field{
			klog.String(klog.TopicType, klog.LogNameModule),
			klog.Float64("dnsLookupCost", cost(t.dnsDoneTime.Sub(t.dnsStartTime))),                       // dns lookup
			klog.Float64("tcpConnectCost", cost(t.connectDoneTime.Sub(t.dnsDoneTime))),                   // tcp connection
			klog.Float64("tlsHandshakeCost", cost(t.tlsHandshakeStartTime.Sub(t.tlsHandshakeStartTime))), // tls handshake
			klog.Float64("serverProcessCost", cost(t.gotFirstRespTime.Sub(t.gotConnTime))),               // server processing
			klog.Float64("contentTransferCost", cost(t.finishTime.Sub(t.gotFirstRespTime))),              // content transfer
			klog.Float64("totalCost", cost(t.finishTime.Sub(t.dnsStartTime))),                            // total cost
		}
		klog.InfoLogger(ctx, "time trace", f...)
	case "http":
		f := []klog.Field{
			klog.String(klog.TopicType, klog.LogNameModule),
			klog.Float64("dnsLookupCost", cost(t.dnsDoneTime.Sub(t.dnsStartTime))),          // dns lookup
			klog.Float64("tcpConnectCost", cost(t.gotConnTime.Sub(t.dnsDoneTime))),          // tcp connection
			klog.Float64("serverProcessCost", cost(t.gotFirstRespTime.Sub(t.gotConnTime))),  // server processing
			klog.Float64("contentTransferCost", cost(t.finishTime.Sub(t.gotFirstRespTime))), // content transfer
			klog.Float64("totalCost", cost(t.finishTime.Sub(t.dnsStartTime))),               // total cost
		}
		klog.InfoLogger(ctx, "time trace", f...)
	}
}

func drainAndCloseBody(resp *http.Response, maxBytes int64) {
	if resp != nil {
		_, _ = io.CopyN(ioutil.Discard, resp.Body, maxBytes)
		_ = resp.Body.Close()
	}
}

// retry 策略
type RetryPolicy func(resp *http.Response, err error) bool

var defaultRetryPolicy RetryPolicy = func(resp *http.Response, err error) bool {
	return err != nil || resp == nil || resp.StatusCode >= 500 || resp.StatusCode == 0
}

// 重试策略
type BackOffPolicy func(attemptCount int) time.Duration

var defaultBackOffPolicy = func(attemptNum int) time.Duration { // retry immediately
	return 0
}
