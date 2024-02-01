package gjson

import (
	"bytes"
	"github.com/goccy/go-json"
)

func Marshal(v any) string {
	val, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(val)
}

func MarshalEscapeHTML(v any) string {
	buff := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(buff)
	jsonEncoder.SetEscapeHTML(false)
	jsonEncoder.Encode(v)
	return buff.String()
}

func MarshalIndent(v any) string {
	val, err := json.MarshalIndent(v, "", "\t")
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
