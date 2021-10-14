package errcode

import "fmt"

type Error struct {
	// 业务错误码
	Code	int
	// 错误信息
	Msg		string
	// 错误细节
	Detail	string
}

func NewError(code int, message string) *Error {
	return &Error{Code: code, Msg: message}
}

func (e *Error) WithDetail(detail string) *Error {
	err := e.Clone()
	err.Detail = detail
	return err
}

func (e *Error) Clone() *Error {
	err := *e
	return &err
}

func (e *Error) Error() string {
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Msg)
}