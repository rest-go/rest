package main

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var tableNameReg = regexp.MustCompile("^[_a-zA-Z][_a-zA-Z0-9]*$")

func isValidTableName(tableName string) bool {
	log.Printf("tablename: %s", tableName)
	return tableNameReg.MatchString(tableName)
}

func extractPage(values url.Values) (int, int) {
	page := 1
	pageSize := 100
	if p, ok := values["page"]; ok {
		page, _ = strconv.Atoi(p[0])
		delete(values, "page")
	}
	if p, ok := values["page_size"]; ok {
		pageSize, _ = strconv.Atoi(p[0])
		delete(values, "page_size")
	}
	log.Print("extract page: ", page, pageSize)
	return page, pageSize
}

// buildSelects...
func buildSelects(values url.Values) string {
	selects := values["select"]
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

// buildOrderQuery...
func buildOrderQuery(values url.Values) string {
	orders := values["order"]
	if len(orders) == 0 {
		return ""
	}
	return strings.ReplaceAll(orders[0], ".", " ")
}

// buildWhereQuery build sql where clause from url values
func buildWhereQuery(index int, query url.Values) (int, string, []any) {
	if len(query) == 0 {
		return index, "", nil
	}

	var queryBuilder strings.Builder
	args := make([]any, 0, len(query))
	first := true
	for k, v := range query {
		if _, ok := ReservedKeys[k]; ok {
			continue
		}
		vals := strings.Split(v[0], ".")
		if len(vals) != 2 {
			log.Print("unsupported vals: ", vals)
			continue
		}
		op, val := vals[0], vals[1]
		operation, ok := Operations[op]
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
			queryBuilder.WriteString(operation)
			queryBuilder.WriteString(fmt.Sprintf("$%d", index))
			args = append(args, val)
			index++
		}
		first = false
	}

	return index, queryBuilder.String(), args
}

func buildSetQuery(index int, data map[string]any) (int, string, []any) {
	if len(data) == 0 {
		return index, "", nil
	}

	var queryBuilder strings.Builder
	args := make([]any, 0, len(data))
	for k, v := range data {
		queryBuilder.WriteString(k)
		queryBuilder.WriteString(" = ")
		queryBuilder.WriteString(fmt.Sprintf("$%d", index))
		args = append(args, v)
		index++
	}

	return index, queryBuilder.String(), args
}

func identKeys(m map[string]any, keys []string) bool {
	if len(m) != len(keys) {
		return false
	}

	for _, k := range keys {
		if _, ok := m[k]; !ok {
			return false
		}
	}
	return true
}

func isDebug(v url.Values) bool {
	_, ok := v["debug"]
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
