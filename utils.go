package main

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
)

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
	return selects[0]
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
	count := 0
	for k, v := range query {
		vals := strings.Split(v[0], ".")
		if len(vals) != 2 {
			continue
		}
		op, val := vals[0], vals[1]
		operation, ok := Operations[op]
		if !ok {
			log.Print("unsupported op: ", op)
			continue
		}

		if count > 0 {
			queryBuilder.WriteString(" AND ")
		}

		queryBuilder.WriteString(k)
		if op == "in" {
			queryBuilder.WriteString(" in ")
			queryBuilder.WriteString(val)
		} else {
			queryBuilder.WriteString(operation)
			queryBuilder.WriteString(fmt.Sprintf("$%d", index))
			args = append(args, val)
			index++
		}
		count++
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
