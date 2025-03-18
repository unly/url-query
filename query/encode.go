package query

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
)

// Encoder custom encoding logic for types to allow a logic. Returned
// values will be added to the final values map.
type Encoder interface {
	// EncodeValues custom encoding of a receiver.
	EncodeValues() (url.Values, error)
}

// Encode sets the url query parameters based on the values of the
// given type. Uses either TagName or the name of the field. If the tag
// is '-' it will be excluded. There is also the option to set 'omitempty'
// to omit the encoding of zero values.
func Encode(obj any) (url.Values, error) {
	values := make(url.Values)

	return values, encode(values, reflect.ValueOf(obj))
}

func encode(v url.Values, val reflect.Value) error {
	if custom, err := encodeCustom(v, val); custom {
		return err
	}

	switch val.Kind() {
	case reflect.Ptr:
		return encode(v, val.Elem())
	case reflect.Struct:
		return encodeStruct(v, val)
	default:
		return fmt.Errorf("unsupported type: %s", val.Type())
	}
}

func encodeStruct(v url.Values, val reflect.Value) error {
	typ := val.Type()

	n := val.NumField()
	for i := 0; i < n; i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if !fieldType.IsExported() {
			continue
		}

		if custom, err := encodeCustom(v, field); custom {
			return err
		}

		key, skip := getEncodingName(&fieldType, field)
		if skip {
			continue
		}

		encodeField(v, field, key)
	}

	return nil
}

func encodeField(v url.Values, field reflect.Value, key string) {
	switch field.Kind() {
	case reflect.String:
		v.Add(key, encodeString(field))
	case reflect.Bool:
		v.Add(key, encodeBool(field))
	case reflect.Float32:
		v.Add(key, encodeFloat32(field))
	case reflect.Float64:
		v.Add(key, encodeFloat64(field))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.Add(key, encodeInt(field))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v.Add(key, encodeUint(field))
	case reflect.Ptr:
		encodeField(v, field.Elem(), key)
	case reflect.Slice:
		encodeSlice(v, field, key)
	default:
		// ignore others
	}
}

func encodeSlice(v url.Values, field reflect.Value, key string) {
	switch field.Type().Elem().Kind() {
	case reflect.String:
		addSlice(v, field, key, encodeString)
	case reflect.Bool:
		addSlice(v, field, key, encodeBool)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		addSlice(v, field, key, encodeInt)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		addSlice(v, field, key, encodeUint)
	case reflect.Float32:
		addSlice(v, field, key, encodeFloat32)
	case reflect.Float64:
		addSlice(v, field, key, encodeFloat64)
	default:
		// ignore others
	}
}

func addSlice(v url.Values, field reflect.Value, key string, fn func(value reflect.Value) string) {
	n := field.Len()
	for i := 0; i < n; i++ {
		v.Add(key, fn(field.Index(i)))
	}
}

func encodeFloat64(val reflect.Value) string {
	return strconv.FormatFloat(val.Float(), 'f', -1, 64)
}

func encodeFloat32(val reflect.Value) string {
	return strconv.FormatFloat(val.Float(), 'f', -1, 32)
}

func encodeBool(val reflect.Value) string {
	return strconv.FormatBool(val.Bool())
}

func encodeString(val reflect.Value) string {
	return val.String()
}

func encodeInt(val reflect.Value) string {
	return strconv.FormatInt(val.Int(), 10)
}

func encodeUint(val reflect.Value) string {
	return strconv.FormatUint(val.Uint(), 10)
}

func getEncodingName(field *reflect.StructField, val reflect.Value) (string, bool) {
	names := getNameTags(field)
	if names[0] == "-" {
		return "", true
	}

	if len(names) > 1 && names[1] == "omitempty" && val.IsZero() {
		return "", true
	}

	return names[0], false
}

var encoderType = reflect.TypeOf(new(Encoder)).Elem()

func encodeCustom(v url.Values, val reflect.Value) (bool, error) {
	typ := val.Type()

	if !typ.Implements(encoderType) {
		if reflect.PointerTo(typ).Implements(encoderType) {
			newValue := reflect.New(typ).Elem()
			newValue.Set(val)
			val = newValue.Addr()
		} else {
			return false, nil // ignore types that do not implement Encoder interface
		}
	}

	m := val.Interface().(Encoder)
	sub, err := m.EncodeValues()
	for k, values := range sub {
		v[k] = append(v[k], values...)
	}

	return true, err
}
