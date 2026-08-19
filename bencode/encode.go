package bencode

import (
	"bytes"
	"fmt"
	"slices"
	"strconv"
)

func Encode(value any) ([]byte, error) {
	var buffer bytes.Buffer

	if err := encodeValue(value, &buffer); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func encodeInt(value int64, buffer *bytes.Buffer) error {
	buffer.WriteString("i")
	buffer.WriteString(strconv.FormatInt(value, 10))
	buffer.WriteString("e")

	return nil
}

func encodeString(value string, buffer *bytes.Buffer) error {
	buffer.WriteString(strconv.Itoa(len(value)))
	buffer.WriteString(":")
	buffer.WriteString(value)

	return nil
}

func encodeList(value []any, buffer *bytes.Buffer) error {
	buffer.WriteString("l")

	for _, item := range value {
		if err := encodeValue(item, buffer); err != nil {
			return err
		}
	}

	buffer.WriteString("e")
	return nil
}

func encodeDictionary(value map[string]any, buffer *bytes.Buffer) error {
	buffer.WriteString("d")

	keys := []string{}

	for key, _ := range value {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	for _, key := range keys {
		err := encodeValue(key, buffer)
		if err != nil {
			return err
		}
		err = encodeValue(value[key], buffer)
		if err != nil {
			return err
		}
	}

	buffer.WriteString("e")
	return nil
}

func encodeValue(value any, buffer *bytes.Buffer) error {
	switch v := value.(type) {
	case int64:
		return encodeInt(v, buffer)
	case string:
		return encodeString(v, buffer)
	case map[string]any:
		return encodeDictionary(v, buffer)
	case []any:
		return encodeList(v, buffer)
	default:
		return fmt.Errorf("unsupported type for bencode encoding: %T", value)
	}
}
