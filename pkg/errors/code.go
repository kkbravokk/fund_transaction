package errors

import "errors"

// 错误码说明
// 1-3位：HTTP状态码
// 4-5位：组件
// 6-8位：组件内部错误码
var (
	CodeBadRequest   = NewAPICode(40000000, "Bad Request")
	CodeUnauthorized = NewAPICode(40100000, "Unauthorized")
	CodeForbidden    = NewAPICode(40300000, "Forbidden")
	CodeNotFound     = NewAPICode(40400000, "Not Found")
	CodeUnknownError = NewAPICode(50000000, "Internal Server Error")
)

type APICoder interface {
	Code() int
	Message() string
	Reference() string
	HTTPStatus() int
}

func NewAPICode(code int, message string, reference ...string) APICoder {
	ref := ""
	if len(reference) > 0 {
		ref = reference[0]
	}
	return &apiCode{
		code:    code,
		message: message,
		ref:     ref,
	}
}

type apiCode struct {
	code    int
	message string
	ref     string
}

func (a *apiCode) Code() int {
	return a.code
}

func (a *apiCode) Message() string {
	return a.message
}

func (a *apiCode) Reference() string {
	return a.ref
}

func (a *apiCode) HTTPStatus() int {
	v := a.code
	for v >= 1000 {
		v /= 10
	}
	return v
}

func ParseCoder(err error) APICoder {
	for {
		if e, ok := err.(interface {
			Coder() APICoder
		}); ok {
			return e.Coder()
		}
		if errors.Unwrap(err) == nil {
			return CodeUnknownError
		}
		err = errors.Unwrap(err)
	}
}

func IsCode(err error, coder APICoder) bool {
	if err == nil {
		return false
	}
	for {
		if e, ok := err.(interface {
			Coder() APICoder
		}); ok {
			if e.Coder().Code() == coder.Code() {
				return true
			}
		}
		if errors.Unwrap(err) == nil {
			return false
		}
		err = errors.Unwrap(err)
	}
}
