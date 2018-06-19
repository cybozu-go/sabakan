package etcd

import (
	"sort"
	"testing"

	"github.com/cybozu-go/sabakan"
)

func TestMachinesIndex(t *testing.T) {
	mi := newMachinesIndex()

	machines := []*sabakan.Machine{
		sabakan.NewMachine(sabakan.MachineSpec{Serial: "1", Product: "R630", Datacenter: "ty3", Role: "boot", BMC: sabakan.MachineBMC{Type: sabakan.BmcIpmi2}}),
		sabakan.NewMachine(sabakan.MachineSpec{Serial: "2", Product: "R630", Datacenter: "ty3", Role: "worker", BMC: sabakan.MachineBMC{Type: sabakan.BmcIpmi2}}),
		sabakan.NewMachine(sabakan.MachineSpec{Serial: "3", Product: "R730xd", Datacenter: "ty3", Role: "worker", BMC: sabakan.MachineBMC{Type: sabakan.BmcIpmi2}}),
	}

	for _, m := range machines {
		mi.AddIndex(m)
	}

	serials := mi.query(&sabakan.Query{Product: "R730xd"})
	if len(serials) != 1 {
		t.Fatal("wrong query count:", len(serials))
	}
	if serials[0] != "3" {
		t.Error("wrong query serial:", serials[0])
	}

	prev := sabakan.NewMachine(
		sabakan.MachineSpec{
			Serial:     "2",
			Product:    "R630",
			Datacenter: "ty3",
			Role:       "worker",
			BMC:        sabakan.MachineBMC{Type: sabakan.BmcIpmi2},
		})
	current := sabakan.NewMachine(
		sabakan.MachineSpec{
			Serial:     "2",
			Product:    "R730xd",
			Datacenter: "ty3",
			Role:       "worker",
			BMC:        sabakan.MachineBMC{Type: sabakan.BmcIpmi2},
		})
	current.Status.State = sabakan.StateRetiring

	mi.UpdateIndex(prev, current)

	serials = mi.query(&sabakan.Query{Product: "R730xd"})
	if len(serials) != 2 {
		t.Fatal("wrong query count:", len(serials))
	}
	sort.Strings(serials)
	if !(serials[0] == "2" && serials[1] == "3") {
		t.Error("wrong query serials:", serials)
	}

	serials = mi.query(&sabakan.Query{State: "retiring"})
	if len(serials) != 1 {
		t.Fatal("wrong query count:", len(serials))
	}

	mi.DeleteIndex(sabakan.NewMachine(
		sabakan.MachineSpec{
			Serial:     "3",
			Product:    "R730xd",
			Datacenter: "ty3",
			Role:       "worker",
			BMC:        sabakan.MachineBMC{Type: sabakan.BmcIpmi2},
		}))

	serials = mi.query(&sabakan.Query{Product: "R730xd"})
	if len(serials) != 1 {
		t.Fatal("wrong query count:", len(serials))
	}
	if serials[0] != "2" {
		t.Error("wrong query serial:", serials[0])
	}
}
