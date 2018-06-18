package main

import (
	"context"
	"path/filepath"
	"regexp"

	"github.com/cybozu-go/sabakan/client"
)

type storageDevice struct {
	path string
	key  []byte
}

func detectStorageDevices(ctx context.Context, patterns []string) ([]*storageDevice, error) {
	devices := make(map[string]*storageDevice)
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

			devices[device] = &storageDevice{path: base}
		}
	}

	ret := make([]*storageDevice, 0, len(devices))
	for _, device := range devices {
		ret = append(ret, device)
	}
	return ret, nil
}

func (s *storageDevice) fetchKey(ctx context.Context, serial string) *client.Status {
	data, status := client.CryptsGet(ctx, serial, s.path)
	if status != nil {
		return status
	}
	s.key = data
	return nil
}

func (s *storageDevice) registerKey(ctx context.Context, serial string) *client.Status {
	return client.CryptsPut(ctx, serial, s.path, s.key)
}

// encrypt the disk, then set properties (d.key)
func (s *storageDevice) encrypt(ctx context.Context) error {
	return nil
}

func (s *storageDevice) decrypt(ctx context.Context) error {
	return nil
}
