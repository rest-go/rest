package jsonutil

import (
	"encoding/json"
)

// MapToStruct converts a map[string]any to a struct
// target is a pointer of the target struct
func MapToStruct(dict map[string]any, target any) error {
	jsonbody, err := json.Marshal(dict)
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonbody, target)
}
