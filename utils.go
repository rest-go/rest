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
