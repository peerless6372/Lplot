package base

import (
	"Lplot/klog"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	json "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

// default render
type DefaultRender struct {
	ErrNo  int         `json:"errNo"`
	ErrMsg string      `json:"errMsg"`
	Data   interface{} `json:"data"`
}

func setCommonHeader(ctx *gin.Context, code int, msg string) {
	ctx.Header("X_BD_UPS_ERR_NO", strconv.Itoa(code))
	ctx.Header("X_BD_UPS_ERR_MSG", msg)
	ctx.Header("Request-Id", klog.GetRequestID(ctx))
}

func RenderJson(ctx *gin.Context, code int, msg string, data interface{}) {
	setCommonHeader(ctx, code, msg)
	renderJson := DefaultRender{code, msg, data}
	ctx.JSON(http.StatusOK, renderJson)
	//ctx.Set("render", renderJson)
	return
}

func RenderJsonSucc(ctx *gin.Context, data interface{}) {
	setCommonHeader(ctx, 0, "succ")
	renderJson := DefaultRender{0, "succ", data}
	ctx.JSON(http.StatusOK, renderJson)
	//ctx.Set("render", renderJson)
	return
}

func RenderJsonFail(ctx *gin.Context, err error) {
	var renderJson DefaultRender

	switch errors.Cause(err).(type) {
	case Error:
		renderJson.ErrNo = errors.Cause(err).(Error).ErrNo
		renderJson.ErrMsg = errors.Cause(err).(Error).ErrMsg
		renderJson.Data = gin.H{}
	default:
		renderJson.ErrNo = -1
		renderJson.ErrMsg = errors.Cause(err).Error()
		renderJson.Data = gin.H{}
	}
	setCommonHeader(ctx, renderJson.ErrNo, renderJson.ErrMsg)
	ctx.JSON(http.StatusOK, renderJson)
	//ctx.Set("render", renderJson)

	// 打印错误栈
	StackLogger(ctx, err)
	return
}

func RenderJsonAbort(ctx *gin.Context, err error) {
	var renderJson DefaultRender

	switch errors.Cause(err).(type) {
	case Error:
		renderJson.ErrNo = errors.Cause(err).(Error).ErrNo
		renderJson.ErrMsg = errors.Cause(err).(Error).ErrMsg
		renderJson.Data = gin.H{}
	default:
		renderJson.ErrNo = -1
		renderJson.ErrMsg = errors.Cause(err).Error()
		renderJson.Data = gin.H{}
	}

	setCommonHeader(ctx, renderJson.ErrNo, renderJson.ErrMsg)
	ctx.AbortWithStatusJSON(http.StatusOK, renderJson)
	//ctx.Set("render", renderJson)

	return
}

// 打印错误栈
func StackLogger(ctx *gin.Context, err error) {
	if !strings.Contains(fmt.Sprintf("%+v", err), "\n") {
		return
	}

	var info []byte
	if ctx != nil {
		info, _ = json.Marshal(map[string]interface{}{"time": time.Now().Format("2006-01-02 15:04:05"), "level": "error", "module": "errorstack", "requestId": klog.GetRequestID(ctx)})
	} else {
		info, _ = json.Marshal(map[string]interface{}{"time": time.Now().Format("2006-01-02 15:04:05"), "level": "error", "module": "errorstack"})
	}

	fmt.Printf("%s\n-------------------stack-start-------------------\n%+v\n-------------------stack-end-------------------\n", string(info), err)
}
