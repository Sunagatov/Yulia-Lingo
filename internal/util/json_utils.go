package util

import (
	"encoding/json"
	"fmt"
	"strings"
)

func ConvertToJSON(entity interface{}) (string, error) {
	jsonBytes, err := json.Marshal(entity)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %w", err)
	}
	return string(jsonBytes), nil
}

func GetMessageDelimiter() string {
	return strings.Repeat("-", 30)
}