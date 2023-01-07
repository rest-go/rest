package database

import (
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type URLQuery struct {
	driver string
	values url.Values
}

func NewURLQuery(driver string, values url.Values) *URLQuery {
	return &URLQuery{driver, values}
}

// SelectQuery return sql projection string
func (q *URLQuery) SelectQuery() string {
	selects := q.values["select"]
	if len(selects) == 0 {
		return "*"
	}

	columns := strings.Split(selects[0], ",")
	for i, c := range columns {
		columns[i] = buildColumn(c, true)
	}
	return strings.Join(columns, ",")
}

// OrderQuery returns sql order query string
func (q *URLQuery) OrderQuery() string {
	orders := q.values["order"]
	if len(orders) == 0 {
		return ""
	}
	return strings.ReplaceAll(orders[0], ".", " ")
}

// WhereQuery returns sql and args for where clause
func (q *URLQuery) WhereQuery(index uint) (newIndex uint, query string, args []any) {
	if len(q.values) == 0 {
		return index, "", nil
	}

	var queryBuilder strings.Builder
	args = make([]any, 0, len(q.values))
	first := true
	for k, v := range q.values {
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

		column := buildColumn(k, false)
		queryBuilder.WriteString(column)
		if op == "in" {
			queryBuilder.WriteString(" IN ")
			queryBuilder.WriteString(val)
		} else {
			queryBuilder.WriteString(operator)
			queryBuilder.WriteString(placeholder(q.driver, index))
			args = append(args, val)
			index++
		}
		first = false
	}

	return index, queryBuilder.String(), args
}

func (q *URLQuery) Page() (page, pageSize int) {
	page = 1
	pageSize = 100
	if p, ok := q.values["page"]; ok {
		page, _ = strconv.Atoi(p[0])
	}
	if p, ok := q.values["page_size"]; ok {
		pageSize, _ = strconv.Atoi(p[0])
	}
	return page, pageSize
}

func (q *URLQuery) IsDebug() bool {
	_, ok := q.values["debug"]
	return ok
}

func (q *URLQuery) IsCount() bool {
	_, ok := q.values["count"]
	return ok
}

func (q *URLQuery) IsSingular() bool {
	_, ok := q.values["singular"]
	return ok
}

func buildColumn(c string, as bool) string {
	columnName := c
	asName := ""

	// a->'b'->>'c' AS c
	var JSONOP = regexp.MustCompile("->>?")
	if JSONOP.MatchString(c) {
		splits := JSONOP.Split(c, -1)
		asName = strings.Trim(splits[len(splits)-1], "'")
	}
	if as && asName != "" {
		columnName += fmt.Sprintf(" AS %s", asName)
	}

	return columnName
}
