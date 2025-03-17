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
		fieldName := []rune(field.Name)
		fieldName[0] = unicode.ToLower(fieldName[0])
		return []string{string(fieldName)}
	}

	return strings.Split(value, ",")
}

func getDefaultTags(field *reflect.StructField) []string {
	value, ok := field.Tag.Lookup(TagDefault)
	if !ok {
		return nil
	}

	return strings.Split(value, ",")
}
