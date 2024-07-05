package sql

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestURLQueryJSON test json related queries
// https://www.postgresql.org/docs/12/functions-json.html
// https://dev.mysql.com/doc/refman/8.0/en/json.html
// https://www.sqlite.org/json1.html
func TestURLQueryJSON(t *testing.T) {
	for _, test := range []struct {
		driver      string
		jsonPath    string
		selectQuery string
		whereQuery  string
	}{
		{
			driver:      "postgres",
			jsonPath:    "object->1->field1->field2->>2",
			selectQuery: "object->1->'field1'->'field2'->>2 AS field2",
			whereQuery:  "object->1->'field1'->'field2'->>2 = ?",
		},
		{
			driver:      "mysql",
			jsonPath:    "object->1->field1->field2->>2",
			selectQuery: "object->'$[1].field1.field2[2]' AS field2",
			whereQuery:  "object->'$[1].field1.field2[2]' = ?",
		},
		{
			driver:      "sqlite",
			jsonPath:    "object->1->field1->field2->>2",
			selectQuery: "object->1->'field1'->'field2'->>2 AS field2",
			whereQuery:  "object->1->'field1'->'field2'->>2 = ?",
		},
	} {
		t.Run(test.driver+" select", func(t *testing.T) {
			v := url.Values{"select": []string{test.jsonPath}}
			q := NewURLQuery(v, test.driver)
			query, err := q.SelectQuery()
			assert.Nil(t, err)
			assert.Equal(t, test.selectQuery, query)
		})
		t.Run(test.driver+" where", func(t *testing.T) {
			v := url.Values{test.jsonPath: []string{"eq.1"}}
			q := NewURLQuery(v, test.driver)
			index, query, args := q.WhereQuery(1)
			assert.Equal(t, uint(2), index)
			assert.Equal(t, test.whereQuery, query)
			assert.Equal(t, []any{"1"}, args)
		})
	}
}

func TestURLQuerySet(t *testing.T) {
	q := URLQuery{values: url.Values{}}
	q.Set("a", "b")
	assert.Equal(t, []string{"b"}, q.values["a"])
}

func TestURLQuerySelectQuery(t *testing.T) {
	v := url.Values{}
	q := NewURLQuery(v, "")
	query, err := q.SelectQuery()
	assert.Nil(t, err)
	assert.Equal(t, "*", query)

	v = url.Values{"select": []string{"a,b"}}
	q = NewURLQuery(v, "")
	query, err = q.SelectQuery()
	assert.Nil(t, err)
	assert.Equal(t, "a,b", query)

	t.Run("allowed func", func(t *testing.T) {
		v := url.Values{"select": []string{"MAX(a)"}}
		q := NewURLQuery(v, "")
		query, err := q.SelectQuery()
		assert.Nil(t, err)
		assert.Equal(t, "MAX(a) AS max", query)
	})

	t.Run("not allowed func", func(t *testing.T) {
		v := url.Values{"select": []string{"setting(a)"}}
		q := NewURLQuery(v, "")
		query, err := q.SelectQuery()
		assert.NotNil(t, err)
		assert.Equal(t, "", query)
	})
}

func TestURLQueryOrderQuery(t *testing.T) {
	v := url.Values{}
	q := NewURLQuery(v, "")
	query := q.OrderQuery()
	assert.Equal(t, "", query)

	v = url.Values{"order": []string{"a.desc,b.asc"}}
	q = NewURLQuery(v, "")
	query = q.OrderQuery()
	assert.Equal(t, "a desc,b asc", query)

	v = url.Values{"order": []string{"a.desc,b.asc;xxx"}}
	q = NewURLQuery(v, "")
	query = q.OrderQuery()
	assert.Equal(t, "", query)
}

