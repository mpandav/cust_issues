package connection

import (
	"fmt"
	"sort"
	"strings"
	"unsafe"

	"github.com/alexbrainman/odbc"
	"github.com/alexbrainman/odbc/api"
	"github.com/tibco/flogo-tdv/src/app/TDV/cgoutil/coredbutils"
)

func getDriverName() (string, error) {
	drivers, err := getDrivers()
	if err != nil {
		return "", err
	}
	logCache.Debug("List of Drivers Installed on System : ", drivers)
	if drivers == nil || len(drivers) == 0 {
		return "", fmt.Errorf("no drivers installed, install odbc driver for TDV")
	}
	sort.Strings(drivers)
	// TODO : we may need to update the list we are searching drivername in
	drvName, err := getDriverNameFromList("TDV", drivers, "TIBCO(R) Data Virtualization", "TIBCO(R) Data Virtualization 8.5", "Data Virtualization", "composite")
	logCache.Debug("Driver Name found is ", drvName)
	return drvName, nil
}
func getDriverNameFromList(dbms string, drivers []string, patterns ...string) (string, error) {
	for _, driver := range drivers {
		for _, pattern := range patterns {
			if strings.Contains(strings.ToLower(driver), strings.ToLower(pattern)) {
				return "{" + driver + "}", nil
			}
		}
	}

	return "", fmt.Errorf("driver name not found, check whether any %s driver is installed", dbms)
}

// returns array of drivers present in the system
func getDrivers() ([]string, error) {
	var drivers []string
	var envHandle api.SQLHENV
	var envHandleBase api.SQLHANDLE
	var sqlReturn api.SQLRETURN
	// Allocating New handle to get drivers on system which uses only DriverManager
	inputHandle := api.SQLHANDLE(api.SQL_NULL_HANDLE)
	sqlReturn = api.SQLAllocHandle(api.SQL_HANDLE_ENV, inputHandle, &envHandleBase)
	if odbc.IsError(sqlReturn) {
		err := odbc.NewError("SQLAllocHandle", inputHandle)
		return nil, err
	}

	envHandle = api.SQLHENV(envHandleBase)

	sqlReturn = api.SQLSetEnvUIntPtrAttr(envHandle, api.SQL_ATTR_ODBC_VERSION, api.SQL_OV_ODBC3, 0)
	if odbc.IsError(sqlReturn) {
		defer coredbutils.ReleaseHandle(envHandle)
		err := odbc.NewError("SQLSetEnvUIntPtrAttr", envHandle)
		return nil, err
	}
	buffer := make([]byte, 1024)
	var bufferLength api.SQLSMALLINT
	direction := api.SQLUSMALLINT(api.SQL_FETCH_FIRST)

	flag := true

	for flag {
		sqlReturn = api.SQLDrivers(envHandle, direction, (*api.SQLCHAR)(unsafe.Pointer(&buffer[0])), api.SQLSMALLINT(len(buffer)), &bufferLength, nil, 0, nil)
		if odbc.IsError(sqlReturn) {
			if sqlReturn == api.SQL_NO_DATA && len(drivers) != 0 {
				break
			}
			err := odbc.NewError("SQLDrivers", envHandle)
			return nil, fmt.Errorf("Fetching drivers from system failed , %s", err.Error())
		}
		direction = api.SQL_FETCH_NEXT
		bLength := int(bufferLength)
		drivers = append(drivers, string(buffer[:bLength]))
	}
	return drivers, nil
}
