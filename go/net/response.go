package net

import (
	"GoForum/go/util"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

// Response API 回调结构体
type Response struct {
	Code     int    `json:"-"`
	Success  bool   `json:"success"`
	Msg      string `json:"msg"`
	Data     any    `json:"data"`
	Time     string `json:"time"`
	response bool
	abort    bool
	defers   []func()
}

// GetDefaultResponse 获取默认Response
func GetDefaultResponse() Response {
	return Response{
		Code:     http.StatusOK,
		Success:  true,
		Msg:      "",
		Data:     nil,
		Time:     "",
		response: true,
		abort:    false,
		defers:   []func(){},
	}
}

// ResponseDefer API响应Defer
func ResponseDefer(response *Response, context *gin.Context) {
	if !response.response {
		fmt.Println("已中止响应")
		return
	}

	response.Time = util.CurrentTimeStampStrMilli()

	defer func() {
		context.JSON(response.Code, response)
		if response.abort {
			context.Abort()
		} else {
			for _, f := range response.defers {
				f()
			}
		}
	}()

	// 捕获panic并向上抛出
	if r := recover(); r != nil {
		response.UnknownError(errors.New(fmt.Sprintf("服务器发生了Panic:%v", r)))
		panic(r)
	}

}

// UnknownError Response未知错误响应
func (response *Response) UnknownError(err error) {
	response.Success = false
	response.Msg = err.Error()
	response.Code = http.StatusInternalServerError
	//context.PostForm("e")
	//formData := context.Request.PostForm
	//for key, values := range formData {
	//	for _, value := range values {
	//		response.ErrorPostForm += "<var>" + key + ": " + value + "</var>"
	//	}
	//}
}

// Error Response错误响应
func (response *Response) Error(errMsg string) {
	response.Success = false
	response.Msg = errMsg
	response.Code = http.StatusBadRequest
}

// Successful Response成功响应
func (response *Response) Successful(data any) {
	response.Success = true
	response.Data = data
}

// AbortResponse 禁止响应
func (response *Response) AbortResponse() {
	response.response = false
}

// Abort 当Response作为中间件响应时，执行拦截操作
func (response *Response) Abort() {
	response.abort = true
}

// AppendResponseDefer 响应成功后执行的函数
func (response *Response) AppendResponseDefer(f func()) {
	response.defers = append(response.defers, f)
}
