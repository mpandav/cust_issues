// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package odbc

import (
	"database/sql/driver"
	"io"

	"github.com/alexbrainman/odbc/api"
)

type Rows struct {
	os *ODBCStmt
}

func (r *Rows) Columns() []string {
	names := make([]string, len(r.os.Cols))
	for i := 0; i < len(names); i++ {
		names[i] = r.os.Cols[i].Name()
	}
	return names
}

func (r *Rows) Next(dest []driver.Value) error {
	ret := api.SQLFetch(r.os.h)
	if ret == api.SQL_NO_DATA {
		return io.EOF
	}
	if IsError(ret) {
		return NewError("SQLFetch", r.os.h)
	}
	for i := range dest {
		v, err := r.os.Cols[i].Value(r.os.h, i)
		if err != nil {
			return err
		}
		dest[i] = v
	}
	return nil
}

func (r *Rows) Close() error {
	return r.os.closeByRows()
}

func (r *Rows) ColumnTypeDatabaseTypeName(index int) string {
	baseColumn, ok := r.os.Cols[index].(*BindableColumn)
	if !ok {
		baseColumn := r.os.Cols[index].(*NonBindableColumn)
		return getSQLType(baseColumn.CType)
	}
	return getSQLType(baseColumn.CType)
}
func getSQLType(sqltype api.SQLSMALLINT) string {
	switch sqltype {
	case api.SQL_BIT:
		return "BOOLEAN"
	case api.SQL_TINYINT:
		return "TINYINT"
	case api.SQL_SMALLINT:
		return "SMALLINT"
	case api.SQL_INTEGER:
		return "INTEGER"
	case api.SQL_BIGINT, api.SQL_C_SBIGINT, api.SQL_C_UBIGINT:
		return "BIGINT"
	case api.SQL_NUMERIC:
		return "NUMERIC"
	case api.SQL_DECIMAL:
		return "DECIMAL"
	case api.SQL_FLOAT:
		return "FLOAT"
	case api.SQL_REAL:
		return "REAL"
	case api.SQL_DOUBLE:
		return "DOUBLE"
	case api.SQL_DATE, api.SQL_TYPE_DATE:
		return "DATE"
	case api.SQL_TIME, api.SQL_TYPE_TIME:
		return "TIME"
	case api.SQL_TIMESTAMP, api.SQL_TYPE_TIMESTAMP:
		return "TIMESTAMP"
	case api.SQL_CHAR:
		return "CHAR"
	case api.SQL_VARCHAR:
		return "VARCHAR"
	case api.SQL_WCHAR:
		return "CHAR"
	case api.SQL_BINARY:
		return "BINARY"
	case api.SQL_VARBINARY:
		return "BYTEA"
	case api.SQL_LONGVARCHAR:
		return "LONGVARCHAR"
	case api.SQL_WLONGVARCHAR:
		return "XML"
	case api.SQL_INTERVAL_DAY:
		return "INTERVAL_DAY"
	case api.SQL_INTERVAL_MONTH:
		return "INTERVAL_MONTH"
	case api.SQL_INTERVAL_YEAR:
		return "INTERVAL_YEAR"
	default:
		return "VARCHAR"
	}
}
