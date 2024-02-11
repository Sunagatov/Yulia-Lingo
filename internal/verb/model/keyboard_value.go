package model

import "encoding/json"

type KeyboardVerbValue struct {
	Request string
	Page    int64
	Latter  string
}

func KeyboardVerbValueFromJSON(jsonStr string) (KeyboardVerbValue, error) {
	var kv KeyboardVerbValue
	err := json.Unmarshal([]byte(jsonStr), &kv)
	if err != nil {
		return KeyboardVerbValue{}, err
	}
	return kv, nil
}
