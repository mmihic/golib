package csvmarshal

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"io"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/mmihic/golib/src/pkg/ptr"
	"github.com/stretchr/testify/assert"
)

const writeFiles = false

type Address struct {
	Street1 string
	City    string
	State   string
	Zipcode string
}

type Person struct {
	FirstName, LastName string
	MailingAddress      *Address
}

type Place struct {
	Name    string
	Address // anonymous
}

type PlaceWithOptionalName struct {
	Name    *string
	Address // anonymous
}

func TestMarshaller(t *testing.T) {
	for _, tt := range []struct {
		name     string
		val      any
		expected string
		anyOrder bool
	}{
		{
			"simple struct",
			Address{
				Street1: "209 W Houston St",
				City:    "New York",
				State:   "NY",
				Zipcode: "10014",
			},
			"testdata/simple_struct.csv",
			false,
		},
		{
			"embedded ptr to struct",
			Person{
				FirstName: "Hanna",
				LastName:  "Banana",
				MailingAddress: &Address{
					Street1: "209 W Houston St",
					City:    "New York",
					State:   "NY",
					Zipcode: "10014",
				},
			},
			"testdata/embedded_ptr_to_struct.csv",
			false,
		},
		{
			"anonymous struct",
			Place{
				Name: "Film Forum",
				Address: Address{
					Street1: "209 W Houston St",
					City:    "New York",
					State:   "NY",
					Zipcode: "10014",
				},
			},
			"testdata/anonymous_struct.csv",
			false,
		},
		{
			"slice of structs",
			[]Address{
				{
					Street1: "209 W Houston St",
					City:    "New York",
					State:   "NY",
					Zipcode: "10014",
				},
				{
					Street1: "636 W 28th St",
					City:    "New York",
					State:   "NY",
					Zipcode: "10001",
				},
				{
					Street1: "375 W Broadway",
					City:    "New York",
					State:   "NY",
					Zipcode: "10012",
				},
			},
			"testdata/slice_of_structs.csv",
			false,
		},
		{
			"slice of pointers to structs",
			[]*Address{
				{
					Street1: "209 W Houston St",
					City:    "New York",
					State:   "NY",
					Zipcode: "10014",
				},
				{
					Street1: "636 W 28th St",
					City:    "New York",
					State:   "NY",
					Zipcode: "10001",
				},
				{
					Street1: "375 W Broadway",
					City:    "New York",
					State:   "NY",
					Zipcode: "10012",
				},
			},
			"testdata/slice_of_ptr_to_structs.csv",
			false,
		},
		{
			"map to ptr to struct",
			map[string]*Address{
				"Hanna Banana": {
					Street1: "209 W Houston St",
					City:    "New York",
					State:   "NY",
					Zipcode: "10014",
				},
				"June Prune": {
					Street1: "636 W 28th St",
					City:    "New York",
					State:   "NY",
					Zipcode: "10001",
				},
				"Barry Cherry": {
					Street1: "375 W Broadway",
					City:    "New York",
					State:   "NY",
					Zipcode: "10012",
				},
			},
			"testdata/map_of_ptr_to_structs.csv",
			true,
		},
		{
			"map of primitives",
			map[string]float64{
				"Hanna Banana": 38.6794,
				"June Prune":   17.452,
				"Barry Cherry": 4.956,
			},
			"testdata/map_of_primitives.csv",
			true,
		},
		{
			"nil slice element",
			[]*Address{
				{
					Street1: "375 W Broadway",
					City:    "New York",
					State:   "NY",
					Zipcode: "10012",
				},
				nil,
				{
					Street1: "209 W Houston St",
					City:    "New York",
					State:   "NY",
					Zipcode: "10014",
				},
			},
			"testdata/nil_slice_element.csv",
			false,
		},
		{
			"nil embedded struct",
			[]Person{
				{
					FirstName: "Hanna",
					LastName:  "Banana",
					MailingAddress: &Address{
						Street1: "209 W Houston St",
						City:    "New York",
						State:   "NY",
						Zipcode: "10014",
					},
				},
				{
					FirstName: "June",
					LastName:  "Prune",
				},
				{
					FirstName: "Barry",
					LastName:  "Cherry",
					MailingAddress: &Address{
						Street1: "375 W Broadway",
						City:    "New York",
						State:   "NY",
						Zipcode: "10012",
					},
				},
			},
			"testdata/nil_embedded_struct.csv",
			false,
		},
		{
			"nil field",
			[]PlaceWithOptionalName{
				{
					Name: ptr.To("Film Forum"),
					Address: Address{
						Street1: "209 W Houston St",
						City:    "New York",
						State:   "NY",
						Zipcode: "10014",
					},
				},
				{
					Address: Address{
						Street1: "375 W Broadway",
						City:    "New York",
						State:   "NY",
						Zipcode: "10012",
					},
				},
			},
			"testdata/nil_field.csv",
			false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assertMarshalMatches(t, tt.val, tt.expected, tt.anyOrder)
		})
	}
}

