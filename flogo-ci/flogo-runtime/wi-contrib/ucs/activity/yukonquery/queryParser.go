package yukonquery

import (
	"fmt"
	"strings"

	"github.com/project-flogo/core/data/coerce"
)

const (
	SELECT     = "select"
	ALL        = "*"
	TOP        = "top"
	SKIP       = "skip"
	FROM       = "from"
	WHERE      = "where"
	ORDERBY    = "orderby"
	ASCENDING  = "asc"
	DESCENDING = "desc"
)

var OpMap = map[string]string{
	"=":  "eq",
	"==": "eq",
	"!=": "ne",
	"<>": "ne",
	">":  "gt",
	">=": "ge",
	"!<": "ge",
	"<":  "lt",
	"<=": "le",
	"!>": "le",
}

const (
	AND = "and"
	OR  = "or"
)

type Query struct {
	Select  string
	Top     string
	Skip    string
	From    string
	Where   string
	Orderby string
}

func parseQuery(queryString string, params map[string]string) (*Query, error) {

	queryString = strings.ReplaceAll(queryString, ",", " ")
	queryString = strings.TrimSpace(queryString)
	if queryString == "" {
		return nil, fmt.Errorf("'query' is required")
	}

	tmp := strings.Split(queryString, " ")
	var queryParts []string
	for _, queryPart := range tmp {
		if strings.TrimSpace(queryPart) != "" {
			queryParts = append(queryParts, queryPart)
		}
	}

	selectIndex := -1
	topIndex := -1
	skipIndex := -1
	fromIndex := -1
	whereIndex := -1
	orderbyIndex := -1

	for i, queryPart := range queryParts {
		switch strings.ToLower(queryPart) {
		case SELECT:
			selectIndex = i
		case TOP:
			topIndex = i
		case SKIP:
			skipIndex = i
		case FROM:
			fromIndex = i
		case WHERE:
			whereIndex = i
		case ORDERBY:
			orderbyIndex = i
		}
	}

	if selectIndex != 0 {
		return nil, fmt.Errorf("invalid query: only select statements are supported")
	}

	if fromIndex == -1 {
		return nil, fmt.Errorf("invalid query: a from clause is required")
	}

	if fromIndex+1 >= len(queryParts) {
		return nil, fmt.Errorf("invalid query: table name is required")
	}

	var queryObj = Query{}

	if topIndex != -1 {
		topValueIndex := topIndex + 1
		if topValueIndex >= len(queryParts) {
			return nil, fmt.Errorf("invalid query: value not found for top")
		}
		topValue := queryParts[topValueIndex]
		intTopValue, err := coerce.ToInt(topValue)
		if err != nil {
			return nil, fmt.Errorf("invalid query: invalid value top '%s'", topValue)
		}
		topValue, _ = coerce.ToString(intTopValue)
		queryObj.Top = topValue

		if topIndex < fromIndex {
			if topValueIndex > selectIndex {
				selectIndex = topValueIndex
			}
		}
	}

	if skipIndex != -1 {
		skipValueIndex := skipIndex + 1
		if skipValueIndex >= len(queryParts) {
			return nil, fmt.Errorf("invalid query: value not found for skip")
		}
		skipValue := queryParts[skipValueIndex]
		intSkipValue, err := coerce.ToInt(skipValue)
		if err != nil {
			return nil, fmt.Errorf("invalid query: invalid value skip '%s'", skipValue)
		}
		skipValue, _ = coerce.ToString(intSkipValue)
		queryObj.Skip = skipValue

		if skipIndex < fromIndex {
			if skipValueIndex > selectIndex {
				selectIndex = skipValueIndex
			}
		}
	}

	columnNames, err := getColumnNames(queryParts[selectIndex+1:])
	if err != nil {
		return nil, err
	}
	queryObj.Select = columnNames

	tableName, err := getTableName(queryParts[fromIndex+1:])
	if err != nil {
		return nil, err
	}
	queryObj.From = tableName

	if whereIndex != -1 {
		where, err := getWhere(queryParts[whereIndex+1:])
		if err != nil {
			return nil, err
		}

		for param, value := range params {
			strParam := ":" + param
			strValue, _ := coerce.ToString(value)
			originalWhere := where
			where = strings.ReplaceAll(where, strParam, strValue)
			if where == originalWhere {
				return nil, fmt.Errorf("invalid query: input param '%s' not found in query", param)
			}
		}

		queryObj.Where = where
	}

	if orderbyIndex != -1 {
		orderbyValueIndex := orderbyIndex + 1
		if orderbyValueIndex >= len(queryParts) {
			return nil, fmt.Errorf("invalid query: value not found for orderby")
		}
		orderbyValue := queryParts[orderbyValueIndex]

		orderbyValueIndex += 1
		if orderbyValueIndex < len(queryParts) {
			ascdesc := strings.ToLower(queryParts[orderbyValueIndex])
			if ascdesc == ASCENDING || ascdesc == DESCENDING {
				orderbyValue = fmt.Sprintf("%s %s", orderbyValue, ascdesc)
			}
		}

		queryObj.Orderby = orderbyValue
	}

	return &queryObj, nil
}

