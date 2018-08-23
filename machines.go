package sabakan

import (
	"errors"
	"regexp"
	"time"
)

const (
	// BmcIdrac9 is BMC type for iDRAC-9
	BmcIdrac9 = "iDRAC-9"
	// BmcIpmi2 is BMC type for IPMI-2.0
	BmcIpmi2 = "IPMI-2.0"
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
	reValidRole      = regexp.MustCompile(`^[a-zA-Z][0-9a-zA-Z._-]*$`)
	reValidLabelVal  = regexp.MustCompile(`^[[:print:]]+$`)
	reValidLabelName = regexp.MustCompile(`^[a-z0-9A-Z-_/.]+$`)
)

// IsValidRole returns true if role is valid as machine role
func IsValidRole(role string) bool {
	return reValidRole.MatchString(role)
}

// IsValidLabelName returns true if label name is valid
func IsValidLabelName(name string) bool {
	return reValidLabelName.MatchString(name)
}

// IsValidLabelValue returns true if label value is valid
func IsValidLabelValue(value string) bool {
	return reValidLabelVal.MatchString(value)
}

// MachineBMC is a bmc interface struct for Machine
type MachineBMC struct {
	IPv4 string `json:"ipv4"`
	IPv6 string `json:"ipv6"`
	Type string `json:"type"`
}

// MachineSpec is a set of attributes to define a machine.
type MachineSpec struct {
	Serial      string            `json:"serial"`
	Labels      map[string]string `json:"labels"`
	Rack        uint              `json:"rack"`
	IndexInRack uint              `json:"index-in-rack"`
	Role        string            `json:"role"`
	IPv4        []string          `json:"ipv4"`
	IPv6        []string          `json:"ipv6"`
	BMC         MachineBMC        `json:"bmc"`
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
