// Package netx contains network related utilities.
package netx

import (
	"encoding"
	"fmt"
	"net"
	"strconv"
)

// A HostPort is a host/port pair
type HostPort struct {
	Host string
	Port int
}

// MustParseHostPort parses the given host:port string, panicking if
// the string cannot be parsed.
func MustParseHostPort(hostport string) HostPort {
	parsed, err := ParseHostPort(hostport)
	if err != nil {
		panic(err)
	}

	return parsed
}

// ParseHostPort parses the given host:port string into a host port structure.
func ParseHostPort(hostport string) (HostPort, error) {
	host, portString, err := net.SplitHostPort(hostport)
	if err != nil {
		return HostPort{}, err
	}

	port, err := strconv.Atoi(portString)
	if err != nil {
		return HostPort{}, fmt.Errorf("unable to parse port: %w", err)
	}

	return HostPort{
		Host: host,
		Port: port,
	}, nil
}

// String converts the given host port into a string.
func (hp HostPort) String() string {
	return net.JoinHostPort(hp.Host, strconv.Itoa(hp.Port))
}

// UnmarshalText unmarshals the host port from a text encoding.
func (hp *HostPort) UnmarshalText(data []byte) error {
	parsed, err := ParseHostPort(string(data))
	if err != nil {
		return err
	}

	*hp = parsed
	return nil
}

var (
	_ encoding.TextUnmarshaler = &HostPort{}
)
