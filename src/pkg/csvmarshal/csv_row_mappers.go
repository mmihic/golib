package csvmarshal

import (
	"fmt"
	"reflect"
	"strconv"
)

// A RowMapper maps a single row into the CSV representation.
type RowMapper interface {
	Headers() []string
	Values(val reflect.Value, nilValue string) ([]string, bool)
}

// A structMapper is a RowMapper that converts a struct into a CSV row.
type structMapper struct {
	headers      []string
	fields       []reflect.StructField
	fieldMappers []RowMapper
}

func newStructMapper(typ reflect.Type) (RowMapper, error) {
	var (
		fields       = make([]reflect.StructField, 0, typ.NumField())
		fieldMappers = make([]RowMapper, 0, typ.NumField())
		headers      = make([]string, 0, typ.NumField())
	)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fields = append(fields, field)

		// Get the converter for the field
		fieldMapper, err := newRowMapper(field.Type)
		if err != nil {
			return nil, err
		}
		fieldMappers = append(fieldMappers, fieldMapper)

		// Add the header(s) for the field. If the field is a compound type, prefix
		// all field internal headers with the field name. If the field is a primitive
		// type, the only header is the name of the field itself
		if fieldHeaders := fieldMapper.Headers(); len(fieldHeaders) > 0 {
			prefix := ""
			if !field.Anonymous {
				prefix = field.Name + "."
			}

			for _, fieldHeader := range fieldHeaders {
				headers = append(headers, prefix+fieldHeader)
			}
		} else {
			headers = append(headers, field.Name)
		}
	}

	return &structMapper{
		headers:      headers,
		fields:       fields,
		fieldMappers: fieldMappers,
	}, nil
}

func (c *structMapper) Headers() []string {
	return c.headers
}

func (c *structMapper) Values(val reflect.Value, nilValue string) ([]string, bool) {
	val = deref(val)
	if val == zeroValue {
		return nil, false
	}

	values := make([]string, 0, len(c.fields))
	for i, field := range c.fields {
		fieldVal := val.FieldByIndex(field.Index)
		if isNillable(fieldVal.Kind()) && fieldVal.IsNil() {
			values = append(values, nilValue)
			continue
		}

		if fieldValues, processRow := c.fieldMappers[i].Values(fieldVal, nilValue); processRow {
			values = append(values, fieldValues...)
		}
	}

	return values, len(values) != 0
}

// newRowMapper creates a new RowMapper for a given type.
func newRowMapper(typ reflect.Type) (RowMapper, error) {
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if isPrimitive(typ.Kind()) {
		return &primitiveRowMapper{}, nil
	}

	if typ.Kind() == reflect.Struct {
		return newStructMapper(typ)
	}

	return nil, fmt.Errorf("cannot convert type '%s' to csv", typ.String())
}

// primitiveRowMapper maps a primitive value to a row.
type primitiveRowMapper struct {
}

func (c *primitiveRowMapper) Headers() []string { return nil }
func (c *primitiveRowMapper) Values(val reflect.Value, _ string) ([]string, bool) {
	val = deref(val)

	if val.CanFloat() {
		return []string{strconv.FormatFloat(val.Float(), 'f', 10, 64)}, true
	}

	if val.CanInt() {
		return []string{strconv.FormatInt(val.Int(), 10)}, true
	}

	if val.CanUint() {
		return []string{strconv.FormatUint(val.Uint(), 10)}, true
	}

	if val.CanComplex() {
		return []string{strconv.FormatComplex(val.Complex(), 'g', 10, 128)}, true
	}

	return []string{deref(val).String()}, true
}

func isNillable(kind reflect.Kind) bool {
	_, ok := nillable[kind]
	return ok
}

var nillable = map[reflect.Kind]struct{}{
	reflect.Chan:          {},
	reflect.Func:          {},
	reflect.Map:           {},
	reflect.Pointer:       {},
	reflect.UnsafePointer: {},
	reflect.Interface:     {},
	reflect.Slice:         {},
}

func isPrimitive(kind reflect.Kind) bool {
	_, ok := primitives[kind]
	return ok
}

var primitives = map[reflect.Kind]struct{}{
	reflect.Int:        {},
	reflect.Int32:      {},
	reflect.Int64:      {},
	reflect.Uint:       {},
	reflect.Uint8:      {},
	reflect.Uint16:     {},
	reflect.Uint32:     {},
	reflect.Uint64:     {},
	reflect.Bool:       {},
	reflect.Float32:    {},
	reflect.Float64:    {},
	reflect.String:     {},
	reflect.Complex128: {},
	reflect.Complex64:  {},
}

var zeroValue reflect.Value

func deref(rv reflect.Value) reflect.Value {
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	return rv
}
