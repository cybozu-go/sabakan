package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/cybozu-go/cmd"
	"github.com/cybozu-go/sabakan/client"
)

type storageDevice struct {
	id             []byte
	byPath         string
	realPath       string
	key            []byte
	CommandContext func(context.Context, string, ...string) *cmd.LogCmd
}

const (
	cryptSetup = "/sbin/cryptsetup"
	cipher     = "aes-xts-plain64"
	keyBytes   = 64
	prefix     = "crypt"
	offset     = 4096 // keep the first 2 MiB for meta data.
	idBytes    = 16
)

var (
	magic = []byte{0x01, 0x02, 0x03, 0x04, 0xff, 0xfe, 0xfd, 0xfc, 0x00, 0x00, 0x00, 0x00, 0x12, 0x34, 0x78, 0x90,
		0x01, 0x02, 0x03, 0x04, 0xff, 0xfe, 0xfd, 0xfc, 0x00, 0x00, 0x00, 0x00, 0x12, 0x34, 0x78, 0x90,
		0x01, 0x02, 0x03, 0x04, 0xff, 0xfe, 0xfd, 0xfc, 0x00, 0x00, 0x00, 0x00, 0x12, 0x34, 0x78, 0x90,
		0x01, 0x02, 0x03, 0x04, 0xff, 0xfe, 0xfd, 0xfc, 0x00, 0x00, 0x00, 0x00, 0x12, 0x34, 0x78, 0x90}
	keySize  = keyBytes * 8
	idOffset = int64(len(magic) + keyBytes)
)

type devfsType struct {
	path string
}

var devfs = &devfsType{path: "/dev/disk/by-path"}

func (d devfsType) detectStorageDevices(ctx context.Context, patterns []string) ([]*storageDevice, error) {
	devices := make(map[string]*storageDevice)
	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(d.path, pattern))
		if err != nil {
			return nil, err
		}

		for _, device := range matches {
			// ignore partition device "*-part[0-9]+"
			base := filepath.Base(device)
			partition, err := regexp.MatchString("^.*-part[0-9]+$", base)
			if err != nil {
				return nil, err
			}
			if partition {
				continue
			}

			rp, err := filepath.EvalSymlinks(device)
			if err != nil {
				return nil, err
			}

			// ignore duplicated device
			if _, ok := devices[rp]; ok {
				continue
			}

			sd := &storageDevice{byPath: device, realPath: rp}
			err = sd.findID()
			if err != nil {
				return nil, err
			}
			devices[rp] = sd
		}
	}

	ret := make([]*storageDevice, 0, len(devices))
	for _, device := range devices {
		ret = append(ret, device)
	}
	return ret, nil
}

func (s *storageDevice) findID() error {
	f, err := os.Open(s.byPath)
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
		return nil
	}

	_, err = f.Seek(idOffset, 0)
	if err != nil {
		return err
	}

	buf := make([]byte, idBytes)
	_, err = io.ReadFull(f, buf)
	if err != nil {
		return err
	}

	s.id = buf
	return nil
}

func (s *storageDevice) idString() string {
	return hex.EncodeToString(s.id)
}

func (s *storageDevice) fetchKey(ctx context.Context, serial string) *client.Status {
	if s.id == nil {
		return client.NewStatus(client.ExitNotFound, errors.New("ID not found"))
	}
	data, status := client.CryptsGet(ctx, serial, s.idString())
	if status != nil {
		return status
	}
	s.key = data
	return nil
}

func (s *storageDevice) registerKey(ctx context.Context, serial string) *client.Status {
	return client.CryptsPut(ctx, serial, s.idString(), s.key)
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

	id := make([]byte, idBytes)
	_, err = rand.Read(id)
	if err != nil {
		return err
	}
	s.id = id

	f, err := os.OpenFile(s.byPath, os.O_RDWR, 0660)
	if err != nil {
		return err
	}
	defer f.Close()

	f.Write(magic)
	f.Write(opad)
	f.Write(id)
	f.Sync()

	xorKey := make([]byte, keyBytes)
	for i := 0; i < keyBytes; i++ {
		xorKey[i] = opad[i] ^ key[i]
	}
	s.key = xorKey

	return nil
}

func (s *storageDevice) decrypt(ctx context.Context) error {
	f, err := os.Open(s.byPath)
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
		return errors.New("Non-formatted device " + s.byPath)
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

	if s.CommandContext == nil {
		s.CommandContext = cmd.CommandContext
	}

	cryptName := fmt.Sprintf("%s-%s-%s", prefix, filepath.Base(s.realPath), s.idString())
	c := s.CommandContext(ctx, cryptSetup, "--hash=plain", "--key-file=-",
		"--cipher="+cipher, "--key-size="+strconv.Itoa(keySize), "--offset="+strconv.Itoa(offset),
		"--allow-discards", "open", s.byPath, "--type=plain", cryptName)
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
