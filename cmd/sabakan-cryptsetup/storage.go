package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/cybozu-go/cmd"
	"github.com/cybozu-go/sabakan/client"
)

type storageDevice struct {
	partUUID string
	byPath   string
	realPath string
	key      []byte
}

const (
	cryptSetup    = "/sbin/cryptsetup"
	gdisk         = "/sbin/gdisk"
	gdiskCommands = "o\nY\nn\n\n\n\n\nw\nY\n" // make partition using whole disk
	cipher        = "aes-xts-plain64"
	keyBytes      = 64
	prefix        = "crypt-"
	offset        = 4096 // keep the first 2 MiB for meta data.
)

var (
	magic = []byte{0x01, 0x02, 0x03, 0x04, 0xff, 0xfe, 0xfd, 0xfc, 0x00, 0x00, 0x00, 0x00, 0x12, 0x34, 0x78, 0x90,
		0x01, 0x02, 0x03, 0x04, 0xff, 0xfe, 0xfd, 0xfc, 0x00, 0x00, 0x00, 0x00, 0x12, 0x34, 0x78, 0x90,
		0x01, 0x02, 0x03, 0x04, 0xff, 0xfe, 0xfd, 0xfc, 0x00, 0x00, 0x00, 0x00, 0x12, 0x34, 0x78, 0x90,
		0x01, 0x02, 0x03, 0x04, 0xff, 0xfe, 0xfd, 0xfc, 0x00, 0x00, 0x00, 0x00, 0x12, 0x34, 0x78, 0x90}
	keySize = keyBytes * 8
)

func (s *storageDevice) partUUIDPath() string {
	if s.partUUID == "" {
		return ""
	}
	return filepath.Join("/dev/disk/by-partuuid", s.partUUID)
}

func detectStorageDevices(ctx context.Context, patterns []string) ([]*storageDevice, error) {
	devices := make(map[string]*storageDevice)
	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join("/dev/disk/by-path", pattern))
		if err != nil {
			return nil, err
		}

		for _, device := range matches {
			base := filepath.Base(device)
			// ignore partition device "*-part[0-9]+"
			partition, err := regexp.MatchString("^.*-part[0-9]+$", base)
			if err != nil {
				return nil, err
			}
			if partition {
				continue
			}

			// ignore duplicated device
			if _, ok := devices[device]; ok {
				continue
			}

			rp, err := filepath.EvalSymlinks(device)
			if err != nil {
				return nil, err
			}
			sd := &storageDevice{byPath: device, realPath: rp}
			err = sd.findPartUUID()
			if err != nil {
				return nil, err
			}
			devices[device] = sd
		}
	}

	ret := make([]*storageDevice, 0, len(devices))
	for _, device := range devices {
		ret = append(ret, device)
	}
	return ret, nil
}

func (s *storageDevice) findPartUUID() error {
	partUUIDList, err := filepath.Glob("/dev/disk/by-partuuid/*")
	if err != nil {
		return err
	}
	for _, partUUID := range partUUIDList {
		partitionRealPath, err := filepath.EvalSymlinks(partUUID)
		if err != nil {
			return err
		}
		if s.realPath+"1" == partitionRealPath {
			s.partUUID = filepath.Base(partUUID)
			return nil
		}
	}
	s.partUUID = ""
	return nil
}

func (s *storageDevice) fetchKey(ctx context.Context, serial string) *client.Status {
	if len(s.partUUID) == 0 {
		return client.NewStatus(client.ExitNotFound, errors.New("partition not found"))
	}
	data, status := client.CryptsGet(ctx, serial, s.partUUID)
	if status != nil {
		return status
	}
	s.key = data
	return nil
}

func (s *storageDevice) registerKey(ctx context.Context, serial string) *client.Status {
	return client.CryptsPut(ctx, serial, s.partUUID, s.key)
}

func (s *storageDevice) makePartition(ctx context.Context) error {
	c := cmd.CommandContext(ctx, gdisk, s.realPath)
	pipe, err := c.StdinPipe()
	if err != nil {
		return err
	}

	go func() error {
		defer pipe.Close()
		_, err := io.WriteString(pipe, gdiskCommands)
		if err != nil {
			return err
		}
		return nil
	}()

	err = c.Run()
	if err != nil {
		return err
	}
	for {
		err = s.findPartUUID()
		if err != nil {
			return err
		}
		if s.partUUID != "" {
			return nil
		}
		select {
		case <-time.After(time.Duration(500) * time.Millisecond):
		case <-ctx.Done():
			return errors.New("failed to get UUID of the partition")
		}
	}
}

// encrypt the disk, then set properties (d.key)
func (s *storageDevice) encrypt(ctx context.Context) error {
	opad := make([]byte, keyBytes)
	_, err := rand.Read(opad)
	if err != nil {
		return err
	}

	key := make([]byte, keyBytes)
	_, err = rand.Read(key)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(s.partUUIDPath(), os.O_RDWR, 0660)
	if err != nil {
		return err
	}
	defer f.Close()

	f.Write(magic)
	f.Write(opad)
	f.Sync()

	xorKey := make([]byte, keyBytes)
	for i := 0; i < keyBytes; i++ {
		xorKey[i] = opad[i] ^ key[i]
	}
	s.key = xorKey

	return nil
}

func (s *storageDevice) decrypt(ctx context.Context) error {
	f, err := os.Open(s.partUUIDPath())
	if err != nil {
		return err
	}
	defer f.Close()

	m := make([]byte, len(magic))
	_, err = io.ReadFull(f, m)
	if err != nil {
		return err
	}

	if !bytes.Equal(m, magic) {
		return errors.New("Non-formatted device " + s.partUUIDPath())
	}

	opad := make([]byte, keyBytes)
	_, err = io.ReadFull(f, opad)
	if err != nil {
		return err
	}

	xorKey := make([]byte, keyBytes)
	for i := 0; i < keyBytes; i++ {
		xorKey[i] = opad[i] ^ s.key[i]
	}

	c := cmd.CommandContext(ctx, cryptSetup, "--hash=plain", "--key-file=-",
		"--cipher="+cipher, "--key-size="+strconv.Itoa(keySize), "--offset="+strconv.Itoa(offset),
		"--allow-discards", "open", s.partUUIDPath(), "--type=plain", prefix+s.partUUID)
	pipe, err := c.StdinPipe()
	if err != nil {
		return err
	}

	go func() error {
		defer pipe.Close()
		_, err := pipe.Write(xorKey)
		if err != nil {
			return err
		}
		return nil
	}()

	err = c.Run()
	if err != nil {
		return err
	}

	return nil
}
