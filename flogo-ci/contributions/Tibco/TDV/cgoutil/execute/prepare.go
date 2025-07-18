package execute

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/alexbrainman/odbc/api"
	"github.com/tibco/flogo-tdv/src/app/TDV/cgoutil/coredbutils"
)

// GetStatementHandle Will receive Connection Handle and return Statement Handle allocated
func GetStatementHandle(ConnHandle api.SQLHDBC) (stmt api.SQLHSTMT, err error) {
	// var statementHandle api.SQLHANDLE
	// ret := api.SQLAllocHandle(api.SQL_HANDLE_STMT, api.SQLHANDLE(ConnHandle), &statementHandle)
	// if coredbutils.IsError(ret) {
	// 	defer coredbutils.ReleaseHandle(statementHandle)
	// 	return nil, coredbutils.NewError("SQLAllocHandle", statementHandle)
	// }
	// stmt := api.SQLHSTMT(statementHandle)
	// return stmt, nil
	var statementHandle api.SQLHANDLE
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to allocate sql statement handle error : %#v", r)
			stmt = api.SQLHSTMT(statementHandle)
		}
	}()
	// coredbutils.ReleaseHandle(ConnHandle)
	var i int
	wg := &sync.WaitGroup{}
	wg.Add(2)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGABRT, syscall.SIGHUP, syscall.SIGINT, syscall.SIGKILL,
		syscall.SIGQUIT, syscall.SIGTERM)
	exitChan := make(chan int, 1)
	defer func() {
		close(exitChan)
		close(sigChan)
	}()
	go func() {
		select {
		case s := <-sigChan:
			fmt.Printf("\n****** GOT signal %#v\n", s)
			wg.Done()
			return
		case <-exitChan:
			wg.Done()
			return
		}
	}()
	go func(wg *sync.WaitGroup, i *int) {
		ret := api.SQLAllocHandle(api.SQL_HANDLE_STMT, api.SQLHANDLE(ConnHandle), &statementHandle)
		if coredbutils.IsError(ret) {
			defer coredbutils.ReleaseHandle(statementHandle)
			*i = 1
		}
		exitChan <- 1
		wg.Done()
	}(wg, &i)
	wg.Wait()
	if i == 1 {
		return stmt, coredbutils.NewError("SQLAllocHandle", statementHandle)
	}
	stmt = api.SQLHSTMT(statementHandle)
	return stmt, nil
}

// PrepareStatment Prepares Statement Based on the Queries
func PrepareStatment(stmt api.SQLHSTMT, query string) (api.SQLHSTMT, error) {
	queryText, err := syscall.ByteSliceFromString(query)
	if err != nil {
		return stmt, fmt.Errorf("could Not convert query in byte slice format : %v", err)
	}
	StatementLength := api.SQL_NTS
	ret := api.SQLPrepare(stmt, (*api.SQLCHAR)(&queryText[0]), api.SQLINTEGER(StatementLength))
	if coredbutils.IsError(ret) {
		defer coredbutils.ReleaseHandle(stmt)
		return stmt, coredbutils.NewError("SQLPrepare", stmt)
	}
	return stmt, nil
}
