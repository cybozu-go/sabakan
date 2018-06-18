package main

import (
	"context"
	"path/filepath"
	"regexp"
)

type StorageDevice struct {
	path string
	key  []byte
}

func detectStorageDevices(ctx context.Context, patterns []string) ([]*StorageDevice, error) {
	devices := make(map[string]*StorageDevice)
	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join("/dev/disk/by-path", pattern))
		if err != nil {
			return nil, err
		}

		for _, device := range matches {
			// ignore partition device "*-part[0-9]+"
			partition, err := regexp.MatchString("^/dev/disk/by-path/.*-part[0-9]+$", device)
			if err != nil {
				return nil, err
			}
			if partition {
				continue
			}

			base := filepath.Base(device)

			// ignore duplicated device
			if _, ok := devices[device]; ok {
				continue
			}

			devices[device] = &StorageDevice{path: base}
		}
	}

	ret := make([]*StorageDevice, 0, len(devices))
	for _, device := range devices {
		ret = append(ret, device)
	}
	return ret, nil
}
