package query

import (
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
		v.Add(key, field.String())
	case reflect.Bool:
		v.Add(key, strconv.FormatBool(field.Bool()))
	case reflect.Float32:
		v.Add(key, strconv.FormatFloat(field.Float(), 'f', -1, 32))
	case reflect.Float64:
		v.Add(key, strconv.FormatFloat(field.Float(), 'f', -1, 64))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.Add(key, strconv.FormatInt(field.Int(), 10))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v.Add(key, strconv.FormatUint(field.Uint(), 10))
	case reflect.Ptr:
		encodeField(v, field.Elem(), key)
	case reflect.Slice:
		encodeSlice(v, &field, key)
	default:
		// ignore others
	}
}

func encodeSlice(v url.Values, field *reflect.Value, key string) {
	switch field.Type().Elem().Kind() {
	case reflect.String:
		n := field.Len()
		for i := 0; i < n; i++ {
			val := field.Index(i)
			v.Add(key, val.String())
		}
	case reflect.Bool:
		n := field.Len()
		for i := 0; i < n; i++ {
			val := field.Index(i)
			v.Add(key, strconv.FormatBool(val.Bool()))
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n := field.Len()
		for i := 0; i < n; i++ {
			val := field.Index(i)
			v.Add(key, strconv.FormatInt(val.Int(), 10))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n := field.Len()
		for i := 0; i < n; i++ {
			val := field.Index(i)
			v.Add(key, strconv.FormatUint(val.Uint(), 10))
		}
	case reflect.Float32:
		n := field.Len()
		for i := 0; i < n; i++ {
			val := field.Index(i)
			v.Add(key, strconv.FormatFloat(val.Float(), 'f', -1, 32))
		}
	case reflect.Float64:
		n := field.Len()
		for i := 0; i < n; i++ {
			val := field.Index(i)
			v.Add(key, strconv.FormatFloat(val.Float(), 'f', -1, 64))
		}
	default:
		// ignore others
	}
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
			if val.CanAddr() {
				val = val.Addr()
			} else {
				newValue := reflect.New(typ).Elem() // Create a new, addressable value of the same type
				newValue.Set(val)                   // Copy the unaddressable value into the new value
				val = newValue.Addr()
			}
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
