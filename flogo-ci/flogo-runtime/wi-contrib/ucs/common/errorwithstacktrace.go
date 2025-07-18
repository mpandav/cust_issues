package common

import (
	"fmt"
	"runtime"
	"strings"
)

type ErrorWithStackTrace struct {
	Err      error
	ErrStack string
}

func NewErrorWithStack(err error) error {
	if err == nil {
		return nil
	}
	errStk := GetStack()
	return &ErrorWithStackTrace{
		Err:      err,
		ErrStack: errStk,
	}
}

func (e *ErrorWithStackTrace) Error() string {
	return fmt.Sprintf("%s\n%s", e.Err.Error(), e.ErrStack)
}

func GetStack() string {
	b := make([]byte, 1000)
	n := runtime.Stack(b, false)
	b1 := make([]byte, n)
	copy(b1, b)
	s := string(b1)
	x := strings.Split(s, "\n")
	//remove 2 stack frames
	y := x[5:]
	z := strings.Join(y, "\n")
	return z
}
