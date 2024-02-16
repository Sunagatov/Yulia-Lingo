package util_services

import (
	"encoding/json"
	"fmt"
	"strings"
)

func ConvertToJson(entity interface{}) (string, error) {
	jsonBytes, err := json.Marshal(entity)
	if err != nil {
		return "", fmt.Errorf("failed to create a json: %v", err)
	}
	return string(jsonBytes), nil
}

func GetMessageDelimiter() string {
	return strings.Repeat("-", 30)
}
