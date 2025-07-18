package execute

import (
	"fmt"
	"unsafe"

	"github.com/alexbrainman/odbc/api"
	"github.com/project-flogo/core/support/log"
	"github.com/tibco/flogo-tdv/src/app/TDV/cgoutil/coredbutils"
)

type ResultColumn struct {
	ColumnName string
	DataBuffer []byte
	CType      api.SQLSMALLINT
}

// ExecuteStatement Executes Given Prepared Statement
func ExecuteStatement(stmt api.SQLHSTMT) (api.SQLHSTMT, error) {
	ret := api.SQLExecute(stmt)
	if coredbutils.IsError(ret) {
		//TODO : should we Always Release Handle for failure??
		defer coredbutils.ReleaseHandle(stmt)
		return stmt, coredbutils.NewError("SQLExecute", stmt)
	}
	return stmt, nil
}

// trimBytes will Trim Byte Slice till First NULL byte
func trimBytes(a []byte) []byte {
	var b []byte
	for i := 0; i < len(a); i++ {
		if a[i] != 0 {
			b = append(b, a[i])
			a[i] = 0
		} else {
			break
		}
	}

	return b
}

// currently implemented only for api.SQL_COLUMN_TYPE_NAME
func SQLColAttribute(h api.SQLHSTMT, idx int, fieldIdentifier api.SQLUSMALLINT) (string, api.SQLRETURN) {
	buff := make([]byte, 100)
	ret := api.SQLColAttribute(h, api.SQLUSMALLINT(idx+1), fieldIdentifier, (api.SQLPOINTER)(unsafe.Pointer(&buff[0])), api.SQLSMALLINT(len(buff)), nil, nil)
	return string(trimBytes(buff)), ret
}

// describeColumn Describes the Column information
func describeColumn(h api.SQLHSTMT, idx int, namebuf []byte) (namelen int, sqltype api.SQLSMALLINT, ret api.SQLRETURN, size api.SQLULEN) {
	var l, decimal, nullable api.SQLSMALLINT

	ret = api.SQLDescribeCol(h, api.SQLUSMALLINT(idx+1),
		(*api.SQLCHAR)(unsafe.Pointer(&namebuf[0])),
		api.SQLSMALLINT(len(namebuf)), &l,
		&sqltype, &size, &decimal, &nullable)
	return int(l), sqltype, ret, size
}

