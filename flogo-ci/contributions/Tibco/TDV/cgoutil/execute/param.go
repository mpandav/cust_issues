package execute

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/alexbrainman/odbc/api"
	"github.com/project-flogo/core/support/log"
	"github.com/tibco/flogo-tdv/src/app/TDV/cgoutil/coredbutils"
)

type Parameter struct {
	SQLType     api.SQLSMALLINT
	Decimal     api.SQLSMALLINT
	Size        api.SQLULEN
	isDescribed bool
	// Following fields store data used later by SQLExecute.
	// The fields keep data alive and away from gc.
	Data             interface{}
	StrLen_or_IndPtr api.SQLLEN
}
type InputParamMetadata struct {
	ParamName string
	ParamType string
}

// StoreStrLen_or_IndPtr stores v into StrLen_or_IndPtr field of p
// and returns address of that field.
func (p *Parameter) StoreStrLen_or_IndPtr(v api.SQLLEN) *api.SQLLEN {
	p.StrLen_or_IndPtr = v
	return &p.StrLen_or_IndPtr

}

func (p *Parameter) BindValue(h api.SQLHSTMT, idx int, v interface{}) error {
	// TODO: Reuse memory for previously bound values. If memory
	// is reused, we, probably, do not need to call SQLBindParameter either.
	var ctype, sqltype, decimal api.SQLSMALLINT
	var size api.SQLULEN
	var buflen api.SQLLEN
	var plen *api.SQLLEN
	var buf unsafe.Pointer
	switch d := v.(type) {
	case nil:
		ctype = api.SQL_C_WCHAR
		p.Data = nil
		buf = nil
		size = 1
		buflen = 0
		plen = p.StoreStrLen_or_IndPtr(api.SQL_NULL_DATA)
		sqltype = api.SQL_WCHAR
	case string:
		ctype = api.SQL_C_CHAR
		b := []byte(d)
		p.Data = b
		buf = unsafe.Pointer(&b[0])
		l := len(b)
		l -= 1 // remove terminating 0
		size = api.SQLULEN(l)
		if size < 1 {
			// size cannot be less then 1 even for empty fields
			size = 1
		}
		//l *= 2 // every char takes 2 bytes
		buflen = api.SQLLEN(l)
		plen = p.StoreStrLen_or_IndPtr(buflen)
		switch {
		case size >= 4000:
			sqltype = api.SQL_WLONGVARCHAR
		case p.isDescribed:
			sqltype = p.SQLType
		case size <= 1:
			sqltype = api.SQL_WVARCHAR
		default:
			sqltype = api.SQL_CHAR
		}
	case int64:
		if -0x80000000 < d && d < 0x7fffffff {
			// Some ODBC drivers do not support SQL_BIGINT.
			// Use SQL_INTEGER if the value fit in int32.
			// See issue #78 for details.
			d2 := int32(d)
			ctype = api.SQL_C_LONG
			p.Data = &d2
			buf = unsafe.Pointer(&d2)
			sqltype = api.SQL_INTEGER
			size = 4
		} else {
			ctype = api.SQL_C_SBIGINT
			p.Data = &d
			buf = unsafe.Pointer(&d)
			sqltype = api.SQL_BIGINT
			size = 8
		}
	case bool:
		var boolval string
		if d {
			boolval = "true"
		} else {
			boolval = "false"
		}
		b := []byte(boolval)
		ctype = api.SQL_C_CHAR
		p.Data = b
		buf = unsafe.Pointer(&b[0])
		l := len(b)
		l -= 1 // remove terminating 0
		size = api.SQLULEN(l)
		if size < 1 {
			// size cannot be less then 1 even for empty fields
			size = 1
		}
		//l *= 2 // every char takes 2 bytes
		buflen = api.SQLLEN(l)
		sqltype = api.SQL_CHAR
	case float64:
		ctype = api.SQL_C_DOUBLE
		p.Data = &d
		buf = unsafe.Pointer(&d)
		sqltype = api.SQL_DOUBLE
		size = 8
	case time.Time:
		ctype = api.SQL_C_TYPE_TIMESTAMP
		y, m, day := d.Date()
		b := api.SQL_TIMESTAMP_STRUCT{
			Year:     api.SQLSMALLINT(y),
			Month:    api.SQLUSMALLINT(m),
			Day:      api.SQLUSMALLINT(day),
			Hour:     api.SQLUSMALLINT(d.Hour()),
			Minute:   api.SQLUSMALLINT(d.Minute()),
			Second:   api.SQLUSMALLINT(d.Second()),
			Fraction: api.SQLUINTEGER(d.Nanosecond()),
		}
		p.Data = &b
		buf = unsafe.Pointer(&b)
		sqltype = api.SQL_TYPE_TIMESTAMP
		if p.isDescribed && p.SQLType == api.SQL_TYPE_TIMESTAMP {
			decimal = p.Decimal
		}
		if decimal <= 0 {
			// represented as yyyy-mm-dd hh:mm:ss.fff format in ms sql server
			decimal = 3
		}
		size = 20 + api.SQLULEN(decimal)
	case []byte:
		ctype = api.SQL_C_BINARY
		b := make([]byte, len(d))
		copy(b, d)
		p.Data = b
		buf = unsafe.Pointer(&b[0])
		buflen = api.SQLLEN(len(b))
		plen = p.StoreStrLen_or_IndPtr(buflen)
		size = api.SQLULEN(len(b))
		switch {
		case p.isDescribed:
			sqltype = p.SQLType
		case size <= 0:
			sqltype = api.SQL_LONGVARBINARY
		case size >= 8000:
			sqltype = api.SQL_LONGVARBINARY
		default:
			sqltype = api.SQL_BINARY
		}
	default:
		panic(fmt.Errorf("unsupported type %T", v))
	}
	ret := api.SQLBindParameter(h, api.SQLUSMALLINT(idx),
		api.SQL_PARAM_INPUT, ctype, sqltype, size, decimal,
		api.SQLPOINTER(buf), buflen, plen)
	if coredbutils.IsError(ret) {
		return coredbutils.NewError("SQLBindParameter", h)
	}
	return nil
}

func ExtractInputParameters(h api.SQLHSTMT, inputParamPositions []int, logCache log.Logger) ([]Parameter, error) {
	// count parameters
	var n, nullable api.SQLSMALLINT
	inputparamlength := len(inputParamPositions)
	n = api.SQLSMALLINT(inputparamlength)
	if n <= 0 {
		// no parameters
		return nil, nil
	}
	ps := make([]Parameter, n)
	// fetch param descriptions
	for i := range ps {
		p := &ps[i]
		ret := api.SQLDescribeParam(h, api.SQLUSMALLINT(inputParamPositions[i]),
			&p.SQLType, &p.Size, &p.Decimal, &nullable)
		if coredbutils.IsError(ret) {
			logCache.Debug("Error while Describing Input Parameter ")
			continue
		}
		p.isDescribed = true
		// SQL Server MAX types (varchar(max), nvarchar(max),
		// varbinary(max) are identified by size = 0
		if p.Size == 0 {
			switch p.SQLType {
			case api.SQL_VARBINARY:
				p.SQLType = api.SQL_LONGVARBINARY
			case api.SQL_VARCHAR:
				p.SQLType = api.SQL_LONGVARCHAR
			case api.SQL_WVARCHAR:
				p.SQLType = api.SQL_WLONGVARCHAR
			}
		}
	}
	return ps, nil
}
