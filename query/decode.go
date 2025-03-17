package query

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
)

// Decoder custom parsing logic for structs. Can be implemented on
// structs to deviate from the default logic.
type Decoder interface {
	// DecodeQuery parses the given url.Values query values and set it
	// to the implementing struct.
	DecodeQuery(q url.Values) error
}

// Decode parses the URL query parameters given in the ur.Values to the
// object passed using the name of the fields or the optional overwrite
// with the TagName. Default values can be provided via the TagDefault
// tag.
func Decode(q url.Values, obj any) error {
	if q == nil {
		return nil
	}

	return parse(q, reflect.ValueOf(obj))
}

func parse(q url.Values, val reflect.Value) error {
	// check for custom types
	if custom, err := decodeCustom(q, val); custom {
		return err
	}

	kind := val.Kind()
	switch kind {
	case reflect.Ptr:
		return parse(q, val.Elem())
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
	if len(values) == 0 {
		values = getDefaultTags(field)
	}

	return values
}

func getName(field *reflect.StructField) string {
	return getNameTags(field)[0]
}

func parseField(q url.Values, field reflect.Value, values []string) error {
	typ := field.Type()

	switch typ.Kind() {
	case reflect.String:
		field.SetString(values[0])
	case reflect.Bool:
		b, err := strconv.ParseBool(values[0])
		if err != nil {
			return err
		}
		field.SetBool(b)
	case reflect.Float64:
		f, err := strconv.ParseFloat(values[0], 64)
		if err != nil {
			return err
		}
		field.SetFloat(f)
	case reflect.Float32:
		f, err := strconv.ParseFloat(values[0], 32)
		if err != nil {
			return err
		}
		field.SetFloat(f)
	case reflect.Int, reflect.Int64:
		v, err := strconv.ParseInt(values[0], 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(v)
	case reflect.Int32:
		v, err := strconv.ParseInt(values[0], 10, 32)
		if err != nil {
			return err
		}
		field.SetInt(v)
	case reflect.Int16:
		v, err := strconv.ParseInt(values[0], 10, 16)
		if err != nil {
			return err
		}
		field.SetInt(v)
	case reflect.Int8:
		v, err := strconv.ParseInt(values[0], 10, 8)
		if err != nil {
			return err
		}
		field.SetInt(v)
	case reflect.Uint, reflect.Uint64:
		v, err := strconv.ParseUint(values[0], 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(v)
	case reflect.Uint32:
		v, err := strconv.ParseUint(values[0], 10, 32)
		if err != nil {
			return err
		}
		field.SetUint(v)
	case reflect.Uint16:
		v, err := strconv.ParseUint(values[0], 10, 16)
		if err != nil {
			return err
		}
		field.SetUint(v)
	case reflect.Uint8:
		v, err := strconv.ParseUint(values[0], 10, 8)
		if err != nil {
			return err
		}
		field.SetUint(v)
	case reflect.Slice:
		return parseSlice(field, values)
	case reflect.Ptr:
		created := reflect.New(typ.Elem())
		err := parseField(q, created.Elem(), values)
		if err != nil {
			return err
		}
		field.Set(created)
	default:
		// ignore other types
	}

	return nil
}

func parseSlice(field reflect.Value, values []string) error {
	n := len(values)

	switch field.Type().Elem().Kind() {
	case reflect.String:
		field.Set(reflect.ValueOf(values))
	case reflect.Bool:
		parsed := make([]bool, n)
		for i := 0; i < n; i++ {
			v, err := strconv.ParseBool(values[i])
			if err != nil {
				return err
			}
			parsed[i] = v
		}
		field.Set(reflect.ValueOf(parsed))
	case reflect.Float64:
		parsed := make([]float64, n)
		for i := 0; i < n; i++ {
			v, err := strconv.ParseFloat(values[i], 64)
			if err != nil {
				return err
			}
			parsed[i] = v
		}
		field.Set(reflect.ValueOf(parsed))
	case reflect.Float32:
		parsed := make([]float32, n)
		for i := 0; i < n; i++ {
			v, err := strconv.ParseFloat(values[i], 32)
			if err != nil {
				return err
			}
			parsed[i] = float32(v)
		}
		field.Set(reflect.ValueOf(parsed))
	case reflect.Int:
		parsed := make([]int, n)
		for i := 0; i < n; i++ {
			v, err := strconv.Atoi(values[i])
			if err != nil {
				return err
			}
			parsed[i] = v
		}
		field.Set(reflect.ValueOf(parsed))
	case reflect.Int64:
		parsed := make([]int64, n)
		for i := 0; i < n; i++ {
			v, err := strconv.ParseInt(values[i], 10, 64)
			if err != nil {
				return err
			}
			parsed[i] = v
		}
		field.Set(reflect.ValueOf(parsed))
	case reflect.Int32:
		parsed := make([]int32, n)
		for i := 0; i < n; i++ {
			v, err := strconv.ParseInt(values[i], 10, 32)
			if err != nil {
				return err
			}
			parsed[i] = int32(v)
		}
		field.Set(reflect.ValueOf(parsed))
	case reflect.Int16:
		parsed := make([]int16, n)
		for i := 0; i < n; i++ {
			v, err := strconv.ParseInt(values[i], 10, 16)
			if err != nil {
				return err
			}
			parsed[i] = int16(v)
		}
		field.Set(reflect.ValueOf(parsed))
	case reflect.Int8:
		parsed := make([]int8, n)
		for i := 0; i < n; i++ {
			v, err := strconv.ParseInt(values[i], 10, 8)
			if err != nil {
				return err
			}
			parsed[i] = int8(v)
		}
		field.Set(reflect.ValueOf(parsed))
	case reflect.Uint:
		parsed := make([]uint, n)
		for i := 0; i < n; i++ {
			v, err := strconv.ParseUint(values[i], 10, 64)
			if err != nil {
				return err
			}
			parsed[i] = uint(v)
		}
		field.Set(reflect.ValueOf(parsed))
	case reflect.Uint64:
		parsed := make([]uint64, n)
		for i := 0; i < n; i++ {
			v, err := strconv.ParseUint(values[i], 10, 64)
			if err != nil {
				return err
			}
			parsed[i] = v
		}
		field.Set(reflect.ValueOf(parsed))
	case reflect.Uint32:
		parsed := make([]uint32, n)
		for i := 0; i < n; i++ {
			v, err := strconv.ParseUint(values[i], 10, 32)
			if err != nil {
				return err
			}
			parsed[i] = uint32(v)
		}
		field.Set(reflect.ValueOf(parsed))
	case reflect.Uint16:
		parsed := make([]uint16, n)
		for i := 0; i < n; i++ {
			v, err := strconv.ParseUint(values[i], 10, 16)
			if err != nil {
				return err
			}
			parsed[i] = uint16(v)
		}
		field.Set(reflect.ValueOf(parsed))
	case reflect.Uint8:
		parsed := make([]uint8, n)
		for i := 0; i < n; i++ {
			v, err := strconv.ParseUint(values[i], 10, 8)
			if err != nil {
				return err
			}
			parsed[i] = uint8(v)
		}
		field.Set(reflect.ValueOf(parsed))
	default:
		// ignore other types
	}

	return nil
}

var decoderType = reflect.TypeOf(new(Decoder)).Elem()

func decodeCustom(q url.Values, val reflect.Value) (bool, error) {
	typ := val.Type()

	if !typ.Implements(decoderType) {
		if val.CanAddr() && val.Addr().Type().Implements(decoderType) {
			val = val.Addr()
		} else {
			return false, nil // ignore types that do not implement Decoder interface
		}
	}

	if !reflect.Indirect(val).IsValid() {
		created := reflect.New(typ.Elem())
		val.Set(created)
		val = created
	}

	m := val.Interface().(Decoder)
	return true, m.DecodeQuery(q)
}
