package query

import (
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

const (
	TagName    = "query"
	TagDefault = "default"
)

// Parser custom parsing logic for structs. Can be implemented on
// self created structs to deviate the default logic.
type Parser interface {
	// ParseQuery parse given url.Values query values and set it to the
	// implementing struct.
	ParseQuery(q url.Values) error
}

// Parse parses the URL query parameters given in the ur.Values to the
// object passed using the name of the fields or the optional overwrite
// with the TagName. Default values can be provided via the TagDefault
// tag.
func Parse(q url.Values, obj any) error {
	if q == nil {
		return nil
	}

	return parse(q, obj)
}

// ParseRequest uses the url of the given request for Parse.
func ParseRequest(r *http.Request, obj any) error {
	if r == nil {
		return errors.New("nil Request")
	}

	return Parse(r.URL.Query(), obj)
}

func parse(q url.Values, obj any) error {
	if p, ok := obj.(Parser); ok {
		return p.ParseQuery(q)
	}

	val := reflect.ValueOf(obj)
	if val.Kind() != reflect.Ptr {
		return errors.New("obj must be a pointer to a struct")
	}

	val = val.Elem()
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

		values, ok := getValues(&fieldType, q)
		if !ok {
			continue // skip empty values
		}

		err := parseField(&field, q, values)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func getValues(field *reflect.StructField, q url.Values) ([]string, bool) {
	if field.Type.Kind() == reflect.Struct ||
		(field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct) {
		return nil, true // for custom structs ignore the values
	}

	values, ok := q[getName(field)]
	if !ok || len(values) == 0 {
		values = getDefaultValues(field)
		if len(values) == 0 {
			return nil, false
		}
	}

	return values, true
}

func getName(field *reflect.StructField) string {
	name, ok := field.Tag.Lookup(TagName)
	if !ok {
		fieldName := []rune(field.Name)
		fieldName[0] = unicode.ToLower(fieldName[0])
		name = string(fieldName)
	}

	return name
}

func getDefaultValues(field *reflect.StructField) []string {
	value, ok := field.Tag.Lookup(TagDefault)
	if !ok {
		return nil
	}

	return strings.Split(value, ",")
}

func parseField(field *reflect.Value, q url.Values, values []string) error {
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
		elem := created.Elem()
		err := parseField(&elem, q, values)
		if err != nil {
			return err
		}
		field.Set(created)
	case reflect.Struct:
		return parseStruct(field, q)
	default:
		// ignore other types
	}

	return nil
}

func parseSlice(field *reflect.Value, values []string) error {
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

func parseStruct(field *reflect.Value, q url.Values) error {
	ptrReceiver := field.Addr()
	typ := ptrReceiver.Type()

	if !typ.Implements(reflect.TypeOf((*Parser)(nil)).Elem()) {
		return nil // ignore structs that do not implement Parser interface
	}

	method, ok := typ.MethodByName("ParseQuery")
	if !ok {
		return errors.New("method ParseQuery not found")
	}

	returns := method.Func.Call([]reflect.Value{ptrReceiver, reflect.ValueOf(q)})
	if len(returns) == 0 {
		return errors.New("calling method ParseQuery did not return anything")
	}

	err, _ := returns[0].Interface().(error)
	return err
}
