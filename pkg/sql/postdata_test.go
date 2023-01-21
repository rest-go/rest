package sql

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostDataSet(t *testing.T) {
	q := PostData{objects: []map[string]any{
		{
			"a": "b",
		},
		{
			"c": "d",
		},
	}}
	q.Set("user_id", 1)
	for _, object := range q.objects {
		assert.Equal(t, int(1), object["user_id"].(int))
	}
}
func TestPostDataValuesQuery(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		data          []byte
		unmarshalData any
		query         *ValuesQuery
	}{
		{
			name: "unmarshal single",
			data: []byte(`{"name":"hello world", "id":1}`),
			unmarshalData: []map[string]any{
				{
					"name": "hello world",
					"id":   float64(1),
				},
			},
			query: &ValuesQuery{
				Index:        3,
				Columns:      []string{"id", "name"},
				Placeholders: []string{"(?,?)"},
				Args:         []any{"hello world", float64(1)},
			},
		},
		{
			name: "unmarshal many",
			data: []byte(`[{"name":"hello world", "id":1}, {"name":"rest-go", "id":2}]`),
			unmarshalData: []map[string]any{
				{
					"name": "hello world",
					"id":   float64(1),
				},
				{
					"name": "rest-go",
					"id":   float64(2),
				},
			},
			query: &ValuesQuery{
				Index:        5,
				Columns:      []string{"id", "name"},
				Placeholders: []string{"(?,?)", "(?,?)"},
				Args:         []any{"hello world", float64(1), "rest-go", float64(2)},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var data PostData
			err := json.Unmarshal(test.data, &data)
			assert.Nil(t, err)
			assert.ElementsMatch(t, test.unmarshalData, data.objects)
			query, err := data.ValuesQuery()
			assert.Nil(t, err)
			assert.Equal(t, test.query.Index, query.Index, "index not equal")
			assert.ElementsMatch(t, test.query.Columns, query.Columns, "columns not equal")
			assert.ElementsMatch(t, test.query.Placeholders, query.Placeholders, "placeholders not equal")
			assert.ElementsMatch(t, test.query.Args, query.Args, "args not equal")
		})
	}

	t.Run("inconsistency columns with different name", func(t *testing.T) {
		var data PostData
		err := json.Unmarshal([]byte(`[{"name":"hello world", "id":1}, {"name2":"rest-go", "id":2}, {"id":3}]`), &data)
		assert.Nil(t, err)
		assert.Equal(t, 3, len(data.objects))
		_, err = data.ValuesQuery()
		assert.NotNil(t, err)
	})

	t.Run("inconsistency columns with different length", func(t *testing.T) {
		var data PostData
		err := json.Unmarshal([]byte(`[{"name":"hello world", "id":1}, {"id":2}, {"id":3}]`), &data)
		assert.Nil(t, err)
		assert.Equal(t, 3, len(data.objects))
		_, err = data.ValuesQuery()
		assert.NotNil(t, err)
	})

	t.Run("invalid json format data", func(t *testing.T) {
		var data PostData
		err := json.Unmarshal([]byte("1234"), &data)
		assert.NotNil(t, err)
		assert.Equal(t, []map[string]interface{}(nil), data.objects)
	})
}

func TestPostDataSetQuery(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		data          []byte
		unmarshalData any
		query         *SetQuery
	}{
		{
			name: "single",
			data: []byte(`{"name":"hello world", "id":1}`),
			unmarshalData: []map[string]any{
				{
					"name": "hello world",
					"id":   float64(1),
				},
			},
			query: &SetQuery{
				Index: 3,
				Query: "name = ?, id = ?",
				Args:  []any{"hello world", float64(1)},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var data PostData
			err := json.Unmarshal(test.data, &data)
			assert.Nil(t, err)
			assert.ElementsMatch(t, test.unmarshalData, data.objects)
			query, err := data.SetQuery(1)
			assert.Nil(t, err)
			assert.Equal(t, test.query.Index, query.Index, "index not equal")
			// order is undetermined
			// assert.ElementsMatch(
			// 	t,
			// 	sort.StringSlice(strings.Split(test.query.Query, ",")),
			// 	sort.StringSlice(strings.Split(query.Query, ",")),
			// 	"query not equal",
			// )
			assert.ElementsMatch(t, test.query.Args, query.Args, "args not equal")
		})
	}
	t.Run("bulk update are not updated", func(t *testing.T) {
		var data PostData
		err := json.Unmarshal([]byte(`[{"name":"hello", "id":1}, {"name":"world", "id":2}]`), &data)
		assert.Nil(t, err)
		_, err = data.SetQuery(1)
		assert.NotNil(t, err)
	})
}
