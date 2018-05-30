package sabakan

const (
	BMC_iDRAC_9 = "iDRAC-9"
	BMC_IPMI_2  = "IPMI-2.0"
)

// Machine represents a server hardware.
type Machine struct {
	Serial      string                    `json:"serial"`
	Product     string                    `json:"product"`
	Datacenter  string                    `json:"datacenter"`
	Rack        uint                      `json:"rack"`
	IndexInRack uint                      `json:"index-in-rack"`
	Role        string                    `json:"role"`
	Network     map[string]MachineNetwork `json:"network"`
	BMC         MachineBMC                `json:"bmc"`
}

// MachineNetwork is a network interface struct for Machine
type MachineNetwork struct {
	IPv4 []string `json:"ipv4"`
	IPv6 []string `json:"ipv6"`
}

func (n MachineNetwork) hasIPv4(ipv4 string) bool {
	for _, t := range n.IPv4 {
		if t == ipv4 {
			return true
		}
	}
	return false
}

func (n MachineNetwork) hasIPv6(ipv6 string) bool {
	for _, t := range n.IPv6 {
		if t == ipv6 {
			return true
		}
	}
	return false
}

// MachineBMC is a bmc interface struct for Machine
type MachineBMC struct {
	IPv4 string `json:"ipv4"`
	IPv6 string `json:"ipv6"`
	Type string `json:"type"`
}
