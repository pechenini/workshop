package todo

import (
	"fmt"
)

const (
	ErrValidation   = iota + 1 //starting from 1 assign to consts below val++
	ErrInternal                // 2
	ErrEventPublish            // 3
	ErrNotFound                // 4
)

type Error struct {
	Code int
	Msg  string
	Err  error
}

func newError(code int, msg string, err error) error {
	return &Error{
		Code: code,
		Msg:  msg,
		Err:  err,
	}
}

func (err *Error) Unwrap() error {
	return err.Err
}

func (err *Error) Error() string {
	if err.Err != nil {
		return fmt.Sprintf("%s: %s", err.Msg, err.Err.Error())
	}
	return err.Msg
}
