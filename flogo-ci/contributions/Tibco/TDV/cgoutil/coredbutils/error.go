package coredbutils

import (
	"errors"
	"fmt"
	"strings"
	"unsafe"

	"github.com/alexbrainman/odbc/api"
)

func IsError(ret api.SQLRETURN) bool {
	return !(ret == api.SQL_SUCCESS || ret == api.SQL_SUCCESS_WITH_INFO)
}

type DiagRecord struct {
	State       string
	NativeError int
	Message     string
}

func (r *DiagRecord) String() string {
	return fmt.Sprintf("{%s} %s", r.State, r.Message)
}

type Error struct {
	APIName string
	Diag    []DiagRecord
}

func (e *Error) Error() string {
	ss := make([]string, len(e.Diag))
	for i, r := range e.Diag {
		ss[i] = r.String()
	}
	return e.APIName + ": " + strings.Join(ss, "\n")
}

func NewError(apiName string, handle interface{}) error {
	h, ht, herr := ToHandleAndType(handle)
	if herr != nil {
		return herr
	}
	err := &Error{APIName: apiName}
	var ne api.SQLINTEGER
	state := make([]byte, 6)
	msg := make([]byte, 1024)
	for i := 1; ; i++ {
		var l api.SQLSMALLINT
		ret := api.SQLGetDiagRec(ht, h, api.SQLSMALLINT(i),
			(*api.SQLCHAR)(unsafe.Pointer(&state[0])), &ne,
			(*api.SQLCHAR)(unsafe.Pointer(&msg[0])),
			api.SQLSMALLINT(len(msg)), &l)
		if ret == api.SQL_NO_DATA {
			logCache.Debug("error retrival no data")
			break
		}
		if IsError(ret) {
			return fmt.Errorf("SQLGetDiagRec failed: ret=%d", ret)
		}
		r := DiagRecord{
			State:       string(state),
			NativeError: int(ne),
			Message:     string(msg[:int(l)]),
		}
		// this is commented but this is tricky in case of connection interruptions
		// LOOK OUT for those kinds of behaviours
		if (apiName == "SQLPrepare" || apiName == "SQLNumResultCols") && r.State == "08S01" {
			return errors.New("08S01")
		}
		err.Diag = append(err.Diag, r)
	}
	return err
}

func IsSuccess(ret api.SQLRETURN) bool {
	return (uint32(ret) & (^uint32(1))) == 0
}
