// utility/json.go
package Utility

import (
	"bytes"
	"encoding/json"
)

// PrettyPrint indents a JSON byte slice.
func PrettyPrint(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
}

// ToJson marshals an object into a pretty-printed JSON string.
func ToJson(obj interface{}) (string, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	b_, err := PrettyPrint(b)
	if err != nil {
		return "", err
	}
	return string(b_), nil
}

// ToMap converts any struct/interface into a map[string]interface{}.
func ToMap(in interface{}) (map[string]interface{}, error) {
	jsonStr, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	var out map[string]interface{}
	if err := json.Unmarshal(jsonStr, &out); err != nil {
		return nil, err
	}
	return out, nil
}

