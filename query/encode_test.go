package query

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

type stringStruct struct {
	String string `query:"string,omitempty"`
}

type customEncoderStruct struct {
	String string
}

func (s *customEncoderStruct) EncodeValues() (url.Values, error) {
	return url.Values{
		"custom": []string{s.String},
	}, nil
}

type nestedCustomEncoderStruct struct {
	Custom customEncoderStruct
}

type nestedPointerCustomEncoderStruct struct {
	Custom *customEncoderStruct
}

type customValueEncoderStruct struct {
	String string
}

func (s customValueEncoderStruct) EncodeValues() (url.Values, error) {
	return url.Values{
		"custom": []string{s.String},
	}, nil
}

type nestedCustomValueEncoderStruct struct {
	Custom customValueEncoderStruct
}

type nestedPointerCustomValueEncoderStruct struct {
	Custom *customValueEncoderStruct
}

type ignoreStruct struct {
	String string `query:"-"`
}

func TestEncode(t *testing.T) {
	tests := []struct {
		name          string
		obj           any
		errorExpected bool
		values        url.Values
	}{
		{
			name: "sample objects",
			obj: stringStruct{
				String: "hello world",
			},
			errorExpected: false,
			values: map[string][]string{
				"string": {"hello world"},
			},
		},
		{
			name: "pointer to sample objects",
			obj: &stringStruct{
				String: "hello world",
			},
			errorExpected: false,
			values: map[string][]string{
				"string": {"hello world"},
			},
		},
		{
			name:          "invalid type",
			obj:           42,
			errorExpected: true,
			values: map[string][]string{
				"string": {"hello world"},
			},
		},
		{
			name: "custom encoder",
			obj: customEncoderStruct{
				String: "hello world",
			},
			errorExpected: false,
			values: map[string][]string{
				"custom": {"hello world"},
			},
		},
		{
			name: "custom encoder pointer",
			obj: &customEncoderStruct{
				String: "hello world",
			},
			errorExpected: false,
			values: map[string][]string{
				"custom": {"hello world"},
			},
		},
		{
			name: "nested custom encoder",
			obj: nestedCustomEncoderStruct{
				Custom: customEncoderStruct{
					String: "hello world",
				},
			},
			errorExpected: false,
			values: map[string][]string{
				"custom": {"hello world"},
			},
		},
		{
			name: "nested pointed custom encoder",
			obj: nestedPointerCustomEncoderStruct{
				Custom: &customEncoderStruct{
					String: "hello world",
				},
			},
			errorExpected: false,
			values: map[string][]string{
				"custom": {"hello world"},
			},
		},
		{
			name: "custom value encoder pointer",
			obj: customValueEncoderStruct{
				String: "hello world",
			},
			errorExpected: false,
			values: map[string][]string{
				"custom": {"hello world"},
			},
		},
		{
			name: "nested custom value encoder",
			obj: nestedCustomValueEncoderStruct{
				Custom: customValueEncoderStruct{
					String: "hello world",
				},
			},
			errorExpected: false,
			values: map[string][]string{
				"custom": {"hello world"},
			},
		},
		{
			name: "nested pointed custom encoder",
			obj: nestedPointerCustomValueEncoderStruct{
				Custom: &customValueEncoderStruct{
					String: "hello world",
				},
			},
			errorExpected: false,
			values: map[string][]string{
				"custom": {"hello world"},
			},
		},
		{
			name: "full test struct",
			obj: testStruct{
				String:  "hello world",
				Bool:    false,
				Float64: 42.2,
				Float32: 43.4,
				Int:     1,
				Int64:   2,
				Int32:   3,
				Int16:   4,
				Int8:    5,
				Uint:    6,
				Uint64:  7,
				Uint32:  8,
				Uint16:  9,
				Uint8:   10,
			},
			errorExpected: false,
			values: map[string][]string{
				"string":  {"hello world"},
				"bool":    {"false"},
				"float64": {"42.2"},
				"float32": {"43.4"},
				"int":     {"1"},
				"int64":   {"2"},
				"int32":   {"3"},
				"int16":   {"4"},
				"int8":    {"5"},
				"uint":    {"6"},
				"uint64":  {"7"},
				"uint32":  {"8"},
				"uint16":  {"9"},
				"uint8":   {"10"},
			},
		},
		{
			name: "pointer test struct",
			obj: pointerTestStruct{
				String:  toPointer("hello world"),
				Bool:    toPointer(false),
				Float64: toPointer(42.2),
				Float32: toPointer[float32](43.4),
				Int:     toPointer(1),
				Int64:   toPointer[int64](2),
				Int32:   toPointer[int32](3),
				Int16:   toPointer[int16](4),
				Int8:    toPointer[int8](5),
				Uint:    toPointer[uint](6),
				Uint64:  toPointer[uint64](7),
				Uint32:  toPointer[uint32](8),
				Uint16:  toPointer[uint16](9),
				Uint8:   toPointer[uint8](10),
			},
			errorExpected: false,
			values: map[string][]string{
				"string":  {"hello world"},
				"bool":    {"false"},
				"float64": {"42.2"},
				"float32": {"43.4"},
				"int":     {"1"},
				"int64":   {"2"},
				"int32":   {"3"},
				"int16":   {"4"},
				"int8":    {"5"},
				"uint":    {"6"},
				"uint64":  {"7"},
				"uint32":  {"8"},
				"uint16":  {"9"},
				"uint8":   {"10"},
			},
		},
		{
			name: "slices struct",
			obj: slicesStruct{
				Strings:  []string{"hello", "world"},
				Bools:    []bool{true, false},
				Float64s: []float64{42.2, 43.4},
				Float32s: []float32{43.4},
				Ints:     []int{1, 2, 3},
				Int64s:   []int64{4, 5, 6},
				Int32s:   []int32{7, 8, 9},
				Int16s:   []int16{10, 11, 12},
				Int8s:    []int8{11, 12, 13},
				Uints:    []uint{1, 2, 3},
				Uint64s:  []uint64{4, 5, 6},
				Uint32s:  []uint32{7, 8, 9},
				Uint16s:  []uint16{10, 11, 12},
				Uint8s:   []uint8{11, 12, 13},
			},
			errorExpected: false,
			values: map[string][]string{
				"strings":  {"hello", "world"},
				"bools":    {"true", "false"},
				"float64s": {"42.2", "43.4"},
				"float32s": {"43.4"},
				"ints":     {"1", "2", "3"},
				"int64s":   {"4", "5", "6"},
				"int32s":   {"7", "8", "9"},
				"int16s":   {"10", "11", "12"},
				"int8s":    {"11", "12", "13"},
				"uints":    {"1", "2", "3"},
				"uint64s":  {"4", "5", "6"},
				"uint32s":  {"7", "8", "9"},
				"uint16s":  {"10", "11", "12"},
				"uint8s":   {"11", "12", "13"},
			},
		},
		{
			name: "non exported field",
			obj: nonExportedStruct{
				int: 42,
			},
			errorExpected: false,
			values:        map[string][]string{},
		},
		{
			name:          "omit empty field",
			obj:           stringStruct{},
			errorExpected: false,
			values:        map[string][]string{},
		},
		{
			name: "omit field with -",
			obj: ignoreStruct{
				String: "hello world",
			},
			errorExpected: false,
			values:        map[string][]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			values, err := Encode(tt.obj)

			if tt.errorExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.values, values)
			}
		})
	}
}
