package cmd

import "testing"

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

	disks, err = findDisks(nil, "./testdata")
	if err != nil {
		t.Fatal(err)
	}
	if len(disks) != 2 {
		t.Fatal("unexpected result: ", disks)
	}

	d1 := disks[0]
	d2 := disks[1]

	if d1.Name() != "sdc" {
		t.Error(`d1.Name() != "sdc"`, d1.Name())
	}
	if d1.SectorSize() != 512 {
		t.Error(`d1.SectorSize() != 512`, d1.SectorSize())
	}
	if d1.Size512() != 2048 {
		t.Error(`d1.Size512() != 2048`, d1.Size512())
	}

	if d2.Name() != "sdd" {
		t.Error(`d2.Name() != "sdd"`, d2.Name())
	}
	if d2.SectorSize() != 4096 {
		t.Error(`d2.SectorSize() != 4096`, d2.SectorSize())
	}
	if d2.Size512() != 2048 {
		t.Error(`d2.Size512() != 2048`, d2.Size512())
	}
}
