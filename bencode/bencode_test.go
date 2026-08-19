package bencode

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"Encode Int", 42, "i42e"},
		{"Encode Negative Int", -10, "i-10e"},
		{"Encode String", "hello", "5:hello"},
		{"Encode List", []any{"spam", 42}, "l4:spami42ee"},
		{"Encode Dictionary", map[string]any{"c": "b", "a": 99}, "d1:ai99e1:c1:be"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := Encode(tt.input)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			resString := string(res)
			if resString != tt.expected {
				t.Errorf("Expected: %q, but got %q", tt.expected, resString)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected any
	}{
		{"Decode Int", "i42e", 42},
		{"Decode String", "5:hello", "hello"},
		{"Decode List", "l4:spami42ee", []any{"spam", 42}},
		{"Decode Dictionary", "d3:bazi99e3:foo3:bare", map[string]any{"baz": 99, "foo": "bar"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)

			res, err := Decode(reader)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !reflect.DeepEqual(res, tt.expected) {
				t.Errorf("Expected: %q, but got %q", tt.expected, res)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	tests := []any{
		31,
		"spam",
		[]any{"spam", 33},
		map[string]any{"cow": "moo", "spam": "bike"},
	}

	for _, c := range tests {
		encoded, _ := Encode(c)
		decoded, _ := Decode(bytes.NewReader(encoded))
		if !reflect.DeepEqual(c, decoded) {
			t.Errorf("Round trip failed for %v: got %v", c, decoded)
		}
	}
}
