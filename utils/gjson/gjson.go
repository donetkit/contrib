package gjson

import "encoding/json"

func Marshal(v any) string {
	val, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(val)
}

func MarshalToByte(v any) []byte {
	val, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return val
}

func Unmarshal(data string, v any) error {
	return json.Unmarshal([]byte(data), v)
}

func UnmarshalToByte(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
