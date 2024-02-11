package service

import "encoding/json"

func ToJSON(entity interface{}) string {
	jsonBytes, err := json.Marshal(entity)
	if err != nil {
		return ""
	}
	return string(jsonBytes)
}
