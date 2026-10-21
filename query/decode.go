package query

import (
	"encoding"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"time"
)

// Decoder custom parsing logic for structs. Can be implemented on
// structs to deviate from the default logic.
type Decoder interface {
	// DecodeQuery parses the given url.Values query values and set it
	// to the implementing struct.
	DecodeQuery(q url.Values) error
}

// Decode parses the URL query parameters given in the url.Values to the
// object passed using the name of the fields or the optional overwrite
// with the TagName. Default values can be provided via the TagDefault
// tag.
//
// Fields whose type implements encoding.TextUnmarshaler, like time.Time,
// are parsed using UnmarshalText instead of the default logic for their
// kind. Fields of type time.Duration are parsed using time.ParseDuration.
//
// Empty values are treated as not provided for all types except strings,
// so default values are used and empty values of slices are dropped.
//
// Floats are parsed with strconv.ParseFloat, so the special values "NaN",
// "Inf" and "-Inf" are accepted. Validate the result if these are not
// acceptable for the use case.
func Decode(q url.Values, obj any) error {
	return parse(q, reflect.ValueOf(obj))
}

func parse(q url.Values, val reflect.Value) error {
	if val.Kind() != reflect.Pointer || val.IsNil() {
		return errors.New("obj must be a non-nil pointer")
	}

	// check for custom types
	if custom, err := decodeCustom(q, val); custom {
		return err
	}

	val = val.Elem()
	kind := val.Kind()
	switch kind {
	case reflect.Struct:
		return parseStruct(q, val)
	default:
		return fmt.Errorf("unsupported type: %s", kind)
	}
}

func parseStruct(q url.Values, val reflect.Value) error {
	typ := val.Type()

	var errs []error
	n := typ.NumField()
	for i := 0; i < n; i++ {
		fieldType := typ.Field(i)
		if !fieldType.IsExported() {
			continue
		}

		field := val.Field(i)
		if !field.CanAddr() || !field.CanSet() {
			continue
		}

		// check if custom decoder and run it
		if custom, err := decodeCustom(q, field); custom {
			if err != nil {
				errs = append(errs, err)
			}
			continue
		}

		values := getValues(q, &fieldType)
		if len(values) == 0 {
			continue // skip empty values
		}

		fieldErr := parseField(q, field, values)
		if fieldErr != nil {
			errs = append(errs, fieldErr)
		}
	}

	return errors.Join(errs...)
}

func getValues(q url.Values, field *reflect.StructField) []string {
	values := q[getName(field)]
	if !keepsEmptyValues(field.Type) {
		values = withoutEmpty(values)
	}

	if len(values) == 0 {
		values = getDefaultTags(field)
	}

	return values
}

func keepsEmptyValues(typ reflect.Type) bool {
	for typ.Kind() == reflect.Pointer || typ.Kind() == reflect.Slice {
		typ = typ.Elem()
	}

	return typ.Kind() == reflect.String
}

func withoutEmpty(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value != "" {
			result = append(result, value)
		}
	}

	return result
}

func getName(field *reflect.StructField) string {
	return getNameTags(field)[0]
}

func parseField(q url.Values, field reflect.Value, values []string) error {
	typ := field.Type()

	if ok, err := parseText(field, values[0]); ok {
		return err
	}

	switch typ.Kind() {
	case reflect.String:
		field.SetString(values[0])
		return nil
	case reflect.Bool:
		return setField(strconv.ParseBool, field.SetBool, values[0])
	case reflect.Float64:
		return setField(parseFloat64, field.SetFloat, values[0])
	case reflect.Float32:
		return setField(parseFloat32, field.SetFloat, values[0])
	case reflect.Int:
		return setField(parseInt, field.SetInt, values[0])
	case reflect.Int64:
		if typ == durationType {
			return setField(parseDuration, field.SetInt, values[0])
		}
		return setField(parseInt64, field.SetInt, values[0])
	case reflect.Int32:
		return setField(parseInt32, field.SetInt, values[0])
	case reflect.Int16:
		return setField(parseInt16, field.SetInt, values[0])
	case reflect.Int8:
		return setField(parseInt8, field.SetInt, values[0])
	case reflect.Uint:
		return setField(parseUint, field.SetUint, values[0])
	case reflect.Uint64:
		return setField(parseUint64, field.SetUint, values[0])
	case reflect.Uint32:
		return setField(parseUint32, field.SetUint, values[0])
	case reflect.Uint16:
		return setField(parseUint16, field.SetUint, values[0])
	case reflect.Uint8:
		return setField(parseUint8, field.SetUint, values[0])
	case reflect.Slice:
		return parseSlice(field, values)
	case reflect.Pointer:
		created := reflect.New(typ.Elem())
		field.Set(created)
		return parseField(q, created.Elem(), values)
	default:
		// ignore other types
		return nil
	}
}

