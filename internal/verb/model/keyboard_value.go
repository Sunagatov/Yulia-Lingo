package model

import "encoding/json"

type KeyboardVerbValue struct {
	Request string
	Page    int64
	Latter  string
}

func (kv KeyboardVerbValue) ToJSON() string {
	jsonBytes, err := json.Marshal(kv)
	if err != nil {
		return ""
	}
	return string(jsonBytes)
}

func (kv KeyboardVerbValue) FromJSON(jsonStr string) (KeyboardVerbValue, error) {
	err := json.Unmarshal([]byte(jsonStr), &kv)
	if err != nil {
		return KeyboardVerbValue{}, err
	}
	return kv, nil
}
