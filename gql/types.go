package gql

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/99designs/gqlgen/graphql"
)

// IPAddress represents "IPAddress" GraphQL custom scalar.
type IPAddress net.IP

// UnmarshalGQL implements graphql.Marshaler interface.
func (a *IPAddress) UnmarshalGQL(v interface{}) error {
	str, err := graphql.UnmarshalString(v)
	if err != nil {
		return fmt.Errorf("invalid IPAddress: %v, %v", v, err)
	}

	ip := net.ParseIP(str)
	if ip == nil {
		return fmt.Errorf("invalid IPAddress: %s", str)
	}

	*a = IPAddress(ip)
	return nil
}

// MarshalGQL implements graphql.Marshaler interface.
func (a IPAddress) MarshalGQL(w io.Writer) {
	graphql.MarshalString(net.IP(a).String()).MarshalGQL(w)
}

// DateTime represents "DateTime" GraphQL custom scalar.
type DateTime time.Time

// UnmarshalGQL implements graphql.Marshaler interface.
func (dt *DateTime) UnmarshalGQL(v interface{}) error {
	t, err := graphql.UnmarshalTime(v)
	if err != nil {
		return fmt.Errorf("invalid DateTime: %v, %v", v, err)
	}

	*dt = DateTime(t)
	return nil
}

// MarshalGQL implements graphql.Marshaler interface.
func (dt DateTime) MarshalGQL(w io.Writer) {
	graphql.MarshalTime(time.Time(dt)).MarshalGQL(w)
}
