package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Disk represents a physical disk to be encrypted.
type Disk struct {
	name       string
	sectorSize int
	size512    int64
}

// FindDisks looks up the system to find disks to be encrypted.
func FindDisks(excludes []string) ([]Disk, error) {
	return findDisks(excludes, "/sys/block")
}

func findDisks(excludes []string, base string) ([]Disk, error) {
	match := func(name string) (bool, error) {
		for _, pat := range excludes {
			ok, err := filepath.Match(pat, name)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
		return false, nil
	}
	// since only possible error is filepath.ErrBadPattern,
	// check that first.
	_, err := match("")
	if err != nil {
		return nil, err
	}

	hasFlag := func(name, flag string) (bool, error) {
		data, err := ioutil.ReadFile(filepath.Join(base, name, flag))
		if err != nil {
			if os.IsNotExist(err) {
				return false, nil
			}
			return false, err
		}
		if len(data) == 0 {
			return false, nil
		}
		return data[0] != '0', nil
	}

	readInt64 := func(name, flag string) (int64, error) {
		data, err := ioutil.ReadFile(filepath.Join(base, name, flag))
		if err != nil {
			return 0, err
		}
		return strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	}

	// non-virtual disks have "device" link.
	diskPaths, err := filepath.Glob(filepath.Join(base, "*", "device"))
	if err != nil {
		return nil, err
	}
	disks := make([]Disk, 0, len(diskPaths))
	for _, p := range diskPaths {
		name := filepath.Base(filepath.Dir(p))
		ok, _ := match(name)
		if ok {
			continue
		}
		ok, err = hasFlag(name, "removable")
		if err != nil {
			return nil, err
		}
		if ok {
			continue
		}
		ok, err = hasFlag(name, "ro")
		if err != nil {
			return nil, err
		}
		if ok {
			continue
		}

		// Note: using "hw_sector_size" is WRONG!
		// See https://github.com/ansible/ansible/pull/7740
		sectorSize, err := readInt64(name, "queue/physical_block_size")
		if err != nil {
			return nil, err
		}
		size512, err := readInt64(name, "size")
		if err != nil {
			return nil, err
		}
		disks = append(disks, Disk{
			name:       name,
			sectorSize: int(sectorSize),
			size512:    size512,
		})
	}

	return disks, nil
}

// Device returns a device filename of this disk.
func (d Disk) Device() string {
	return filepath.Join("/dev", d.name)
}

// CryptName returns the crypt device name of this disk.
func (d Disk) CryptName() string {
	return "crypt-" + d.name
}

// CryptDevice returns the crypt device filename of this disk.
func (d Disk) CryptDevice() string {
	return filepath.Join("/dev/mapper", d.CryptName())
}

// Name returns the name of this disk.
func (d Disk) Name() string {
	return d.name
}

// SectorSize returns the physical block size of this disk.
func (d Disk) SectorSize() int {
	return d.sectorSize
}

// Size512 returns the device size / 512.
func (d Disk) Size512() int64 {
	return d.size512
}
