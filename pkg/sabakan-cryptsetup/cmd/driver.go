package cmd

import (
	"context"
	"crypto/rand"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cybozu-go/log"
	sabakan "github.com/cybozu-go/sabakan/v2/client"
)

// Driver setup crypt devices.
type Driver struct {
	serial  string
	sabakan *sabakan.Client
	disks   []Disk
	cipher  string
	keySize int
}

// NewDriver creates Driver.
//
// It may return nil when the serial code of the machine cannot be identified,
// or sabakanURL is not valid.
func NewDriver(sabakanURL, cipher string, keySize int, disks []Disk) (*Driver, error) {
	hc := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          1,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	saba, err := sabakan.NewClient(sabakanURL, hc)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile("/sys/devices/virtual/dmi/id/product_serial")
	if err != nil {
		return nil, err
	}
	serial := strings.TrimSpace(string(data))

	return &Driver{
		serial:  serial,
		sabakan: saba,
		disks:   disks,
		cipher:  cipher,
		keySize: keySize,
	}, nil
}

// Setup setup crypt devices.
func (d *Driver) Setup(ctx context.Context) error {
	for _, disk := range d.disks {
		err := d.setupDisk(ctx, disk)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Driver) setupDisk(ctx context.Context, disk Disk) error {
	log.Info("setting up full disk encryption", map[string]interface{}{
		"disk": disk.Name(),
	})
	f, err := os.OpenFile(disk.Device(), os.O_RDWR, 0660)
	if err != nil {
		return err
	}
	defer f.Close()

	md, err := ReadMetadata(f)
	if err == ErrNotFound {
		log.Info("disk is not formatted", map[string]interface{}{
			"disk": disk.Name(),
		})
		return d.formatDisk(ctx, disk, f)
	}
	if err != nil {
		return err
	}

	ek, err := d.sabakan.CryptsGet(ctx, d.serial, md.HexID())
	if err == nil {
		log.Info("encryption key is found", map[string]interface{}{
			"disk": disk.Name(),
		})
		return Cryptsetup(disk, md, ek)
	}
	if sabakan.IsNotFound(err) {
		log.Info("encryption key is not found in sabakan", map[string]interface{}{
			"disk": disk.Name(),
		})
		return d.formatDisk(ctx, disk, f)
	}
	return err
}

func (d *Driver) formatDisk(ctx context.Context, disk Disk, f *os.File) error {
	md, err := NewMetadata(d.cipher, d.keySize)
	if err != nil {
		return err
	}

	key := make([]byte, d.keySize)
	_, err = rand.Read(key)
	if err != nil {
		return err
	}

	ek, err := md.EncryptKey(key)
	if err != nil {
		return err
	}

	err = Cryptsetup(disk, md, ek)
	if err != nil {
		return err
	}

	// write metadata before sending ek to sabakan.
	err = md.Write(f)
	if err != nil {
		return err
	}

	return d.sabakan.CryptsPut(ctx, d.serial, md.HexID(), ek)
}
