package etcd

import (
	"sort"
	"testing"

	"github.com/cybozu-go/sabakan"
)

func TestMachinesIndex(t *testing.T) {
	mi := newMachinesIndex()

	machines := []*sabakan.Machine{
		{Serial: "1", Product: "R630", Datacenter: "ty3", Role: "boot"},
		{Serial: "2", Product: "R630", Datacenter: "ty3", Role: "worker"},
		{Serial: "3", Product: "R730xd", Datacenter: "ty3", Role: "worker"},
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

	mi.UpdateIndex(
		&sabakan.Machine{
			Serial:     "2",
			Product:    "R630",
			Datacenter: "ty3",
			Role:       "worker",
		},
		&sabakan.Machine{
			Serial:     "2",
			Product:    "R730xd",
			Datacenter: "ty3",
			Role:       "worker",
		})

	serials = mi.query(&sabakan.Query{Product: "R730xd"})
	if len(serials) != 2 {
		t.Fatal("wrong query count:", len(serials))
	}
	sort.Strings(serials)
	if !(serials[0] == "2" && serials[1] == "3") {
		t.Error("wrong query serials:", serials)
	}

	mi.DeleteIndex(&sabakan.Machine{
		Serial:     "3",
		Product:    "R730xd",
		Datacenter: "ty3",
		Role:       "worker",
	})

	serials = mi.query(&sabakan.Query{Product: "R730xd"})
	if len(serials) != 1 {
		t.Fatal("wrong query count:", len(serials))
	}
	if serials[0] != "2" {
		t.Error("wrong query serial:", serials[0])
	}
}
