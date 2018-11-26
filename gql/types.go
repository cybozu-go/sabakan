package gql

import (
	fmt "fmt"
	io "io"
	"net"
	"time"
)

// IPAddress represents "IPAddress" GraphQL custom scalar.
type IPAddress net.IP

// UnmarshalGQL implements graphql.Marshaler interface.
func (a *IPAddress) UnmarshalGQL(v interface{}) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("invalid IPAddress: %T, %v", v, v)
	}

	ip := net.ParseIP(s)
	if ip == nil {
		return fmt.Errorf("invalid IPAddress: %s", s)
	}

	*a = IPAddress(ip)
	return nil
}

// MarshalGQL implements graphql.Marshaler interface.
func (a IPAddress) MarshalGQL(w io.Writer) {
	io.WriteString(w, net.IP(a).String())
}

// DateTime represents "DateTime" GraphQL custom scalar.
type DateTime time.Time

// UnmarshalGQL implements graphql.Marshaler interface.
func (dt *DateTime) UnmarshalGQL(v interface{}) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("invalid DateTime: %T, %v", v, v)
	}

	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return fmt.Errorf("invalid DateTime: %s, %v", s, err)
	}

	*dt = DateTime(t)
	return nil
}

// MarshalGQL implements graphql.Marshaler interface.
func (dt DateTime) MarshalGQL(w io.Writer) {
	io.WriteString(w, time.Time(dt).Format(time.RFC3339Nano))
}
