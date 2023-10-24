// Package jsonx contains helper functions for dealing with JSON encoded types.
package jsonx

import (
	"fmt"
	"reflect"

	"github.com/fatih/structtag"
)

// Field is a JSON encoded field.
type Field struct {
	Name     string
	JSONName string
}

// StructFields returns the set of fields in a structure that will be encoded as JSON.
func StructFields(v any) ([]Field, error) {
	typ := reflect.TypeOf(v)
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%s is not a struct", typ)
	}

	var fields []Field
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if !f.IsExported() {
			continue
		}

		tags, err := structtag.Parse(string(f.Tag))
		if err != nil {
			return nil, fmt.Errorf("unable to parse tags '%s' for %s: %w", f.Tag, f.Name, err)
		}

		jsonTag := findTag(tags, "json")
		if jsonTag == nil || jsonTag.Name == "" {
			fields = append(fields, Field{
				Name:     f.Name,
				JSONName: f.Name,
			})
			continue
		}

		if jsonTag.Name == "-" {
			continue
		}

		fields = append(fields, Field{
			Name:     f.Name,
			JSONName: jsonTag.Name,
		})
	}

	return fields, nil
}

func findTag(tags *structtag.Tags, key string) *structtag.Tag {
	for _, tag := range tags.Tags() {
		if tag.Key == key {
			return tag
		}
	}

	return nil
}
