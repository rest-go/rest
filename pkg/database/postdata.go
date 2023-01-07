package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// e.g. INSERT INTO a (c1, c2) VALUES (v1,v2),(v3,v4)
// index=4
// columns=["c1", "c2"]
// vals=["$1,$2", "$3,$4"]
// args=[v1,v2,v3,v4]
type ValuesQuery struct {
	Index   uint // index for next field, args number plus 1
	Columns []string
	Vals    []string
	Args    []any
}

// e.g. UPDATE table SET a="a",b="b"
// index=3
// sql="a=$1, b=$2"
// args=["a", "b"]
type SetQuery struct {
	Index uint // index for next field, args number plus 1
	Query string
	Args  []any
}

type PostData struct {
	driver  string
	objects []map[string]any
}

func (pd *PostData) WithDriver(driver string) {
	pd.driver = driver
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

// valuesQuery convert post data to values query for insertion
func (pd *PostData) ValuesQuery() (*ValuesQuery, error) {
	objects := pd.objects

	// use first object's keys as columns
	columns := make([]string, 0, len(objects[0]))
	for k := range objects[0] {
		columns = append(columns, k)
	}

	// build vals and args
	vals := make([]string, 0, len(objects))
	args := make([]any, 0, cap(columns)*cap(vals))
	var index uint = 1
	for i, object := range objects {
		val := make([]string, 0, len(object))
		if i > 0 && !identKeys(object, columns) {
			// validate object's keys with columns
			return nil, fmt.Errorf("columns must be same for all objects, invalid object: %v", object)
		}
		// consistent column order with first object
		for _, c := range columns {
			val = append(val, placeholder(pd.driver, index))
			args = append(args, object[c])
			index++
		}
		vals = append(vals, fmt.Sprintf("(%s)", strings.Join(val, ",")))
	}
	return &ValuesQuery{
		index, columns, vals, args,
	}, nil
}

// SetQuery return set sql for update
// TODO: bulk update
func (pd *PostData) SetQuery(index uint) (*SetQuery, error) {
	if len(pd.objects) != 1 {
		return nil, errors.New("wrong set data")
	}

	data := pd.objects[0]
	var queryBuilder strings.Builder
	args := make([]any, 0, len(data))
	first := true
	for k, v := range data {
		if !first {
			queryBuilder.WriteString(", ")
		}
		queryBuilder.WriteString(k)
		queryBuilder.WriteString(" = ")
		queryBuilder.WriteString(placeholder(pd.driver, index))
		args = append(args, v)
		index++
		first = false
	}
	query := queryBuilder.String()
	return &SetQuery{
		index,
		query,
		args,
	}, nil
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
