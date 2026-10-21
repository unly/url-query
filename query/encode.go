package query

import (
	"encoding"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"time"
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
	case reflect.Invalid:
		return nil
	case reflect.Pointer:
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

		key, skip := getEncodingName(&fieldType, field)
		if skip {
			continue
		}

		custom, err := encodeCustom(v, field)
		if custom {
			if err != nil {
				return err
			}
			continue
		}

		err = encodeField(v, field, key)
		if err != nil {
			return err
		}
	}

	return nil
}

func encodeField(v url.Values, field reflect.Value, key string) error {
	if !field.IsValid() {
		return nil
	}

	if ok, err := encodeText(v, key, field); ok {
		return err
	}

	if field.Type() == durationType {
		v.Add(key, encodeDuration(field))
		return nil
	}

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
	case reflect.Pointer:
		return encodeField(v, field.Elem(), key)
	case reflect.Slice:
		return encodeSlice(v, field, key)
	default:
		// ignore others
	}

	return nil
}

func encodeSlice(v url.Values, field reflect.Value, key string) error {
	elemType := field.Type().Elem()
	if isTextMarshaler(elemType) {
		return addTextSlice(v, field, key)
	}

	if elemType == durationType {
		addSlice(v, field, key, encodeDuration)
		return nil
	}

	switch elemType.Kind() {
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

	return nil
}

func addTextSlice(v url.Values, field reflect.Value, key string) error {
	n := field.Len()
	for i := 0; i < n; i++ {
		if _, err := encodeText(v, key, field.Index(i)); err != nil {
			return err
		}
	}

	return nil
}

func addSlice(v url.Values, field reflect.Value, key string, fn func(value reflect.Value) string) {
	n := field.Len()
	for i := 0; i < n; i++ {
		v.Add(key, fn(field.Index(i)))
	}
}

func encodeDuration(val reflect.Value) string {
	return time.Duration(val.Int()).String()
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

var textMarshalerType = reflect.TypeOf(new(encoding.TextMarshaler)).Elem()

func encodeText(v url.Values, key string, val reflect.Value) (bool, error) {
	if !val.IsValid() {
		return false, nil
	}

	typ := val.Type()

	if !typ.Implements(textMarshalerType) {
		if !reflect.PointerTo(typ).Implements(textMarshalerType) {
			return false, nil
		}

		ptr := reflect.New(typ)
		ptr.Elem().Set(val)
		val = ptr
	}

	if (val.Kind() == reflect.Pointer || val.Kind() == reflect.Interface) && val.IsNil() {
		return true, nil
	}

	text, err := val.Interface().(encoding.TextMarshaler).MarshalText()
	if err != nil {
		return true, err
	}

	v.Add(key, string(text))
	return true, nil
}

func isTextMarshaler(typ reflect.Type) bool {
	if typ.Kind() == reflect.Interface {
		return false
	}

	return typ.Implements(textMarshalerType) || reflect.PointerTo(typ).Implements(textMarshalerType)
}

var encoderType = reflect.TypeOf(new(Encoder)).Elem()

func encodeCustom(v url.Values, val reflect.Value) (bool, error) {
	if !val.IsValid() {
		return false, nil
	}

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

	if (val.Kind() == reflect.Pointer || val.Kind() == reflect.Interface) && val.IsNil() {
		return true, nil
	}

	m := val.Interface().(Encoder)
	sub, err := m.EncodeValues()
	for k, values := range sub {
		v[k] = append(v[k], values...)
	}

	return true, err
}
