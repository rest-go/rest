package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// e.g. INSERT INTO a (c1, c2) VALUES (v1,v2),(v3,v4)
// index=4
// columns=["c1", "c2"]
// vals=["$1,$2", "$3,$4"]
// args=[v1,v2,v3,v4]
type PostQuery struct {
	index   int // index for next field, args number plus 1
	columns []string
	vals    []string
	args    []any
}

type PostData struct {
	objects []map[string]any
}

// valuesQuery convert post data to values query for insertion
func (pd *PostData) valuesQuery() (*PostQuery, error) {
	objects := pd.objects
	if len(objects) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	// use first object's keys as columns
	columns := make([]string, 0, len(objects[0]))
	for k := range objects[0] {
		columns = append(columns, k)
	}

	// build vals and args
	vals := make([]string, 0, len(objects))
	args := make([]any, 0, cap(columns)*cap(vals))
	index := 1
	for i, object := range objects {
		val := make([]string, 0, len(object))
		if i > 0 && !identKeys(object, columns) {
			// validate object's keys with columns
			return nil, fmt.Errorf("columns must be same for all objects, invalid object: %v", object)
		}
		// consistent column order with first object
		for _, c := range columns {
			val = append(val, fmt.Sprintf("$%d", index))
			args = append(args, object[c])
			index++
		}
		vals = append(vals, fmt.Sprintf("(%s)", strings.Join(val, ",")))
	}
	return &PostQuery{
		index, columns, vals, args,
	}, nil
}

// UnmarshalJSON implements json.Unmarshaler
func (pd *PostData) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		return fmt.Errorf("no bytes to unmarshal")
	}
	// guess based on the first character
	switch b[0] {
	case '{':
		return pd.unmarshalSingle(b)
	case '[':
		return pd.unmarshalMany(b)
	}
	// This shouldn't really happen as the standard library seems to strip
	// whitespace from the bytes being passed in, but just in case let's guess at
	// multiple tags and fall back to a single one if that doesn't work.
	err := pd.unmarshalMany(b)
	if err != nil {
		return pd.unmarshalSingle(b)
	}
	return nil
}

func (pd *PostData) unmarshalSingle(b []byte) error {
	var data map[string]any
	err := json.Unmarshal(b, &data)
	if err != nil {
		return fmt.Errorf("failed to unmarshal single data, %v", err)
	}

	pd.objects = []map[string]any{data}
	return nil
}

func (pd *PostData) unmarshalMany(b []byte) error {
	var data []map[string]any
	err := json.Unmarshal(b, &data)
	if err != nil {
		return fmt.Errorf("failed to unmarshal many data, %v", err)
	}

	pd.objects = data
	return nil
}
