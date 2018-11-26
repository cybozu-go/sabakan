// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package gql

import (
	fmt "fmt"
	io "io"
	strconv "strconv"
)

// Label represents an arbitrary key-value pairs.
type Label struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// LabelInput represents a label to search machines.
type LabelInput struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// MachineParams is a set of input parameters to search machines.
type MachineParams struct {
	Labels              []LabelInput   `json:"labels"`
	Racks               []int          `json:"racks"`
	Roles               []string       `json:"roles"`
	States              []MachineState `json:"states"`
	MinDaysBeforeRetire *int           `json:"minDaysBeforeRetire"`
}

// MachineState enumerates machine states.
type MachineState string

const (
	MachineStateUninitialized MachineState = "UNINITIALIZED"
	MachineStateHealthy       MachineState = "HEALTHY"
	MachineStateUnhealthy     MachineState = "UNHEALTHY"
	MachineStateUnreachable   MachineState = "UNREACHABLE"
	MachineStateUpdating      MachineState = "UPDATING"
	MachineStateRetiring      MachineState = "RETIRING"
	MachineStateRetired       MachineState = "RETIRED"
)

func (e MachineState) IsValid() bool {
	switch e {
	case MachineStateUninitialized, MachineStateHealthy, MachineStateUnhealthy, MachineStateUnreachable, MachineStateUpdating, MachineStateRetiring, MachineStateRetired:
		return true
	}
	return false
}

func (e MachineState) String() string {
	return string(e)
}

func (e *MachineState) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = MachineState(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid MachineState", str)
	}
	return nil
}

func (e MachineState) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
