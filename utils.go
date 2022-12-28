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

func filterEmpty(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
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
		if k == "_" {
			continue
		}

		v := filterEmpty(v)
		if len(v) == 0 {
			log.Printf("empty query: %s\n", k)
			continue
		}
		if count > 0 {
			queryBuilder.WriteString(" AND ")
		}
		queryBuilder.WriteString(k)
		if len(v) == 1 {
			queryBuilder.WriteString(" = ")
			queryBuilder.WriteString(fmt.Sprintf("$%d", index))
			args = append(args, v[0])
			index++
		} else {
			queryBuilder.WriteString(" in (")
			for i, vv := range v {
				queryBuilder.WriteString(fmt.Sprintf("$%d", index))
				if i != len(v)-1 {
					queryBuilder.WriteString(",")
				}
				args = append(args, vv)
				index++
			}
			queryBuilder.WriteString(")")
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
