package etcd

import (
	"sort"
	"testing"

	"github.com/cybozu-go/sabakan/v3"
)

func TestMachinesIndex(t *testing.T) {
	t.Parallel()

	mi := newMachinesIndex()

	machines := []*sabakan.Machine{
		sabakan.NewMachine(sabakan.MachineSpec{Serial: "1", Labels: map[string]string{"product": "R630", "datacenter": "ty3"}, Role: "boot", BMC: sabakan.MachineBMC{Type: "IPMI-2.0"}}),
		sabakan.NewMachine(sabakan.MachineSpec{Serial: "2", Labels: map[string]string{"product": "R630", "datacenter": "ty3"}, Role: "worker", BMC: sabakan.MachineBMC{Type: "IPMI-2.0"}}),
		sabakan.NewMachine(sabakan.MachineSpec{Serial: "3", Labels: map[string]string{"product": "R730xd", "datacenter": "ty3"}, Role: "worker", BMC: sabakan.MachineBMC{Type: "IPMI-2.0"}}),
	}

	for _, m := range machines {
		mi.AddIndex(m)
	}

	serials := mi.query(sabakan.Query{"labels": "product=R730xd"})
	if len(serials) != 1 {
		t.Fatal("wrong query count:", len(serials))
	}
	if serials[0] != "3" {
		t.Error("wrong query serial:", serials[0])
	}

	prev := sabakan.NewMachine(
		sabakan.MachineSpec{
			Serial: "2",
			Labels: map[string]string{
				"product":    "R630",
				"datacenter": "ty3",
			},
			Role: "worker",
			BMC:  sabakan.MachineBMC{Type: "IPMI-2.0"},
		})
	current := sabakan.NewMachine(
		sabakan.MachineSpec{
			Serial: "2",
			Labels: map[string]string{
				"product":    "R730xd",
				"datacenter": "ty3",
			},
			Role: "worker",
			BMC:  sabakan.MachineBMC{Type: "IPMI-2.0"},
		})
	current.Status.State = sabakan.StateRetiring

	mi.UpdateIndex(prev, current)

	serials = mi.query(sabakan.Query{"labels": "product=R730xd"})
	if len(serials) != 2 {
		t.Fatal("wrong query count:", len(serials))
	}
	sort.Strings(serials)
	if !(serials[0] == "2" && serials[1] == "3") {
		t.Error("wrong query serials:", serials)
	}

	serials = mi.query(sabakan.Query{"state": "retiring"})
	if len(serials) != 1 {
		t.Fatal("wrong query count:", len(serials))
	}

	mi.DeleteIndex(sabakan.NewMachine(
		sabakan.MachineSpec{
			Serial: "3",
			Labels: map[string]string{
				"product":    "R730xd",
				"datacenter": "ty3",
			},
			Role: "worker",
			BMC:  sabakan.MachineBMC{Type: "IPMI-2.0"},
		}))

	serials = mi.query(sabakan.Query{"labels": "product=R730xd"})
	if len(serials) != 1 {
		t.Fatal("wrong query count:", len(serials))
	}
	if serials[0] != "2" {
		t.Error("wrong query serial:", serials[0])
	}
}
