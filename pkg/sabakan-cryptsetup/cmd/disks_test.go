package cmd

import (
	"reflect"
	"testing"
)

func TestDisk(t *testing.T) {
	t.Parallel()

	d := Disk{
		name:       "sda",
		sectorSize: 4096,
		size512:    2048,
	}

	if d.Device() != "/dev/sda" {
		t.Error(`d.Device() != "/dev/sda"`, d.Device())
	}
	if d.CryptName() != "crypt-sda" {
		t.Error(`d.CryptName() != "crypt-sda"`, d.CryptName())
	}
	if d.CryptDevice() != "/dev/mapper/crypt-sda" {
		t.Error(`d.CryptDevice() != "/dev/mapper/crypt-sda"`, d.CryptDevice())
	}
	if d.Name() != "sda" {
		t.Error(`d.Name() != "sda"`, d.Name())
	}
	if d.SectorSize() != 4096 {
		t.Error(`d.SectorSize() != 4096`, d.SectorSize())
	}
	if d.Size512() != 2048 {
		t.Error(`d.Size512() != 2048`, d.Size512())
	}
}

func TestFindDisks(t *testing.T) {
	t.Parallel()

	disks, err := findDisks([]string{"sd*"}, "./testdata")
	if err != nil {
		t.Fatal(err)
	}
	if len(disks) != 0 {
		t.Error("disks should be excluded:", disks)
	}

	//   - loop0 has no "device" link      -> excluded
	//   - sda is read-only                -> excluded
	//   - sdb is removable                -> excluded
	//   - sdc is normal (hidden=0, ro=0)  -> included
	//   - sdd is normal (hidden=0, ro=0)  -> included
	//   - sde is hidden (hidden=1)        -> excluded
	disks, err = findDisks(nil, "./testdata")
	if err != nil {
		t.Fatal(err)
	}
	expected := []Disk{
		{name: "sdc", sectorSize: 512, size512: 2048},
		{name: "sdd", sectorSize: 4096, size512: 2048},
	}
	if !reflect.DeepEqual(disks, expected) {
		t.Errorf("unexpected disks: got %+v, want %+v", disks, expected)
	}
}
