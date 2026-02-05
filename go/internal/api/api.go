package api

import (
	"aigcpanel/go/internal/ask"
	"aigcpanel/go/internal/errs"
	"github.com/gin-gonic/gin"
	"net/http"
	"reflect"
	"strings"
)

// Err 错误处理
func Err(ctx *gin.Context, err error) {
	switch err.(type) {
	case *errs.HTTPException:
		ctx.JSON(http.StatusOK, ask.R{
			Message: err.Error(),
			Data:    nil,
			Code:    err.(*errs.HTTPException).Code,
		})
	default:
		ctx.JSON(http.StatusOK, ask.R{
			Message: errs.SystemError.Message,
			Data:    nil,
			Code:    errs.SystemError.Code,
		})
	}
	ctx.Abort()
	return
}

// ErrWithMessage ...
func ErrWithMessage(ctx *gin.Context, err error, message string) {
	switch err.(type) {
	case *errs.HTTPException:
		ctx.JSON(http.StatusOK, ask.R{
			Message: err.Error(),
			Data:    nil,
			Code:    err.(*errs.HTTPException).Code,
		})
	default:
		responseMessage := errs.SystemError.Message
		if message != "" {
			responseMessage = message
		}
		ctx.JSON(http.StatusOK, ask.R{
			Message: responseMessage,
			Data:    nil,
			Code:    errs.SystemError.Code,
		})
	}
	ctx.Abort()
	return
}

// OK 成功处理
func OK(ctx *gin.Context, data ...interface{}) {
	ctx.JSON(http.StatusOK, ask.Success(data...))
}

// Message ...
func Message(ctx *gin.Context, message string, data ...interface{}) {
	ctx.JSON(http.StatusOK, ask.Message(message, data...))
}

// FieldHide 隐藏字段
func FieldHide(obj interface{}, fields ...string) map[string]interface{} {
	data := struct2map(obj)
	for _, v := range fields {
		delete(data, v)
	}
	return data
}

func struct2map(obj interface{}) map[string]interface{} {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	var data = make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldName := ""
		if jsonTag, ok := field.Tag.Lookup("json"); ok {
			jsonInfo := strings.Split(jsonTag, ",")
			fieldName = jsonInfo[0]
		} else {
			fieldName = t.Field(i).Name
		}
		data[fieldName] = v.Field(i).Interface()
	}
	return data
}
