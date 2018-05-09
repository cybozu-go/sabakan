package sabakan

// MachineJSON is a struct to encode/decode Machine as JSON.
type MachineJSON struct {
	Serial           string                    `json:"serial"`
	Product          string                    `json:"product"`
	Datacenter       string                    `json:"datacenter"`
	Rack             *uint32                   `json:"rack"`
	NodeNumberOfRack *uint32                   `json:"node-number-of-rack"`
	Role             string                    `json:"role"`
	Network          map[string]MachineNetwork `json:"network"`
	BMC              MachineBMC                `json:"bmc"`
}

// ToMachine creates *Machine.
func (mj *MachineJSON) ToMachine() *Machine {
	return &Machine{
		Serial:           mj.Serial,
		Product:          mj.Product,
		Datacenter:       mj.Datacenter,
		Rack:             *mj.Rack,
		NodeNumberOfRack: *mj.NodeNumberOfRack,
		Role:             mj.Role,
		Network:          mj.Network,
		BMC:              mj.BMC,
	}
}

// Machine represents a server hardware.
type Machine struct {
	Serial           string
	Product          string
	Datacenter       string
	Rack             uint32
	NodeNumberOfRack uint32
	Role             string
	Network          map[string]MachineNetwork
	BMC              MachineBMC
}

// ToJSON creates *MachineJSON
func (m *Machine) ToJSON() *MachineJSON {
	rack := m.Rack
	num := m.NodeNumberOfRack
	return &MachineJSON{
		Serial:           m.Serial,
		Product:          m.Product,
		Datacenter:       m.Datacenter,
		Rack:             &rack,
		NodeNumberOfRack: &num,
		Role:             m.Role,
		Network:          m.Network,
		BMC:              m.BMC,
	}
}

// MachineNetwork is a network interface struct for Machine
type MachineNetwork struct {
	IPv4 []string `json:"ipv4"`
	IPv6 []string `json:"ipv6"`
	Mac  string   `json:"mac"`
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
	IPv4 []string `json:"ipv4"`
}