func parseSlice(field reflect.Value, values []string) error {
	elemType := field.Type().Elem()
	if isTextUnmarshaler(elemType) {
		return setTextSlice(field, values)
	}

	switch elemType.Kind() {
	case reflect.String:
		return setSlice(parseString, field, values)
	case reflect.Bool:
		return setSlice(strconv.ParseBool, field, values)
	case reflect.Float64:
		return setSlice(parseFloat64, field, values)
	case reflect.Float32:
		return setSlice(parseFloat32, field, values)
	case reflect.Int:
		return setSlice(strconv.Atoi, field, values)
	case reflect.Int64:
		if elemType == durationType {
			return setSlice(parseDuration, field, values)
		}
		return setSlice(parseInt64, field, values)
	case reflect.Int32:
		return setSlice(parseInt32, field, values)
	case reflect.Int16:
		return setSlice(parseInt16, field, values)
	case reflect.Int8:
		return setSlice(parseInt8, field, values)
	case reflect.Uint:
		return setSlice(parseUint, field, values)
	case reflect.Uint64:
		return setSlice(parseUint64, field, values)
	case reflect.Uint32:
		return setSlice(parseUint32, field, values)
	case reflect.Uint16:
		return setSlice(parseUint16, field, values)
	case reflect.Uint8:
		return setSlice(parseUint8, field, values)
	default:
		// ignore other types
		return nil
	}
}

func parseString(s string) (string, error) {
	return s, nil
}

func parseFloat64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func parseFloat32(s string) (float64, error) {
	return strconv.ParseFloat(s, 32)
}

var durationType = reflect.TypeOf(time.Duration(0))

func parseDuration(s string) (int64, error) {
	d, err := time.ParseDuration(s)
	return int64(d), err
}

func parseInt(s string) (int64, error) {
	return strconv.ParseInt(s, 10, strconv.IntSize)
}

func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func parseInt32(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 32)
}

func parseInt16(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 16)
}

func parseInt8(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 8)
}

func parseUint(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, strconv.IntSize)
}

func parseUint64(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}

func parseUint32(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 32)
}

func parseUint16(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 16)
}

func parseUint8(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 8)
}

func setField[T any](fn func(s string) (T, error), set func(T), value string) error {
	v, err := fn(value)
	if err != nil {
		return err
	}

	set(v)
	return nil
}

func setSlice[T any](fn func(s string) (T, error), field reflect.Value, values []string) error {
	n := len(values)
	elemType := field.Type().Elem()
	sliceVal := reflect.MakeSlice(field.Type(), n, n)

	for i := 0; i < n; i++ {
		v, err := fn(values[i])
		if err != nil {
			return err
		}

		rv := reflect.ValueOf(v)
		if !rv.Type().AssignableTo(elemType) {
			rv = rv.Convert(elemType)
		}
		sliceVal.Index(i).Set(rv)
	}

	field.Set(sliceVal)
	return nil
}

var textUnmarshalerType = reflect.TypeOf(new(encoding.TextUnmarshaler)).Elem()

func parseText(field reflect.Value, value string) (bool, error) {
	if !field.CanAddr() || !field.Addr().Type().Implements(textUnmarshalerType) {
		return false, nil
	}

	u := field.Addr().Interface().(encoding.TextUnmarshaler)
	return true, u.UnmarshalText([]byte(value))
}

func isTextUnmarshaler(typ reflect.Type) bool {
	if typ.Kind() == reflect.Interface {
		return false
	}

	return typ.Implements(textUnmarshalerType) || reflect.PointerTo(typ).Implements(textUnmarshalerType)
}

func setTextSlice(field reflect.Value, values []string) error {
	n := len(values)
	sliceVal := reflect.MakeSlice(field.Type(), n, n)

	for i := 0; i < n; i++ {
		if err := parseField(nil, sliceVal.Index(i), values[i:i+1]); err != nil {
			return err
		}
	}

	field.Set(sliceVal)
	return nil
}

var decoderType = reflect.TypeOf(new(Decoder)).Elem()

func decodeCustom(q url.Values, val reflect.Value) (bool, error) {
	if !val.IsValid() {
		return false, nil
	}

	typ := val.Type()

	if !typ.Implements(decoderType) {
		if reflect.PointerTo(typ).Implements(decoderType) && val.CanAddr() {
			val = val.Addr()
		} else {
			return false, nil // ignore types that do not implement Decoder interface
		}
	}

	if val.Kind() == reflect.Interface && val.IsNil() {
		return true, nil
	}

	if !reflect.Indirect(val).IsValid() {
		created := reflect.New(typ.Elem())
		val.Set(created)
		val = created
	}

	m := val.Interface().(Decoder)
	return true, m.DecodeQuery(q)
}