func isDelimiter(queryPart string) bool {

	lowerQueryPart := strings.ToLower(queryPart)

	return lowerQueryPart == SELECT ||
		lowerQueryPart == TOP ||
		lowerQueryPart == SKIP ||
		lowerQueryPart == FROM ||
		lowerQueryPart == WHERE ||
		lowerQueryPart == ORDERBY
}

func getColumnNames(subParts []string) (string, error) {

	columnNames := ""
	for _, queryPart := range subParts {
		if queryPart == ALL {
			columnNames = ALL
			break
		} else if isDelimiter(queryPart) {
			break
		} else {
			if columnNames != "" {
				columnNames += ", "
			}
			columnNames += queryPart
		}
	}
	if columnNames == "" {
		return "", fmt.Errorf("invalid query: select requires column list or * for all")
	}
	return columnNames, nil
}

func getTableName(subParts []string) (string, error) {

	tableName := ""
	for _, queryPart := range subParts {
		if isDelimiter(queryPart) {
			break
		} else {
			tableName = queryPart
		}
	}
	if tableName == "" {
		return "", fmt.Errorf("invalid query: table name not found")
	}
	return tableName, nil
}

func getWhere(subParts []string) (string, error) {

	where := ""

	left := ""
	op := ""
	right := ""
	logicOp := ""

	for _, queryPart := range subParts {

		if isDelimiter(queryPart) {
			break
		} else {
			if left == "" {
				left = queryPart
			} else if op == "" {
				op = queryPart
			} else if right == "" {
				right = queryPart
			} else {
				logicOp = queryPart

				wherePart, err := buildWherePart(left, op, right, logicOp)
				if err != nil {
					return "", err
				}

				where += wherePart

				left = ""
				op = ""
				right = ""
				logicOp = ""
			}
		}
	}

	if left != "" {
		wherePart, err := buildWherePart(left, op, right, logicOp)
		if err != nil {
			return "", err
		}

		where += wherePart
	}

	where = strings.TrimSpace(where)

	if where == "" {
		return "", fmt.Errorf("invalid query: empty where clause")
	}
	return where, nil
}

func buildWherePart(left string, op string, right string, logicOp string) (string, error) {

	if left == "" || op == "" || right == "" {
		return "", fmt.Errorf("invalid query: invalid where clause '%s %s %s'", left, op, right)
	}

	opStr, ok := OpMap[strings.ToLower(op)]
	if ok == false {
		return "", fmt.Errorf("invalid query: unknown operator '%s %s %s'", left, op, right)
	}

	lowerLogicOp := strings.ToLower(logicOp)
	if lowerLogicOp != "" && lowerLogicOp != AND && lowerLogicOp != OR {
		return "", fmt.Errorf("invalid query: unknown logical operator '%s'", logicOp)
	}

	return fmt.Sprintf("%s %s %s %s ", left, opStr, right, lowerLogicOp), nil
}
