package bencode

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

func Decode(file io.Reader) (any, error) {
	reader := bufio.NewReader(file)

	val, err := decodeValue(reader)
	if err != nil {
		fmt.Print(err)
		return nil, err
	}
	return val, nil
}

func decodeInt(reader *bufio.Reader) (int64, error) {
	str, err := reader.ReadString('e')
	if err != nil {
		return 0, err
	}

	number, err := strconv.ParseInt(str[:len(str)-1], 10, 64)
	if err != nil {
		return 0, err
	}

	return number, nil
}

func decodeString(reader *bufio.Reader) (string, error) {
	str, err := reader.ReadString(':')
	if err != nil {
		return "", err
	}

	length, err := strconv.Atoi(str[:len(str)-1])
	if err != nil {
		return "", err
	}

	buffer := make([]byte, length)

	_, err = io.ReadFull(reader, buffer)
	if err != nil {
		return "", err
	}

	return string(buffer), nil
}

func decodeList(reader *bufio.Reader) ([]any, error) {
	list := []any{}
	for {
		b, err := reader.Peek(1)
		if err != nil {
			return list, err
		}

		if b[0] == 'e' {
			reader.ReadByte()
			return list, nil
		}

		val, err := decodeValue(reader)
		if err != nil {
			return list, err
		}

		list = append(list, val)
	}
}

func decodeDictionary(reader *bufio.Reader) (map[string]any, error) {
	dictionary := make(map[string]any)
	for {
		b, err := reader.Peek(1)
		if err != nil {
			return dictionary, err
		}

		if b[0] == 'e' {
			reader.ReadByte()
			return dictionary, nil
		}

		keyAny, err := decodeValue(reader)
		if err != nil {
			return dictionary, err
		}

		key, ok := keyAny.(string)
		if !ok {
			return dictionary, fmt.Errorf("Key is not a string")
		}

		val, err := decodeValue(reader)
		if err != nil {
			return dictionary, err
		}

		dictionary[key] = val

	}
}

func decodeValue(reader *bufio.Reader) (any, error) {
	b, err := reader.Peek(1)
	if err != nil {
		return nil, err
	}
	firstChar := string(b)
	switch firstChar {
	case "i":
		reader.ReadByte()
		val, err := decodeInt(reader)
		if err != nil {
			return nil, err
		}
		return val, nil

	case "l":
		reader.ReadByte()
		list, err := decodeList(reader)
		if err != nil {
			return nil, err
		}
		return list, nil

	case "d":
		reader.ReadByte()
		dictionary, err := decodeDictionary(reader)
		if err != nil {
			return nil, err
		}
		return dictionary, nil

	default:
		str, err := decodeString(reader)
		if err != nil {
			return nil, err
		}
		return str, nil
	}
}
