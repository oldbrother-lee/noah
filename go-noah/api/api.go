package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func HandleSuccess(ctx *gin.Context, data interface{}) {
	if data == nil {
		data = map[string]interface{}{}
	}
	resp := Response{Code: errorCodeMap[ErrSuccess], Message: ErrSuccess.Error(), Data: data}
	if _, ok := errorCodeMap[ErrSuccess]; !ok {
		resp = Response{Code: 0, Message: "", Data: data}
	}
	ctx.JSON(http.StatusOK, resp)
}

func HandleError(ctx *gin.Context, httpCode int, err error, data interface{}) {
	if data == nil {
		data = map[string]string{}
	}

	// 默认错误码为 1（表示业务错误）
	code := 1
	msg := err.Error()

	// 尝试在 errorCodeMap 中查找自定义错误码
	func() {
		defer func() {
			if r := recover(); r != nil {
				// 错误类型不可哈希，使用默认错误码
			}
		}()
		if c, ok := errorCodeMap[err]; ok {
			code = c
		}
	}()

	resp := Response{Code: code, Message: msg, Data: data}
	ctx.JSON(httpCode, resp)
}

type Error struct {
	Code    int
	Message string
}

var errorCodeMap = map[error]int{}

func newError(code int, msg string) error {
	err := errors.New(msg)
	errorCodeMap[err] = code
	return err
}
func (e Error) Error() string {
	return e.Message
}
