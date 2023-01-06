package database

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
)

type URLQuery url.Values

// SelectQuery return sql projection string
func (q URLQuery) SelectQuery() string {
	selects := q["select"]
	if len(selects) == 0 {
		return "*"
	}

	columns := strings.Split(selects[0], ",")
	for i, c := range columns {
		column, err := buildColumn(c, true)
		if err != nil {
			log.Print("invalid column: ", c)
			continue
		}
		columns[i] = column
	}
	return strings.Join(columns, ",")
}

// OrderQuery returns sql order query string
func (q URLQuery) OrderQuery() string {
	orders := q["order"]
	if len(orders) == 0 {
		return ""
	}
	return strings.ReplaceAll(orders[0], ".", " ")
}

// WhereQuery returns sql and args for where clause
func (q URLQuery) WhereQuery(index int) (newIndex int, query string, args []any) {
	if len(q) == 0 {
		return index, "", nil
	}

	var queryBuilder strings.Builder
	args = make([]any, 0, len(q))
	first := true
	for k, v := range q {
		if _, ok := ReservedWords[k]; ok {
			continue
		}
		vals := strings.Split(v[0], ".")
		if len(vals) != 2 {
			log.Print("unsupported vals: ", vals)
			continue
		}
		op, val := vals[0], vals[1]
		operator, ok := Operators[op]
		if !ok {
			log.Print("unsupported op: ", op)
			continue
		}

		if !first {
			queryBuilder.WriteString(" AND ")
		}

		column, err := buildColumn(k, false)
		if err != nil {
			log.Print("invalid field: ", k)
			continue
		}
		queryBuilder.WriteString(column)
		if op == "in" {
			queryBuilder.WriteString(" in ")
			queryBuilder.WriteString(val)
		} else {
			queryBuilder.WriteString(operator)
			queryBuilder.WriteString(fmt.Sprintf("$%d", index))
			args = append(args, val)
			index++
		}
		first = false
	}

	return index, queryBuilder.String(), args
}

func (q URLQuery) Page() (page, pageSize int) {
	page = 1
	pageSize = 100
	if p, ok := q["page"]; ok {
		page, _ = strconv.Atoi(p[0])
	}
	if p, ok := q["page_size"]; ok {
		pageSize, _ = strconv.Atoi(p[0])
	}
	return page, pageSize
}

func (q URLQuery) IsDebug() bool {
	_, ok := q["debug"]
	return ok
}

func (q URLQuery) IsCount() bool {
	_, ok := q["count"]
	return ok
}

func (q URLQuery) IsSingular() bool {
	_, ok := q["singular"]
	return ok
}

func buildColumn(c string, as bool) (string, error) {
	isJSON := false
	columnName := c
	asName := ""
	if strings.Contains(c, "->") {
		parts := strings.SplitN(c, "->", 2)
		if len(parts) != 2 {
			return "", errors.New("invalid json operation")
		}
		columnName = fmt.Sprintf("%s->'%s'", parts[0], parts[1])
		isJSON = true
		asName = parts[1]
	}
	if strings.Contains(c, "->>") {
		parts := strings.SplitN(c, "->>", 2)
		if len(parts) != 2 {
			return "", errors.New("invalid json operation")
		}
		columnName = fmt.Sprintf("%s->>'%s'", parts[0], parts[1])

		isJSON = true
		asName = parts[1]
	}
	if isJSON && as {
		columnName += fmt.Sprintf(" AS %s", asName)
	}

	return columnName, nil
}
