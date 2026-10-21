package query

import (
	"fmt"
	"math"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	String  string
	Bool    bool
	Float64 float64
	Float32 float32
	Int     int
	Int64   int64
	Int32   int32
	Int16   int16
	Int8    int8
	Uint    uint
	Uint64  uint64
	Uint32  uint32
	Uint16  uint16
	Uint8   uint8
}

type pointerTestStruct struct {
	String  *string
	Bool    *bool
	Float64 *float64
	Float32 *float32
	Int     *int
	Int64   *int64
	Int32   *int32
	Int16   *int16
	Int8    *int8
	Uint    *uint
	Uint64  *uint64
	Uint32  *uint32
	Uint16  *uint16
	Uint8   *uint8
}

type namedStruct struct {
	Int int `query:"sample-name"`
}

type defaultedStruct struct {
	Int  int   `default:"42"`
	Ints []int `default:"42,43"`
}

type customStruct struct {
	Int int
}

func (s *customStruct) DecodeQuery(q url.Values) error {
	i, err := strconv.Atoi(q.Get("custom"))
	s.Int = i
	return err
}

type customStructFieldStruct struct {
	Custom customStruct
}

type pointedCustomStructFieldStruct struct {
	Custom *customStruct
}

type nonImplementingStruct struct {
	Int testing.T
}

type nonExportedStruct struct {
	int int `query:"int"`
}

type slicesStruct struct {
	Strings  []string
	Bools    []bool
	Float64s []float64
	Float32s []float32
	Ints     []int
	Int64s   []int64
	Int32s   []int32
	Int16s   []int16
	Int8s    []int8
	Uints    []uint
	Uint64s  []uint64
	Uint32s  []uint32
	Uint16s  []uint16
	Uint8s   []uint8
}

type slicesDefaultedStruct struct {
	Strings  []string  `default:"hello,world"`
	Bools    []bool    `default:"true,false"`
	Float64s []float64 `default:"1.2,4.5"`
	Float32s []float32 `default:"1.2,4.5"`
	Ints     []int     `default:"12,42"`
	Int64s   []int64   `default:"12,42"`
	Int32s   []int32   `default:"12,42"`
	Int16s   []int16   `default:"12,42"`
	Int8s    []int8    `default:"12,42"`
	Uints    []uint    `default:"12,42"`
	Uint64s  []uint64  `default:"12,42"`
	Uint32s  []uint32  `default:"12,42"`
	Uint16s  []uint16  `default:"12,42"`
	Uint8s   []uint8   `default:"12,42"`
}

type customDecoderType string

func (s *customDecoderType) DecodeQuery(_ url.Values) error {
	*s = "called"
	return nil
}

type customValueStruct struct {
	Int int
}

func (s customValueStruct) DecodeQuery(_ url.Values) error {
	return assert.AnError
}

type fieldCustomValueType struct {
	Custom customValueStruct
}

type pointerCustomValueType struct {
	Custom *customValueStruct
}

type interfaceDecoderFieldStruct struct {
	Decoder Decoder
}

type namedStringSliceStruct struct {
	Tags []namedString
}

type namedString string

type commaDefaultStruct struct {
	String  string   `default:"Hello, World"`
	Pointer *string  `default:"Hello, World"`
	Strings []string `default:"hello,world"`
}

var (
	testTime      = time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	otherTestTime = time.Date(2025, 6, 7, 8, 9, 10, 0, time.UTC)
)

type timeStruct struct {
	Time    time.Time
	Pointer *time.Time
	Times   []time.Time
}

type timeDefaultStruct struct {
	Time time.Time `default:"2024-01-02T03:04:05Z"`
}

type level int

const (
	levelLow  level = 1
	levelHigh level = 2
)

func (l *level) UnmarshalText(text []byte) error {
	switch string(text) {
	case "low":
		*l = levelLow
	case "high":
		*l = levelHigh
	default:
		return fmt.Errorf("invalid level: %s", text)
	}

	return nil
}

type levelStruct struct {
	Level   level
	Pointer *level
	Levels  []level
}

type durationStruct struct {
	Duration  time.Duration
	Pointer   *time.Duration
	Durations []time.Duration
}

