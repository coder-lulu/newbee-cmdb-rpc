package jsonhelper

import (
	"encoding/json"
	"github.com/suyuan32/simple-admin-common/utils/pointy"
)

// StringToJSONMap converts a JSON string to map[string]interface{}
func StringToJSONMap(data *string) map[string]interface{} {
	if data == nil || *data == "" {
		return nil
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(*data), &result); err != nil {
		return nil
	}
	return result
}

// StringToJSONArray converts a JSON string to []interface{}
func StringToJSONArray(data *string) []interface{} {
	if data == nil || *data == "" {
		return nil
	}
	var result []interface{}
	if err := json.Unmarshal([]byte(*data), &result); err != nil {
		return nil
	}
	return result
}

// StringToStringArray converts a JSON string to []string
func StringToStringArray(data *string) []string {
	if data == nil || *data == "" {
		return nil
	}
	var result []string
	if err := json.Unmarshal([]byte(*data), &result); err != nil {
		return nil
	}
	return result
}

// JSONToString converts any JSON data to string pointer
func JSONToString(data interface{}) *string {
	if data == nil {
		return nil
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return pointy.GetPointer("")
	}
	return pointy.GetPointer(string(bytes))
}