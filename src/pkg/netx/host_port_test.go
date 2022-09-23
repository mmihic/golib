package netx

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHostPort(t *testing.T) {
	hostPort, err := ParseHostPort("foo.bar.com:8080")
	require.NoError(t, err)
	assert.Equal(t, hostPort.Host, "foo.bar.com")
	assert.Equal(t, hostPort.Port, 8080)
}

func TestHostPortError(t *testing.T) {
	_, err := ParseHostPort("foo.bar.com:8080:384")
	require.Error(t, err)
	assert.Equal(t, "address foo.bar.com:8080:384: too many colons in address", err.Error())

	_, err = ParseHostPort("foo.bar.com:not_an_int")
	require.Error(t, err)
	assert.Equal(t, "unable to parse port: strconv.Atoi: parsing \"not_an_int\": invalid syntax", err.Error())
}

func TestHostPortUnmarshalJSON(t *testing.T) {
	type server struct {
		Name     string   `json:"name"`
		HostPort HostPort `json:"hostPort"`
	}

	type servers struct {
		Servers []server `json:"servers"`
	}

	var s servers
	err := json.Unmarshal([]byte(`
{
	"servers": [
		{"name": "foo", "hostPort":"foo.boodle.com:9005"},
		{"name": "zed", "hostPort":"zed.zeemie.com:9008"}
	]
}
`), &s)

	require.NoError(t, err)
	assert.Equal(t, &servers{
		Servers: []server{
			{Name: "foo", HostPort: HostPort{
				Host: "foo.boodle.com",
				Port: 9005,
			}},
			{Name: "zed", HostPort: HostPort{
				Host: "zed.zeemie.com",
				Port: 9008,
			}},
		},
	}, &s)
}