type durationDefaultStruct struct {
	Duration time.Duration `default:"5s"`
}

type emptyStringStruct struct {
	String  string   `default:"hello"`
	Strings []string `default:"a,b"`
}

func intOverflow() string {
	if strconv.IntSize == 32 {
		return "2147483648"
	}

	return "9223372036854775808"
}

func uintOverflow() string {
	if strconv.IntSize == 32 {
		return "4294967296"
	}

	return "18446744073709551616"
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name        string
		query       url.Values
		obj         any
		expectedErr bool
		expectedObj any
		verify      func(t *testing.T, query url.Values, obj any)
	}{
		{
			name:        "nil query",
			obj:         &testStruct{},
			expectedObj: &testStruct{},
		},
		{
			name:        "no object",
			query:       url.Values{},
			obj:         42,
			expectedErr: true,
		},
		{
			name:        "nil",
			query:       url.Values{},
			obj:         nil,
			expectedErr: true,
		},
		{
			name:        "nil pointer to struct",
			query:       url.Values{},
			obj:         (*testing.T)(nil),
			expectedErr: true,
		},
		{
			name:        "nil pointer to custom struct",
			query:       url.Values{},
			obj:         (*customDecoderType)(nil),
			expectedErr: true,
		},
		{
			name:        "pointer to string",
			query:       url.Values{},
			obj:         toPointer("hello world"),
			expectedErr: true,
		},
		{
			name: "sample objects",
			query: map[string][]string{
				"string":  {"hello world"},
				"bool":    {"true"},
				"float64": {"4.4"},
				"float32": {"4.4"},
				"int":     {"123"},
				"int64":   {"123"},
				"int32":   {"123"},
				"int16":   {"123"},
				"int8":    {"123"},
				"uint":    {"123"},
				"uint64":  {"123"},
				"uint32":  {"123"},
				"uint16":  {"123"},
				"uint8":   {"123"},
			},
			obj: &testStruct{},
			expectedObj: &testStruct{
				String:  "hello world",
				Bool:    true,
				Float64: 4.4,
				Float32: 4.4,
				Int:     123,
				Int64:   123,
				Int32:   123,
				Int16:   123,
				Int8:    123,
				Uint:    123,
				Uint64:  123,
				Uint32:  123,
				Uint16:  123,
				Uint8:   123,
			},
		},
		{
			name:        "zero values",
			query:       url.Values{},
			obj:         &testStruct{},
			expectedObj: &testStruct{},
		},
		{
			name: "sample objects for pointer",
			query: map[string][]string{
				"string":  {"hello world"},
				"bool":    {"true"},
				"float64": {"4.4"},
				"float32": {"4.4"},
				"int":     {"123"},
				"int64":   {"123"},
				"int32":   {"123"},
				"int16":   {"123"},
				"int8":    {"123"},
				"uint":    {"123"},
				"uint64":  {"123"},
				"uint32":  {"123"},
				"uint16":  {"123"},
				"uint8":   {"123"},
			},
			obj: &pointerTestStruct{},
			expectedObj: &pointerTestStruct{
				String:  toPointer("hello world"),
				Bool:    toPointer(true),
				Float64: toPointer(4.4),
				Float32: toPointer(float32(4.4)),
				Int:     toPointer(123),
				Int64:   toPointer(int64(123)),
				Int32:   toPointer(int32(123)),
				Int16:   toPointer(int16(123)),
				Int8:    toPointer(int8(123)),
				Uint:    toPointer(uint(123)),
				Uint64:  toPointer(uint64(123)),
				Uint32:  toPointer(uint32(123)),
				Uint16:  toPointer(uint16(123)),
				Uint8:   toPointer(uint8(123)),
			},
		},
		{
			name:        "zero values for pointers",
			query:       url.Values{},
			obj:         &pointerTestStruct{},
			expectedObj: &pointerTestStruct{},
		},
		{
			name: "name tag",
			query: map[string][]string{
				"sample-name": {"42"},
			},
			obj: &namedStruct{},
			expectedObj: &namedStruct{
				Int: 42,
			},
		},
		{
			name:  "default values",
			query: url.Values{},
			obj:   &defaultedStruct{},
			expectedObj: &defaultedStruct{
				Int:  42,
				Ints: []int{42, 43},
			},
		},
		{
			name: "custom parse logic success",
			query: map[string][]string{
				"custom": {"42"},
			},
			obj: &customStruct{},
			expectedObj: &customStruct{
				Int: 42,
			},
		},
		{
			name: "custom parse logic fails",
			query: map[string][]string{
				"custom": {"invalid"},
			},
			obj:         &customStruct{},
			expectedErr: true,
		},
		{
			name: "custom struct field",
			query: map[string][]string{
				"custom": {"42"},
			},
			obj: &customStructFieldStruct{},
			expectedObj: &customStructFieldStruct{
				Custom: customStruct{
					Int: 42,
				},
			},
		},
		{
			name: "custom struct field fails",
			query: map[string][]string{
				"custom": {"invalid"},
			},
			obj:         &customStructFieldStruct{},
			expectedErr: true,
		},
		{
			name: "custom struct field as pointer",
			query: map[string][]string{
				"custom": {"42"},
			},
			obj: &pointedCustomStructFieldStruct{},
			expectedObj: &pointedCustomStructFieldStruct{
				Custom: &customStruct{
					Int: 42,
				},
			},
		},
		{
			name: "custom struct field as pointer fails",
			query: map[string][]string{
				"int": {"invalid"},
			},
			obj:         &pointedCustomStructFieldStruct{},
			expectedErr: true,
		},
		{
			name:        "struct field that does not implement interface",
			query:       url.Values{},
			obj:         &nonImplementingStruct{},
			expectedObj: &nonImplementingStruct{},
		},
		{
			name: "ignore non exported fields",
			query: map[string][]string{
				"int": {"42"},
			},
			obj:         &nonExportedStruct{},
			expectedObj: &nonExportedStruct{},
		},
		{
			name: "invalid bool",
			query: map[string][]string{
				"bool": {"invalid"},
			},
			obj:         &testStruct{},
			expectedErr: true,
		},
		{
			name: "invalid float64",
			query: map[string][]string{
				"float64": {"invalid"},
			},
			obj:         &testStruct{},
			expectedErr: true,
		},
		{
			name: "invalid float32",
			query: map[string][]string{
				"float32": {strconv.FormatFloat(math.MaxFloat64, 'f', 2, 64)},
			},
			obj:         &testStruct{},
			expectedErr: true,
		},
		{
			name: "invalid int",
			query: map[string][]string{
				"int": {"9223372036854775808"},
			},
			obj:         &testStruct{},
			expectedErr: true,
		},
		{
			name: "invalid int64",
			query: map[string][]string{
				"int64": {"9223372036854775808"},
			},
			obj:         &testStruct{},
			expectedErr: true,
		},
		{
			name: "invalid int32",
			query: map[string][]string{
				"int32": {"-2147483649"},
			},
			obj:         &testStruct{},
			expectedErr: true,
		},
		{
			name: "invalid int16",
			query: map[string][]string{
				"int16": {"-32769"},
			},
			obj:         &testStruct{},
			expectedErr: true,
		},
		{
			name: "invalid int8",
			query: map[string][]string{
				"int8": {"-129"},
			},
			obj:         &testStruct{},
			expectedErr: true,
		},
		{
			name: "invalid uint",
			query: map[string][]string{
				"uint": {"18446744073709551616"},
			},
			obj:         &testStruct{},
			expectedErr: true,
		},
		{
			name: "invalid uint64",
			query: map[string][]string{
				"uint64": {"18446744073709551616"},
			},
			obj:         &testStruct{},
			expectedErr: true,
		},
		{
			name: "invalid uint32",
			query: map[string][]string{
				"uint32": {"4294967296"},
			},
			obj:         &testStruct{},
			expectedErr: true,
		},
		{
			name: "invalid uint16",
			query: map[string][]string{
				"uint16": {"65536"},
			},
			obj:         &testStruct{},
			expectedErr: true,
		},
		{
			name: "invalid uint8",
			query: map[string][]string{
				"uint8": {"256"},
			},
			obj:         &testStruct{},
			expectedErr: true,
		},
		{
			name: "slices",
			query: map[string][]string{
				"strings":  {"hello", "world"},
				"bools":    {"true", "false"},
				"float64s": {"1.2", "4.5"},
				"float32s": {"1.2", "4.5"},
				"ints":     {"12", "42"},
				"int64s":   {"12", "42"},
				"int32s":   {"12", "42"},
				"int16s":   {"12", "42"},
				"int8s":    {"12", "42"},
				"uints":    {"12", "42"},
				"uint64s":  {"12", "42"},
				"uint32s":  {"12", "42"},
				"uint16s":  {"12", "42"},
				"uint8s":   {"12", "42"},
			},
			obj: &slicesStruct{},
			expectedObj: &slicesStruct{
				Strings:  []string{"hello", "world"},
				Bools:    []bool{true, false},
				Float64s: []float64{1.2, 4.5},
				Float32s: []float32{1.2, 4.5},
				Ints:     []int{12, 42},
				Int64s:   []int64{12, 42},
				Int32s:   []int32{12, 42},
				Int16s:   []int16{12, 42},
				Int8s:    []int8{12, 42},
				Uints:    []uint{12, 42},
				Uint64s:  []uint64{12, 42},
				Uint32s:  []uint32{12, 42},
				Uint16s:  []uint16{12, 42},
				Uint8s:   []uint8{12, 42},
			},
		},
		{
			name:  "slices defaults",
			query: url.Values{},
			obj:   &slicesDefaultedStruct{},
			expectedObj: &slicesDefaultedStruct{
				Strings:  []string{"hello", "world"},
				Bools:    []bool{true, false},
				Float64s: []float64{1.2, 4.5},
				Float32s: []float32{1.2, 4.5},
				Ints:     []int{12, 42},
				Int64s:   []int64{12, 42},
				Int32s:   []int32{12, 42},
				Int16s:   []int16{12, 42},
				Int8s:    []int8{12, 42},
				Uints:    []uint{12, 42},
				Uint64s:  []uint64{12, 42},
				Uint32s:  []uint32{12, 42},
				Uint16s:  []uint16{12, 42},
				Uint8s:   []uint8{12, 42},
			},
		},
		{
			name: "slices invalid bools",
			query: map[string][]string{
				"bools": {"true", "invalid"},
			},
			obj:         &slicesStruct{},
			expectedErr: true,
		},
		{
			name: "slices invalid float64s",
			query: map[string][]string{
				"float64s": {"1.2", "invalid"},
			},
			obj:         &slicesStruct{},
			expectedErr: true,
		},
		{
			name: "slices invalid float32s",
			query: map[string][]string{
				"float32s": {"1.2", "invalid"},
			},
			obj:         &slicesStruct{},
			expectedErr: true,
		},
		{
			name: "slices invalid ints",
			query: map[string][]string{
				"ints": {"42", "9223372036854775808"},
			},
			obj:         &slicesStruct{},
			expectedErr: true,
		},
		{
			name: "slices invalid int64s",
			query: map[string][]string{
				"int64s": {"42", "9223372036854775808"},
			},
			obj:         &slicesStruct{},
			expectedErr: true,
		},
		{
			name: "slices invalid int32s",
			query: map[string][]string{
				"int32s": {"42", "2147483648"},
			},
			obj:         &slicesStruct{},
			expectedErr: true,
		},
		{
			name: "slices invalid int16s",
			query: map[string][]string{
				"int16s": {"42", "32768"},
			},
			obj:         &slicesStruct{},
			expectedErr: true,
		},
		{
			name: "slices invalid int8s",
			query: map[string][]string{
				"int8s": {"42", "128"},
			},
			obj:         &slicesStruct{},
			expectedErr: true,
		},
		{
			name: "slices invalid uints",
			query: map[string][]string{
				"uints": {"42", "18446744073709551616"},
			},
			obj:         &slicesStruct{},
			expectedErr: true,
		},
		{
			name: "slices invalid uint64s",
			query: map[string][]string{
				"uint64s": {"42", "18446744073709551616"},
			},
			obj:         &slicesStruct{},
			expectedErr: true,
		},
		{
			name: "slices invalid uint32s",
			query: map[string][]string{
				"uint32s": {"42", "4294967296"},
			},
			obj:         &slicesStruct{},
			expectedErr: true,
		},
		{
			name: "slices invalid uint16s",
			query: map[string][]string{
				"uint16s": {"42", "65536"},
			},
			obj:         &slicesStruct{},
			expectedErr: true,
		},
		{
			name: "slices invalid uint8s",
			query: map[string][]string{
				"uint8s": {"42", "256"},
			},
			obj:         &slicesStruct{},
			expectedErr: true,
		},
		{
			name:        "custom type with Decode interface",
			query:       map[string][]string{},
			obj:         toPointer(customDecoderType("")),
			expectedObj: toPointer(customDecoderType("called")),
		},
		{
			name: "value receiver cannot decode",
			query: map[string][]string{
				"custom": {"42"},
			},
			obj:         &customValueStruct{},
			expectedErr: true,
		},
		{
			name: "value receiver as field cannot decode",
			query: map[string][]string{
				"custom": {"42"},
			},
			obj:         &fieldCustomValueType{},
			expectedErr: true,
		},
		{
			name: "value receiver as field pointer cannot decode",
			query: map[string][]string{
				"custom": {"42"},
			},
			obj:         &pointerCustomValueType{},
			expectedErr: true,
		},
		{
			name:        "nil interface field is ignored",
			query:       url.Values{},
			obj:         &interfaceDecoderFieldStruct{},
			expectedObj: &interfaceDecoderFieldStruct{},
		},
		{
			name: "interface field with custom decoder",
			query: map[string][]string{
				"custom": {"42"},
			},
			obj: &interfaceDecoderFieldStruct{
				Decoder: &customStruct{},
			},
			expectedObj: &interfaceDecoderFieldStruct{
				Decoder: &customStruct{
					Int: 42,
				},
			},
		},
		{
			name: "slice of named string type",
			query: map[string][]string{
				"tags": {"hello", "world"},
			},
			obj: &namedStringSliceStruct{},
			expectedObj: &namedStringSliceStruct{
				Tags: []namedString{"hello", "world"},
			},
		},
		{
			name: "string slice does not alias query",
			query: map[string][]string{
				"strings": {"hello", "world"},
			},
			obj: &slicesStruct{},
			expectedObj: &slicesStruct{
				Strings: []string{"hello", "world"},
			},
			verify: func(t *testing.T, query url.Values, obj any) {
				obj.(*slicesStruct).Strings[0] = "mutated"

				assert.Equal(t, []string{"hello", "world"}, query["strings"])
			},
		},
		{
			name:  "default values with comma only split for slices",
			query: url.Values{},
			obj:   &commaDefaultStruct{},
			expectedObj: &commaDefaultStruct{
				String:  "Hello, World",
				Pointer: toPointer("Hello, World"),
				Strings: []string{"hello", "world"},
			},
		},
		{
			name: "int out of range for platform",
			query: map[string][]string{
				"int": {intOverflow()},
			},
			obj:         &testStruct{},
			expectedErr: true,
		},
		{
			name: "uint out of range for platform",
			query: map[string][]string{
				"uint": {uintOverflow()},
			},
			obj:         &testStruct{},
			expectedErr: true,
		},
		{
			name: "slices int out of range for platform",
			query: map[string][]string{
				"ints": {"42", intOverflow()},
			},
			obj:         &slicesStruct{},
			expectedErr: true,
		},
		{
			name: "slices uint out of range for platform",
			query: map[string][]string{
				"uints": {"42", uintOverflow()},
			},
			obj:         &slicesStruct{},
			expectedErr: true,
		},
		{
			name: "text unmarshaler",
			query: map[string][]string{
				"time":    {"2024-01-02T03:04:05Z"},
				"pointer": {"2024-01-02T03:04:05Z"},
				"times":   {"2024-01-02T03:04:05Z", "2025-06-07T08:09:10Z"},
			},
			obj: &timeStruct{},
			expectedObj: &timeStruct{
				Time:    testTime,
				Pointer: toPointer(testTime),
				Times:   []time.Time{testTime, otherTestTime},
			},
		},
		{
			name:  "text unmarshaler default",
			query: url.Values{},
			obj:   &timeDefaultStruct{},
			expectedObj: &timeDefaultStruct{
				Time: testTime,
			},
		},
		{
			name: "text unmarshaler invalid value",
			query: map[string][]string{
				"time": {"invalid"},
			},
			obj:         &timeStruct{},
			expectedErr: true,
		},
		{
			name: "text unmarshaler invalid slice value",
			query: map[string][]string{
				"times": {"2024-01-02T03:04:05Z", "invalid"},
			},
			obj:         &timeStruct{},
			expectedErr: true,
		},
		{
			name: "text unmarshaler takes precedence over kind",
			query: map[string][]string{
				"level":   {"high"},
				"pointer": {"low"},
				"levels":  {"low", "high"},
			},
			obj: &levelStruct{},
			expectedObj: &levelStruct{
				Level:   levelHigh,
				Pointer: toPointer(levelLow),
				Levels:  []level{levelLow, levelHigh},
			},
		},
		{
			name: "text unmarshaler on kind invalid value",
			query: map[string][]string{
				"level": {"1"},
			},
			obj:         &levelStruct{},
			expectedErr: true,
		},
		{
			name: "empty values of non string fields are not provided",
			query: map[string][]string{
				"bool":    {""},
				"float64": {""},
				"int":     {""},
				"uint":    {""},
			},
			obj:         &testStruct{},
			expectedObj: &testStruct{},
		},
		{
			name: "empty values of pointer fields are not provided",
			query: map[string][]string{
				"int": {""},
			},
			obj:         &pointerTestStruct{},
			expectedObj: &pointerTestStruct{},
		},
		{
			name: "empty values use defaults",
			query: map[string][]string{
				"int":  {""},
				"ints": {""},
			},
			obj: &defaultedStruct{},
			expectedObj: &defaultedStruct{
				Int:  42,
				Ints: []int{42, 43},
			},
		},
		{
			name: "empty values of slices are dropped",
			query: map[string][]string{
				"ints": {"1", "", "2"},
			},
			obj: &slicesStruct{},
			expectedObj: &slicesStruct{
				Ints: []int{1, 2},
			},
		},
		{
			name: "first non empty value is used for scalars",
			query: map[string][]string{
				"int": {"", "5"},
			},
			obj: &testStruct{},
			expectedObj: &testStruct{
				Int: 5,
			},
		},
		{
			name: "empty values are kept for strings",
			query: map[string][]string{
				"string":  {""},
				"strings": {"", "a"},
			},
			obj: &emptyStringStruct{},
			expectedObj: &emptyStringStruct{
				String:  "",
				Strings: []string{"", "a"},
			},
		},
		{
			name: "empty value of text unmarshaler is not provided",
			query: map[string][]string{
				"time": {""},
			},
			obj: &timeDefaultStruct{},
			expectedObj: &timeDefaultStruct{
				Time: testTime,
			},
		},
		{
			name: "duration",
			query: map[string][]string{
				"duration":  {"1h30m"},
				"pointer":   {"1s"},
				"durations": {"1s", "250ms"},
			},
			obj: &durationStruct{},
			expectedObj: &durationStruct{
				Duration:  90 * time.Minute,
				Pointer:   toPointer(time.Second),
				Durations: []time.Duration{time.Second, 250 * time.Millisecond},
			},
		},
		{
			name:  "duration default",
			query: url.Values{},
			obj:   &durationDefaultStruct{},
			expectedObj: &durationDefaultStruct{
				Duration: 5 * time.Second,
			},
		},
		{
			name: "duration without unit",
			query: map[string][]string{
				"duration": {"5"},
			},
			obj:         &durationStruct{},
			expectedErr: true,
		},
		{
			name: "duration invalid slice value",
			query: map[string][]string{
				"durations": {"1s", "invalid"},
			},
			obj:         &durationStruct{},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Decode(tt.query, tt.obj)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedObj, tt.obj)

				if tt.verify != nil {
					tt.verify(t, tt.query, tt.obj)
				}
			}
		})
	}
}

func toPointer[T any](v T) *T {
	return &v
}
