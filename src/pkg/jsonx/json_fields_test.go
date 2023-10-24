package jsonx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONFields(t *testing.T) {
	for _, tt := range []struct {
		name    string
		in      interface{}
		wantErr string
		want    []Field
	}{
		{
			"simple struct",
			struct {
				Name       string `json:""`                // has json tag but no explicit json name
				Value      string `json:"value,omitempty"` // has explicit JSON name
				Ignored    string `json:"-"`               // ignored from JSON output
				unexported string // not exported, so ignored from JSON output
				NoJSONTag  string // has no JSON tag so uses same name as struct
			}{},
			"",
			[]Field{
				{Name: "Name", JSONName: "Name"},
				{Name: "Value", JSONName: "value"},
				{Name: "NoJSONTag", JSONName: "NoJSONTag"},
			},
		},
		{
			"ptr to struct",
			&struct {
				Name  string `json:""`                // has json tag but no explicit json name
				Value string `json:"value,omitempty"` // has explicit JSON name
			}{},
			"",
			[]Field{
				{Name: "Name", JSONName: "Name"},
				{Name: "Value", JSONName: "value"},
			},
		},
		{
			"not a struct",
			[]string{},
			"[]string is not a struct",
			nil,
		},
		{
			"bad struct tag",
			struct {
				InvalidField string `json::`
			}{},
			"unable to parse tags 'json::' for InvalidField: bad syntax for struct tag value",
			nil,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fields, err := StructFields(tt.in)
			if tt.wantErr != "" {
				if !assert.Error(t, err) {
					return
				}

				assert.Equal(t, tt.wantErr, err.Error())
				return
			}

			assert.Equal(t, tt.want, fields)
		})
	}
}
