package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostData(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		data          []byte
		unmarshalData any
		query         *PostQuery
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
			query: &PostQuery{
				index:   3,
				columns: []string{"id", "name"},
				vals:    []string{"($1,$2)"},
				args:    []any{"hello world", float64(1)},
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
			query: &PostQuery{
				index:   5,
				columns: []string{"id", "name"},
				vals:    []string{"($1,$2)", "($3,$4)"},
				args:    []any{"hello world", float64(1), "rest-go", float64(2)},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var data PostData
			err := json.Unmarshal(test.data, &data)
			assert.Nil(t, err)
			assert.ElementsMatch(t, test.unmarshalData, data.objects)
			query, err := data.valuesQuery()
			assert.Nil(t, err)
			assert.Equal(t, test.query.index, query.index, "index not equal")
			assert.ElementsMatch(t, test.query.columns, query.columns, "columns not equal")
			assert.ElementsMatch(t, test.query.vals, query.vals, "vals not equal")
			assert.ElementsMatch(t, test.query.args, query.args, "args not equal")
		})
	}
	t.Run("invalid json data", func(t *testing.T) {
		var data PostData
		err := json.Unmarshal([]byte("{"), &data)
		assert.NotNil(t, err)
		assert.Equal(t, []map[string]interface{}(nil), data.objects)
	})

}