// FetchAllCursorResults will fetch the Data from all cursors and return it in array of map
func FetchAllCursorResults(stmt api.SQLHSTMT, cursors []string, logCache log.Logger) (map[string]interface{}, error) {

	//
	cursorNumber := 0
	AllCursorResults := make(map[string]interface{})
	if len(cursors) > 0 {
		for {
			var nCols api.SQLSMALLINT
			ret := api.SQLNumResultCols(stmt, &nCols)
			if coredbutils.IsError(ret) {
				defer coredbutils.ReleaseHandle(stmt)
				return nil, coredbutils.NewError("SQLNumResultCols", stmt)
			}
			if int(nCols) < 1 {
				logCache.Info("There are 0 Columns in ", cursors[cursorNumber])
				if !(api.SQLMoreResults(stmt) == api.SQL_SUCCESS) {
					break
				} else {
					continue
				}
			}
			//var ColumnData [MAX_COLS][]api.SQLCHAR
			ColumnData := make([]ResultColumn, nCols)
			for i := 0; i < int(nCols); i++ {

				//TODO : Increase initial buffer size //Decide what could be max bufer size for Column name
				namebuf := make([]byte, 64)

				namelen, sqltype, ret, size := describeColumn(stmt, i, namebuf)
				if namelen > len(namebuf) {
					// try again with bigger buffer
					logCache.Debug("trying again with bigger buffer for column name")
					namebuf = make([]byte, namelen)
					namelen, sqltype, ret, size = describeColumn(stmt, i, namebuf)
				}
				if coredbutils.IsError(ret) {
					defer coredbutils.ReleaseHandle(stmt)
					return nil, coredbutils.NewError("SQLDescribeCol", stmt)
				}
				if namelen > len(namebuf) {
					// still complaining about buffer size
					//TODO : check if we want to fail here
					logCache.Info("Failed to allocate column name buffer")
				}

				ColumnData[i].ColumnName = string(namebuf[:namelen])
				var ColumnDataLen api.SQLLEN
				l := int(size)
				switch sqltype {
				case api.SQL_BIT:
					ColumnData[i].CType = api.SQL_C_BIT
				case api.SQL_TINYINT, api.SQL_SMALLINT, api.SQL_INTEGER:
					ColumnData[i].CType = api.SQL_C_LONG
				case api.SQL_BIGINT:
					ColumnData[i].CType = api.SQL_C_SBIGINT
				case api.SQL_NUMERIC, api.SQL_DECIMAL, api.SQL_FLOAT, api.SQL_REAL, api.SQL_DOUBLE:
					ColumnData[i].CType = api.SQL_C_DOUBLE
				case api.SQL_TYPE_TIMESTAMP:
					ColumnData[i].CType = api.SQL_C_CHAR
				case api.SQL_TYPE_DATE:
					ColumnData[i].CType = api.SQL_C_CHAR
				case api.SQL_GUID:
					ColumnData[i].CType = api.SQL_C_CHAR
				case api.SQL_CHAR, api.SQL_VARCHAR:
					l = VariableWidthColumn(api.SQL_C_CHAR, l)
					ColumnData[i].CType = api.SQL_C_CHAR
				case api.SQL_WCHAR, api.SQL_WVARCHAR:
					l = VariableWidthColumn(api.SQL_C_WCHAR, l)
					ColumnData[i].CType = api.SQL_C_WCHAR
				case api.SQL_BINARY, api.SQL_VARBINARY:
					l = VariableWidthColumn(api.SQL_C_BINARY, l)
					ColumnData[i].CType = api.SQL_C_BINARY
				case api.SQL_LONGVARCHAR:
					l = VariableWidthColumn(api.SQL_C_CHAR, l)
					ColumnData[i].CType = api.SQL_C_CHAR
				case api.SQL_WLONGVARCHAR, api.SQL_SS_XML:
					l = VariableWidthColumn(api.SQL_C_WCHAR, l)
					ColumnData[i].CType = api.SQL_C_WCHAR
				case api.SQL_LONGVARBINARY:
					l = VariableWidthColumn(api.SQL_C_BINARY, l)
					ColumnData[i].CType = api.SQL_C_BINARY
				default:
					return nil, fmt.Errorf("unsupported column type %d", sqltype)
				}
				logCache.Debugf("Maximum size of data for column %s : %s", ColumnData[i].ColumnName, l)
				ColumnData[i].DataBuffer = make([]byte, l)
				api.SQLBindCol(stmt, api.SQLUSMALLINT(i+1), ColumnData[i].CType, api.SQLPOINTER(unsafe.Pointer(&ColumnData[i].DataBuffer[0])), api.SQLLEN(l), &ColumnDataLen)

			}

			TableData := make([]map[string]interface{}, 0)
			for i := 0; ; i++ {
				ret = api.SQLFetch(stmt)

				if coredbutils.IsError(ret) {
					logCache.Debugf("No more data in %s while fetching. error : %s", cursors[cursorNumber], coredbutils.NewError("SQLFetch", stmt))
					break
				}
				var rowData = make(map[string]interface{})
				var p unsafe.Pointer
				for j := 0; j < int(nCols); j++ {
					buf := ColumnData[j].DataBuffer
					if len(buf) > 0 {
						p = unsafe.Pointer(&buf[0])
					}
					switch ColumnData[j].CType {
					case api.SQL_C_BIT:
						rowData[ColumnData[j].ColumnName] = (buf[0] != 0)
					case api.SQL_C_SBIGINT:
						rowData[ColumnData[j].ColumnName] = *((*int64)(p))
					case api.SQL_C_LONG:
						rowData[ColumnData[j].ColumnName] = *((*int32)(p))
					case api.SQL_C_DOUBLE:
						rowData[ColumnData[j].ColumnName] = *((*float64)(p))
					default:
						temp := string(trimBytes(buf))
						rowData[ColumnData[j].ColumnName] = temp
					}
				}
				TableData = append(TableData, rowData)
			}
			AllCursorResults[cursors[cursorNumber]] = TableData
			if !(api.SQLMoreResults(stmt) == api.SQL_SUCCESS) {
				break
			}
			cursorNumber = cursorNumber + 1
		}
	}
	return AllCursorResults, nil
}
func VariableWidthColumn(ctype api.SQLSMALLINT, l int) int {
	switch ctype {
	case api.SQL_C_WCHAR:
		l += 1 // room for null-termination character
		l *= 2 // wchars take 2 bytes each
	case api.SQL_C_CHAR:
		l += 1 // room for null-termination character
	case api.SQL_C_BINARY:
		// nothing to do
	}
	return l
}
