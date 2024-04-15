package csv

import (
	"encoding/csv"
	"fmt"
	"reflect"
)

// A Marshaller marshals objects into CSV format.
type Marshaller interface {
	Headers() []string
	Encode(w *csv.Writer, val any, nilValue string) error
}

// NewMarshaller creates a marshaller for the given type.
func NewMarshaller(typ reflect.Type) (Marshaller, error) {
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array {
		return newSliceMarshaller(typ)
	}

	if typ.Kind() == reflect.Map {
		return newMapMarshaller(typ)
	}

	if typ.Kind() == reflect.Struct {
		return newSingleRowMarshaller(typ)
	}

	return nil, fmt.Errorf("unable to create Marshaller for '%s'", typ.Name())
}

// singleRowMarshaller is a Marshaller for a single struct.
type singleRowMarshaller struct {
	rowMapper RowMapper
}

func newSingleRowMarshaller(typ reflect.Type) (Marshaller, error) {
	rowMapper, err := newStructMapper(typ)
	if err != nil {
		return nil, err
	}

	return &singleRowMarshaller{
		rowMapper: rowMapper,
	}, nil
}

func (m *singleRowMarshaller) Headers() []string {
	return m.rowMapper.Headers()
}

func (m *singleRowMarshaller) Encode(w *csv.Writer, val any, nilValue string) error {
	rv, ok := val.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(val)
	}

	if values, processRow := m.rowMapper.Values(rv, nilValue); processRow {
		return w.Write(values)
	}

	return nil
}

// A sliceMarshaller is a Marshaller for slices of structs or slices of primitives.
type sliceMarshaller struct {
	elemMapper RowMapper
}

func newSliceMarshaller(typ reflect.Type) (Marshaller, error) {
	if typ.Kind() != reflect.Slice && typ.Kind() != reflect.Array {
		return nil, fmt.Errorf("cannot create slice Marshaller for '%s'", typ.String())
	}

	elemMapper, err := newRowMapper(typ.Elem())
	if err != nil {
		return nil, err
	}

	return &sliceMarshaller{
		elemMapper: elemMapper,
	}, nil
}

func (m *sliceMarshaller) Headers() []string {
	return m.elemMapper.Headers()
}

func (m *sliceMarshaller) Encode(w *csv.Writer, val any, nilValue string) error {
	rv, ok := val.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(val)
	}

	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i)

		if values, processRow := m.elemMapper.Values(elem, nilValue); processRow {
			if err := w.Write(values); err != nil {
				return err
			}
		}
	}

	return nil
}

// A mapMarshaller is a Marshaller for maps of primitives or maps of structs.
type mapMarshaller struct {
	headers   []string
	keyMapper RowMapper
	valMapper RowMapper
}

func newMapMarshaller(typ reflect.Type) (Marshaller, error) {
	keyMapper := &primitiveRowMapper{}
	valMapper, err := newRowMapper(typ.Elem())
	if err != nil {
		return nil, err
	}

	valHeaders := valMapper.Headers()
	headers := make([]string, 0, len(valHeaders)+1)
	headers = append(headers, "key")
	if len(valHeaders) == 0 {
		// value is a primitive type
		headers = append(headers, "val")
	} else {
		headers = append(headers, valHeaders...)
	}
	return &mapMarshaller{
		headers:   headers,
		keyMapper: keyMapper,
		valMapper: valMapper,
	}, nil
}

func (m *mapMarshaller) Encode(w *csv.Writer, val any, nilValue string) error {
	rv, ok := val.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(val)
	}

	iter := rv.MapRange()
	for iter.Next() {
		key, val := iter.Key(), iter.Value()
		rowValues := make([]string, 0, len(m.headers))

		if keyValues, processKey := m.keyMapper.Values(key, nilValue); processKey {
			rowValues = append(rowValues, keyValues...)
		}

		if valValues, processVal := m.valMapper.Values(val, nilValue); processVal {
			rowValues = append(rowValues, valValues...)
		}

		if len(rowValues) != 0 {
			if err := w.Write(rowValues); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *mapMarshaller) Headers() []string {
	return m.headers
}
