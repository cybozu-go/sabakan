package main

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

const (
	backingFileSize = 3 * 1024 * 1024 // >2MiB (metadata size)
)

func createPseudoDevice(t *testing.T, dir, backing, device string, deviceMap map[string]*storageDevice) {
	f, err := os.Create(filepath.Join(dir, backing))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	err = f.Truncate(backingFileSize)
	if err != nil {
		t.Fatal(err)
	}

	devDir := filepath.Join(dir, "devices")
	err = os.MkdirAll(devDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	devPath := filepath.Join(devDir, device)
	err = os.Symlink(f.Name(), devPath)
	if err != nil {
		t.Fatal(err)
	}

	deviceMap[device] = &storageDevice{
		byPath:   devPath,
		realPath: f.Name(),
	}
}

func setupTestStorage(t *testing.T, dir string) (map[string]*storageDevice, *devfsType) {
	deviceMap := make(map[string]*storageDevice)

	createPseudoDevice(t, dir, "f1", "nvme-1", deviceMap)
	createPseudoDevice(t, dir, "f1", "nvme-1-dup", deviceMap)
	createPseudoDevice(t, dir, "f2", "nvme-1-part1", deviceMap)
	createPseudoDevice(t, dir, "f3", "nvme-2", deviceMap)
	createPseudoDevice(t, dir, "f4", "sata-1", deviceMap)
	createPseudoDevice(t, dir, "f5", "sata-2", deviceMap)

	testDevfs := &devfsType{path: filepath.Join(dir, "devices")}

	return deviceMap, testDevfs
}

func sameDevices(x, y []*storageDevice) bool {
	xMap := make(map[string]*storageDevice)
	for _, d := range x {
		xMap[d.realPath] = d
	}

	yMap := make(map[string]*storageDevice)
	for _, d := range y {
		yMap[d.realPath] = d
	}

	return reflect.DeepEqual(xMap, yMap)
}

func testDetectStorageDevices(t *testing.T) {
	t.Parallel()

	d, err := ioutil.TempDir("", "sabakan-cryptsetup-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(d)

	deviceMap, testDevfs := setupTestStorage(t, d)

	devices, err := testDevfs.detectStorageDevices(context.Background(), []string{"nvme-*", "*-1"})
	if err != nil {
		t.Fatal(err)
	}
	expected := []*storageDevice{deviceMap["nvme-1"], deviceMap["nvme-2"], deviceMap["sata-1"]}
	if !sameDevices(devices, expected) {
		t.Error("detected wrong storage devices")
		for _, device := range devices {
			t.Log(device)
		}
	}
}

func testEncrypt(t *testing.T) {
	t.Parallel()

	d, err := ioutil.TempDir("", "sabakan-cryptsetup-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(d)

	deviceMap, _ := setupTestStorage(t, d)

	device := deviceMap["nvme-1"]
	err = device.encrypt(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if device.id == nil {
		t.Error("device ID not set")
	}
	if device.key == nil {
		t.Error("device key not set")
	}

	f, err := os.Open(device.byPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	readMagic := make([]byte, len(magic))
	_, err = io.ReadFull(f, readMagic)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(readMagic, magic) {
		t.Error("invalid magic:", readMagic)
	}

	readID := make([]byte, idBytes)
	_, err = f.Seek(idOffset, 0)
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.ReadFull(f, readID)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(readID, device.id) {
		t.Error("ID mismatch", readID, device.id)
	}
}

func TestStorage(t *testing.T) {
	t.Run("Detect", testDetectStorageDevices)
	t.Run("Encrypt", testEncrypt)
}
