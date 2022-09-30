package middleware

import (
	"Lplot/base"
	"Lplot/klog"
	"Lplot/utils"
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

const (
	printRequestLen  = 10240
	printResponseLen = 10240
)

var (
	// 暂不需要，后续考虑看是否需要支持用户配置
	mcpackReqUris []string
	ignoreReqUris []string
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) WriteString(s string) (int, error) {
	s = strings.Replace(s, "\n", "", -1)
	if w.body != nil {
		w.body.WriteString(s)
	}
	return w.ResponseWriter.WriteString(s)
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	if w.body != nil {
		//idx := len(b)
		// gin render json 后后面会多一个换行符
		//if b[idx-1] == '\n' {
		//	b = b[:idx-1]
		//}
		w.body.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

// access日志打印
func AccessLog() gin.HandlerFunc {
	// 当前模块名
	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()
		// 请求url
		path := c.Request.URL.Path
		// 请求报文
		var requestBody []byte
		if c.Request.Body != nil {
			var err error
			requestBody, err = c.GetRawData()
			if err != nil {
				klog.Warnf(c, "get http request body error: %s", err.Error())
			}
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))
		}

		blw := new(bodyLogWriter)
		if printResponseLen <= 0 {
			blw = &bodyLogWriter{body: nil, ResponseWriter: c.Writer}
		} else {
			blw = &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		}
		c.Writer = blw

		c.Set(klog.ContextKeyUri, path)
		_ = klog.GetRequestID(c)
		// 处理请求
		c.Next()

		response := ""
		if blw.body != nil {
			if len(blw.body.String()) <= printResponseLen {
				response = blw.body.String()
			} else {
				response = blw.body.String()[:printResponseLen]
			}
		}

		bodyStr := ""
		flag := false
		// macpack的请求，以二进制输出日志
		for _, val := range mcpackReqUris {
			if strings.Contains(path, val) {
				bodyStr = fmt.Sprintf("%v", requestBody)
				flag = true
				break
			}
		}
		if !flag {
			// 不打印RequestBody的请求
			for _, val := range ignoreReqUris {
				if strings.Contains(path, val) {
					bodyStr = ""
					flag = true
					break
				}
			}
		}
		if !flag {
			bodyStr = string(requestBody)
		}

		if c.Request.URL.RawQuery != "" {
			bodyStr += "&" + c.Request.URL.RawQuery
		}

		if len(bodyStr) > printRequestLen {
			bodyStr = bodyStr[:printRequestLen]
		}

		// 结束时间
		end := time.Now()

		// 用户自定义notice
		var customerFields []klog.Field
		for k, v := range klog.GetCustomerKeyValue(c) {
			customerFields = append(customerFields, klog.Reflect(k, v))
		}

		// 固定notice
		commonFields := []klog.Field{
			klog.String("cuid", getReqValueByKey(c, "cuid")),
			klog.String("device", getReqValueByKey(c, "device")),
			klog.String("channel", getReqValueByKey(c, "channel")),
			klog.String("os", getReqValueByKey(c, "os")),
			klog.String("vc", getReqValueByKey(c, "vc")),
			klog.String("vcname", getReqValueByKey(c, "vcname")),
			klog.String("userid", getReqValueByKey(c, "userid")),
			klog.String("uri", c.Request.RequestURI),
			klog.String("host", c.Request.Host),
			klog.String("method", c.Request.Method),
			klog.String("httpProto", c.Request.Proto),
			klog.String("handle", c.HandlerName()),
			klog.String("userAgent", c.Request.UserAgent()),
			klog.String("refer", c.Request.Referer()),
			klog.String("clientIp", utils.GetClientIp(c)),
			klog.String("cookie", getCookie(c)),
			klog.String("requestStartTime", utils.GetFormatRequestTime(start)),
			klog.String("requestEndTime", utils.GetFormatRequestTime(end)),
			klog.Float64("cost", utils.GetRequestCost(start, end)),
			klog.String("requestParam", bodyStr),
			klog.Int("responseStatus", c.Writer.Status()),
			klog.String("response", response),
		}

		commonFields = append(commonFields, customerFields...)
		klog.InfoLogger(c, "notice", commonFields...)
	}
}

// 从request body中解析特定字段作为notice key打印
func getReqValueByKey(ctx *gin.Context, k string) string {
	if vs, exist := ctx.Request.Form[k]; exist && len(vs) > 0 {
		return vs[0]
	}
	return ""
}

func getCookie(ctx *gin.Context) string {
	cStr := ""
	for _, c := range ctx.Request.Cookies() {
		cStr += fmt.Sprintf("%s=%s&", c.Name, c.Value)
	}
	return strings.TrimRight(cStr, "&")
}

// access 添加kv打印
func AddNotice(k string, v interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		klog.AddNotice(c, k, v)
		c.Next()
	}
}

func LoggerBeforeRun(ctx *gin.Context) {
	customCtx := ctx.CustomContext
	fields := []klog.Field{
		klog.String("handle", customCtx.HandlerName()),
		klog.String("type", customCtx.Type),
	}

	klog.InfoLogger(ctx, "start", fields...)
}

func LoggerAfterRun(ctx *gin.Context) {
	customCtx := ctx.CustomContext
	cost := utils.GetRequestCost(customCtx.StartTime, customCtx.EndTime)
	var err error
	if customCtx.Error != nil {
		err = errors.Cause(customCtx.Error)
		base.StackLogger(ctx, customCtx.Error)
	}

	// 用户自定义notice
	notices := klog.GetCustomerKeyValue(ctx)

	var fields []klog.Field
	for k, v := range notices {
		fields = append(fields, klog.Reflect(k, v))
	}

	fields = append(fields,
		klog.String("handle", customCtx.HandlerName()),
		klog.String("type", customCtx.Type),
		klog.Float64("cost", cost),
		klog.String("error", fmt.Sprintf("%+v", err)),
	)

	klog.InfoLogger(ctx, "end", fields...)
}