func TestInvalidMarshallers(t *testing.T) {
	for _, tt := range []struct {
		val         any
		expectedErr string
	}{
		{100, "unable to create Marshaller for 'int'"},
		{103.45, "unable to create Marshaller for 'float64'"},
		{"floorp", "unable to create Marshaller for 'string'"},
		{ptr.To("floorp"), "unable to create Marshaller for 'string'"},
		{
			map[string][]string{
				"foo": {"bar", "zed", "foo"},
			},
			"cannot convert type '[]string' to csv",
		},
	} {
		typ := reflect.TypeOf(tt.val)
		t.Run(typ.Name(), func(t *testing.T) {
			_, err := NewMarshaller(typ)
			if !assert.Error(t, err) {
				return
			}

			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestPrimitives(t *testing.T) {
	for _, tt := range []struct {
		val      any
		expected []string
	}{
		{
			complex(100.45, 3.56),
			[]string{"(100.45+3.56i)"},
		},
		{
			100.23,
			[]string{"100.2300000000"},
		},
		{
			float32(100.23),
			[]string{"100.2300033569"},
		},
		{
			3456,
			[]string{"3456"},
		},
		{
			int32(3456),
			[]string{"3456"},
		},
		{
			int64(math.MaxInt64 - 1),
			[]string{"9223372036854775806"},
		},
		{
			uint8(math.MaxUint8 - 1),
			[]string{"254"},
		},
		{
			uint(math.MaxUint32 - 1),
			[]string{"4294967294"},
		},
		{
			uint32(math.MaxUint32 - 1),
			[]string{"4294967294"},
		},
		{
			uint64(math.MaxUint64 - 1),
			[]string{"18446744073709551614"},
		},
	} {
		t.Run(reflect.TypeOf(tt.val).Name(), func(t *testing.T) {
			val := reflect.ValueOf(tt.val)
			rm, err := newRowMapper(val.Type())
			if !assert.NoError(t, err) {
				return
			}

			actual, process := rm.Values(val, "")
			assert.True(t, process)
			assert.Equal(t, actual, tt.expected)
		})
	}
}

func assertMarshalMatches(t *testing.T, val any, filename string, anyOrder bool) bool {
	m, err := NewMarshaller(reflect.TypeOf(val))
	if !assert.NoError(t, err) {
		return false
	}

	var buf bytes.Buffer

	w := csv.NewWriter(&buf)
	err = w.Write(m.Headers())
	if !assert.NoError(t, err) {
		return false
	}

	err = m.Marshal(w, val, "")
	if !assert.NoError(t, err) {
		return false
	}

	w.Flush()

	return assertCSVFilesMatch(t, filename, buf.String(), anyOrder)
}

func assertCSVFilesMatch(t *testing.T, filename string, actual string, anyOrder bool) bool {
	if writeFiles {
		err := os.WriteFile(filename, []byte(actual), 0664)
		if !assert.NoError(t, err) {
			return false
		}
		return true
	}

	f, err := os.Open(filename)
	if !assert.NoError(t, err) {
		return false
	}

	if anyOrder {
		return compareLinesAnyOrder(t, f, strings.NewReader(actual))
	}

	expected, err := io.ReadAll(f)
	if !assert.NoError(t, err) {
		return false
	}

	return assert.Equal(t, string(expected), actual)
}

func compareLinesAnyOrder(t *testing.T, r1 io.Reader, r2 io.Reader) bool {
	lines := make(map[string]struct{})
	b1 := bufio.NewScanner(r1)
	for b1.Scan() {
		lines[b1.Text()] = struct{}{}
	}

	if !assert.NoError(t, b1.Err()) {
		return false
	}

	b2 := bufio.NewScanner(r2)
	for b2.Scan() {
		_, ok := lines[b2.Text()]
		if !assert.Truef(t, ok, "unable to find line %s", b2.Text()) {
			return false
		}

		delete(lines, b2.Text())
	}

	return assert.NoError(t, b2.Err())
}
