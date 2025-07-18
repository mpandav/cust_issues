package sqldatatypes

import "github.com/alexbrainman/odbc/api"

func getPostgreSQLType(sqltype api.SQLSMALLINT) string {
	switch sqltype {
	case api.SQL_BIT:
		return "BOOLEAN"
	case api.SQL_TINYINT:
		return "TINYINT"
	case api.SQL_SMALLINT:
		return "SMALLINT"
	case api.SQL_INTEGER:
		return "INTEGER"
	case api.SQL_BIGINT:
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
	case api.SQL_TYPE_TIMESTAMP:
		return "TIMESTAMP"
	case api.SQL_TYPE_DATE:
		return "DATE"
	case api.SQL_TYPE_TIME:
		return "TIME"
	case api.SQL_SS_TIME2:
		return "TIME2"
	case api.SQL_GUID:
		return "GUID"
	case api.SQL_CHAR:
		return "CHAR"
	case api.SQL_VARCHAR:
		return "VARCHAR"
	case api.SQL_WCHAR:
		return "CHAR"
	case api.SQL_WVARCHAR:
		return "VARCHAR"
	case api.SQL_BINARY:
		return "BINARY"
	case api.SQL_VARBINARY:
		return "BYTEA"
	case api.SQL_LONGVARCHAR:
		return "TEXT"
	case api.SQL_WLONGVARCHAR:
		return "XML"
	case api.SQL_SS_XML:
		return "SS_XML"
	case api.SQL_LONGVARBINARY:
		return "LONGVARBINARY"
	default:
		return "VARCHAR"
	}
}

// GetSQLType returns the respective database sql types...
func GetSQLType(dbms string, sqltype api.SQLSMALLINT) string {
	switch dbms {
	case "PostgreSQL", "postgres":
		return getPostgreSQLType(sqltype)
	default:
		return GetDBSQLType(sqltype)
	}
}

// GetGOType returns the datatype for input fields...
func GetGOType(sqltype string) string {
	switch sqltype {
	case "BIT":
		return "boolean"
	case "TINYINT":
		return "boolean"
	case "SMALLINT":
		return "integer"
	case "INTEGER":
		return "integer"
	case "BIGINT":
		return "integer"
	case "NUMERIC", "DECIMAL", "FLOAT", "REAL", "DOUBLE":
		return "integer"
	case "TYPE_TIMESTAMP":
		return "string"
	case "TYPE_DATE":
		return "string"
	case "TYPE_TIME":
		return "string"
	case "SS_TIME2":
		return "string"
	case "GUID":
		return "string"
	case "CHAR", "VARCHAR":
		return "string"
	case "WCHAR", "WVARCHAR":
		return "string"
	case "BINARY", "VARBINARY":
		return "string"
	case "LONGVARCHAR":
		return "string"
	case "WLONGVARCHAR", "SS_XML":
		return "string"
	case "LONGVARBINARY":
		return "string"
	default:
		return "string"
	}
}

// GetDBSQLType ...
func GetDBSQLType(sqltype api.SQLSMALLINT) string {
	switch sqltype {
	case api.SQL_BIT:
		return "BIT"
	case api.SQL_TINYINT:
		return "TINYINT"
	case api.SQL_SMALLINT:
		return "SMALLINT"
	case api.SQL_INTEGER:
		return "INTEGER"
	case api.SQL_BIGINT:
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
	case api.SQL_TYPE_TIMESTAMP:
		return "TIMESTAMP"
	case api.SQL_TYPE_DATE:
		return "DATE"
	case api.SQL_TYPE_TIME:
		return "TIME"
	case api.SQL_SS_TIME2:
		return "TIME2"
	case api.SQL_GUID:
		return "GUID"
	case api.SQL_CHAR:
		return "CHAR"
	case api.SQL_VARCHAR:
		return "VARCHAR"
	case api.SQL_WCHAR:
		return "WCHAR"
	case api.SQL_WVARCHAR:
		return "WVARCHAR"
	case api.SQL_BINARY:
		return "BINARY"
	case api.SQL_VARBINARY:
		return "VARBINARY"
	case api.SQL_LONGVARCHAR:
		return "LONGVARCHAR"
	case api.SQL_WLONGVARCHAR:
		return "WLONGVARCHAR"
	case api.SQL_SS_XML:
		return "SS_XML"
	case api.SQL_LONGVARBINARY:
		return "LONGVARBINARY"
	default:
		return "VARCHAR"
	}
}

func GetProcedureParamType(paramType api.SQLSMALLINT) string {
	// switch paramType {
	// case api.SQL_PARAM_TYPE_UNKNOWN:
	// 	return "UNKNOWN"
	// case api.SQL_PARAM_INPUT:
	// 	return "IN"
	// case api.SQL_PARAM_INPUT_OUTPUT:
	// 	return "INOUT"
	// case api.SQL_PARAM_OUTPUT:
	// 	return "OUT"
	// case api.SQL_RETURN_VALUE:
	// 	return "RETURNVALUE"
	// case api.SQL_RESULT_COL:
	// 	return "SQLRESULTCOL"
	// default:
	// 	return "UNKNOWN"
	// }
	// logger.Log("INFO", "paramtype: ", paramType)
	switch paramType {
	case 0:
		return "UNKNOWN"
	case 1:
		return "IN"
	case 2:
		return "INOUT"
	case 3:
		return "OUT"
	case 4:
		return "RETURNVALUE"
	case 5:
		return "SQLRESULTCOL"
	default:
		return "UNKNOWN"
	}
}
