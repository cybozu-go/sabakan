package sabakan

import (
	"errors"
	"regexp"
	"time"
)

// MachineState represents a machine's state.
type MachineState string

// String implements fmt.Stringer interface.
func (ms MachineState) String() string {
	return string(ms)
}

// Machine state definitions.
const (
	StateHealthy   = MachineState("healthy")
	StateUnhealthy = MachineState("unhealthy")
	StateDead      = MachineState("dead")
	StateRetiring  = MachineState("retiring")
	StateRetired   = MachineState("retired")
)

var (
	reValidRole    = regexp.MustCompile(`^[a-zA-Z][0-9a-zA-Z._-]*$`)
	reValidBmcType = regexp.MustCompile(`^[a-z0-9A-Z-_/.]+$`)
)

// IsValidRole returns true if role is valid as machine role
func IsValidRole(role string) bool {
	return reValidRole.MatchString(role)
}

// IsValidBmcType returns true if role is valid as BMC type
func IsValidBmcType(bmcType string) bool {
	return reValidBmcType.MatchString(bmcType)
}

// MachineBMC is a bmc interface struct for Machine
type MachineBMC struct {
	IPv4 string `json:"ipv4"`
	IPv6 string `json:"ipv6"`
	Type string `json:"type"`
}

// MachineSpec is a set of attributes to define a machine.
type MachineSpec struct {
	Serial      string     `json:"serial"`
	Product     string     `json:"product"`
	Datacenter  string     `json:"datacenter"`
	Rack        uint       `json:"rack"`
	IndexInRack uint       `json:"index-in-rack"`
	Role        string     `json:"role"`
	IPv4        []string   `json:"ipv4"`
	IPv6        []string   `json:"ipv6"`
	BMC         MachineBMC `json:"bmc"`
}

// Machine represents a server hardware.
type Machine struct {
	Spec MachineSpec `json:"spec"`

	Status struct {
		Timestamp time.Time    `json:"timestamp"`
		State     MachineState `json:"state"`
	} `json:"status"`
}

// NewMachine creates a new machine instance.
func NewMachine(spec MachineSpec) *Machine {
	return &Machine{
		Spec: spec,
		Status: struct {
			Timestamp time.Time    `json:"timestamp"`
			State     MachineState `json:"state"`
		}{
			time.Now().UTC(),
			StateHealthy,
		},
	}
}

// SetState sets the state of the machine.
func (m *Machine) SetState(ms MachineState) error {
	switch m.Status.State {
	case StateHealthy, StateUnhealthy, StateDead:
		if ms == StateRetired {
			return errors.New("transition to retired is forbidden")
		}
	case StateRetiring:
		if ms != StateRetired {
			return errors.New("transition to state other than retired is forbidden")
		}
	case StateRetired:
		if ms != StateHealthy {
			return errors.New("transition to state other than healthy is forbidden")
		}
	}

	if m.Status.State != ms {
		m.Status.State = ms
		m.Status.Timestamp = time.Now().UTC()
	}
	return nil
}