// WhereQuery returns sql and args for where clause
func TestURLQueryWhereQuery(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		v := url.Values{}
		q := NewURLQuery(v, "sqlite")
		index, query, args := q.WhereQuery(1)
		assert.Equal(t, uint(1), index)
		assert.Equal(t, "", query)
		assert.Equal(t, 0, len(args))
	})

	t.Run("skip no op query", func(t *testing.T) {
		v := url.Values{"select": []string{"*"}, "count": []string{""}, "noop": []string{"noop.1"}, "invalid_val": []string{"a.b.c=1"}}
		q := NewURLQuery(v, "sqlite")
		index, query, args := q.WhereQuery(1)
		assert.Equal(t, uint(1), index)
		assert.Equal(t, "", query)
		assert.Equal(t, 0, len(args))
	})

	t.Run("skip invalid character", func(t *testing.T) {
		v := url.Values{"select": []string{"a;xxx"}}
		q := NewURLQuery(v, "sqlite")
		index, query, args := q.WhereQuery(1)
		assert.Equal(t, uint(1), index)
		assert.Equal(t, "", query)
		assert.Equal(t, 0, len(args))
	})

	t.Run("operators", func(t *testing.T) {
		for op, operator := range Operators {
			if op == "in" || op == "is" {
				continue
			}
			v := url.Values{"a": []string{fmt.Sprintf("%s.1", op)}}
			q := NewURLQuery(v, "sqlite")
			index, query, args := q.WhereQuery(1)
			assert.Equal(t, uint(2), index)
			assert.Equal(t, fmt.Sprintf("a%s?", operator), query)
			assert.Equal(t, 1, len(args))
		}

		v := url.Values{"a": []string{"in.(1,2)"}}
		q := NewURLQuery(v, "sqlite")
		index, query, args := q.WhereQuery(1)
		assert.Equal(t, uint(3), index)
		assert.Equal(t, "a IN (?,?)", query)
		assert.Equal(t, 2, len(args))

		v = url.Values{"a": []string{"is.null"}}
		q = NewURLQuery(v, "sqlite")
		index, query, args = q.WhereQuery(1)
		assert.Equal(t, uint(1), index)
		assert.Equal(t, "a is null", query)
		assert.Equal(t, 0, len(args))

		v = url.Values{"a": []string{"gt.1", "lt.100"}}
		q = NewURLQuery(v, "sqlite")
		index, query, args = q.WhereQuery(1)
		assert.Equal(t, uint(3), index)
		assert.Equal(t, "a > ? AND a < ?", query)
		assert.Equal(t, 2, len(args))
	})

	t.Run("AND", func(t *testing.T) {
		v := url.Values{"a": []string{"eq.1"}, "b": []string{"eq.2"}}
		q := NewURLQuery(v, "sqlite")
		index, query, args := q.WhereQuery(1)
		assert.Equal(t, uint(3), index)
		assert.Contains(t, query, " AND ")
		assert.Equal(t, 2, len(args))
	})
}

func TestURLQueryPage(t *testing.T) {
	v := url.Values{}
	q := NewURLQuery(v, "")
	page, pageSize := q.Page()
	assert.Equal(t, 1, page)
	assert.Equal(t, 100, pageSize)

	v = url.Values{"page": []string{"2"}, "page_size": []string{"20"}}
	q = NewURLQuery(v, "")
	page, pageSize = q.Page()
	assert.Equal(t, 2, page)
	assert.Equal(t, 20, pageSize)
}

func TestURLQueryIsDebug(t *testing.T) {
	v := url.Values{}
	q := NewURLQuery(v, "")
	assert.False(t, q.IsDebug())

	v = url.Values{"debug": []string{"1"}}
	q = NewURLQuery(v, "")
	assert.True(t, q.IsDebug())
}

func TestURLQueryIsCount(t *testing.T) {
	v := url.Values{}
	q := NewURLQuery(v, "")
	assert.False(t, q.IsCount())

	v = url.Values{"count": []string{"1"}}
	q = NewURLQuery(v, "")
	assert.True(t, q.IsCount())
}

func TestURLQueryIsSingular(t *testing.T) {
	v := url.Values{}
	q := NewURLQuery(v, "")
	assert.False(t, q.IsSingular())

	v = url.Values{"singular": []string{"1"}}
	q = NewURLQuery(v, "")
	assert.True(t, q.IsSingular())
}

func TestURLQueryIsMine(t *testing.T) {
	v := url.Values{}
	q := NewURLQuery(v, "")
	assert.False(t, q.IsMine())

	v = url.Values{"mine": []string{"1"}}
	q = NewURLQuery(v, "")
	assert.True(t, q.IsMine())
}
