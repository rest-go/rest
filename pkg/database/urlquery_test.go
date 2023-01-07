package database

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURLQuerySelectQuery(t *testing.T) {
	v := url.Values{}
	q := URLQuery(v)
	query := q.SelectQuery()
	assert.Equal(t, "*", query)

	v = url.Values{"select": []string{"a,b,object->>field"}}
	q = URLQuery(v)
	query = q.SelectQuery()
	assert.Equal(t, "a,b,object->>'field' AS field", query)
}

func TestURLQueryOrderQuery(t *testing.T) {
	v := url.Values{}
	q := URLQuery(v)
	query := q.OrderQuery()
	assert.Equal(t, "", query)

	v = url.Values{"order": []string{"a.desc,b.asc"}}
	q = URLQuery(v)
	query = q.OrderQuery()
	assert.Equal(t, "a desc,b asc", query)
}

// WhereQuery returns sql and args for where clause
func TestURLQueryWhereQuery(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		v := url.Values{}
		q := URLQuery(v)
		index, query, args := q.WhereQuery(1)
		assert.Equal(t, uint(1), index)
		assert.Equal(t, "", query)
		assert.Equal(t, 0, len(args))
	})

	t.Run("operators", func(t *testing.T) {
		for op, operator := range Operators {
			if op == "in" {
				continue
			}
			v := url.Values{"a": []string{fmt.Sprintf("%s.1", op)}}
			q := URLQuery(v)
			index, query, args := q.WhereQuery(1)
			assert.Equal(t, uint(2), index)
			assert.Equal(t, fmt.Sprintf("a%s$1", operator), query)
			assert.Equal(t, 1, len(args))
		}

		v := url.Values{"a": []string{"in.(1,2)"}}
		q := URLQuery(v)
		index, query, args := q.WhereQuery(1)
		assert.Equal(t, uint(1), index)
		assert.Equal(t, "a IN (1,2)", query)
		assert.Equal(t, 0, len(args))
	})

	t.Run("AND", func(t *testing.T) {
		v := url.Values{"a": []string{"eq.1"}, "b": []string{"eq.2"}}
		q := URLQuery(v)
		index, query, args := q.WhereQuery(1)
		assert.Equal(t, uint(3), index)
		assert.Contains(t, query, " AND ")
		assert.Equal(t, 2, len(args))
	})

	t.Run("json", func(t *testing.T) {
		v := url.Values{"object->>field": []string{"eq.1"}}
		q := URLQuery(v)
		index, query, args := q.WhereQuery(1)
		assert.Equal(t, uint(2), index)
		assert.Equal(t, "object->>'field' = $1", query)
		assert.Equal(t, 1, len(args))
	})
}

func TestURLQueryPage(t *testing.T) {
	v := url.Values{}
	q := URLQuery(v)
	page, pageSize := q.Page()
	assert.Equal(t, 1, page)
	assert.Equal(t, 100, pageSize)

	v = url.Values{"page": []string{"2"}, "page_size": []string{"20"}}
	q = URLQuery(v)
	page, pageSize = q.Page()
	assert.Equal(t, 2, page)
	assert.Equal(t, 20, pageSize)
}

func TestURLQueryIsDebug(t *testing.T) {
	v := url.Values{}
	q := URLQuery(v)
	assert.False(t, q.IsDebug())

	v = url.Values{"debug": []string{"1"}}
	q = URLQuery(v)
	assert.True(t, q.IsDebug())
}

func TestURLQueryIsCount(t *testing.T) {
	v := url.Values{}
	q := URLQuery(v)
	assert.False(t, q.IsCount())

	v = url.Values{"count": []string{"1"}}
	q = URLQuery(v)
	assert.True(t, q.IsCount())
}

func TestURLQueryIsSingular(t *testing.T) {
	v := url.Values{}
	q := URLQuery(v)
	assert.False(t, q.IsSingular())

	v = url.Values{"singular": []string{"1"}}
	q = URLQuery(v)
	assert.True(t, q.IsSingular())
}
