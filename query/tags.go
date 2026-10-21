package query

import (
	"reflect"
	"strings"
	"unicode"
)

const (
	TagName    = "query"
	TagDefault = "default"
)

func getNameTags(field *reflect.StructField) []string {
	value, ok := field.Tag.Lookup(TagName)
	if !ok {
		return []string{defaultName(field)}
	}

	parts := strings.Split(value, ",")
	if parts[0] == "" {
		parts[0] = defaultName(field)
	}

	return parts
}

func defaultName(field *reflect.StructField) string {
	fieldName := []rune(field.Name)
	fieldName[0] = unicode.ToLower(fieldName[0])
	return string(fieldName)
}

func getDefaultTags(field *reflect.StructField) []string {
	value, ok := field.Tag.Lookup(TagDefault)
	if !ok {
		return nil
	}

	if !isSlice(field.Type) {
		return []string{value}
	}

	return strings.Split(value, ",")
}

func isSlice(typ reflect.Type) bool {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	return typ.Kind() == reflect.Slice
}
