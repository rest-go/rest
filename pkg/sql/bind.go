package sql

import (
	"strconv"
	"strings"
)

// Bindvar types supported by Rebind, BindMap and BindStruct.
const (
	UNKNOWN = iota
	QUESTION
	DOLLAR
)

var binds = map[string]int{
	"postgres": DOLLAR,
	"mysql":    QUESTION,
	"sqlite":   QUESTION,
}

// Rebind a query from the default driverName(mysql) to the target driverName.
func Rebind(driverName, query string) string {
	bindType := binds[driverName]
	switch bindType {
	case QUESTION, UNKNOWN:
		return query
	}

	rqb := make([]byte, 0, len(query))
	var i, j int
	for i = strings.Index(query, "?"); i != -1; i = strings.Index(query, "?") {
		rqb = append(rqb, query[:i]...)

		switch bindType { //nolint:gocritic
		case DOLLAR:
			rqb = append(rqb, '$')
		}

		j++
		rqb = strconv.AppendInt(rqb, int64(j), 10)

		query = query[i+1:]
	}

	return string(append(rqb, query...))
}
