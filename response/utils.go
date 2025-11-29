package response

import (
	"github.com/qbhy/goal-admin/enums"
	"github.com/qbhy/goal-admin/results"

	"github.com/goal-web/contracts"
)

func ParseReqErr(err error) any {
	return results.ResponseResult{
		Message:    err.Error(),
		Code:       int32(enums.CodeParseReqErr),
		ErrMessage: enums.CodeParseReqErr.Message(),
		Data:       nil,
	}
}

func InvalidReq(err error) any {
	return results.ResponseResult{
		Message:    err.Error(),
		Code:       int32(enums.CodeParseReqErr),
		ErrMessage: enums.CodeParseReqErr.Message(),
		Data:       nil,
	}
}

func BizErr(err error) any {
	return results.ResponseResult{
		Message:    err.Error(),
		Code:       int32(enums.CodeBizErr),
		ErrMessage: err.Error(),
		Data:       nil,
	}
}

// Unauthorized 未登录
func Unauthorized() any {
	return results.ResponseResult{
		Message:    "",
		Code:       int32(enums.CodeUnauthorized),
		ErrMessage: enums.CodeUnauthorized.Message(),
		Data:       nil,
	}
}

// BizErrStr 业务错误（字符串消息）
func BizErrStr(msg string) any {
	return results.ResponseResult{
		Message:    msg,
		Code:       int32(enums.CodeBizErr),
		ErrMessage: msg,
		Data:       nil,
	}
}

func Forbid(err error) any {
	return results.ResponseResult{
		Message:    err.Error(),
		Code:       int32(enums.CodeForbidden),
		ErrMessage: enums.CodeForbidden.Message(),
		Data:       nil,
	}
}

func Internal(err error, data any) any {
	return results.ResponseResult{
		Message:    err.Error(),
		Code:       int32(enums.CodeInternalErr),
		ErrMessage: enums.CodeInternalErr.Message(),
		Data:       data,
	}
}

func Success(data any) any {
	result := results.ResponseResult{
		Message:    "",
		Code:       int32(enums.CodeSuccess),
		ErrMessage: enums.CodeSuccess.Message(),
	}

	if fieldsProvider, ok := data.(contracts.FieldsProvider); ok {
		result.Data = fieldsProvider.ToFields()
	} else {
		result.Data = data
	}

	return result
}
