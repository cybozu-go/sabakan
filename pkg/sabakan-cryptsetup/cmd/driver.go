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
	tpmdev  string
}

// NewDriver creates Driver.
//
// It may return nil when the serial code of the machine cannot be identified,
// or sabakanURL is not valid.
func NewDriver(sabakanURL, cipher string, keySize int, tpmdev string, disks []Disk) (*Driver, error) {
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
		tpmdev:  tpmdev,
	}, nil
}

// Setup setup crypt devices.
func (d *Driver) Setup(ctx context.Context) error {
	kek := []byte("")

	_, err := os.Stat(d.tpmdev)
	switch {
	case !os.IsNotExist(err):
		return err
	case os.IsNotExist(err):
		log.Info("no TPM is found. disk encryption proceeds without TPM", map[string]interface{}{
			"device":    d.tpmdev,
			log.FnError: err,
		})
	default:
		t, err := newTPMDriver(d.tpmdev)
		if err != nil {
			return err
		}
		defer t.device.Close()

		err = t.checkTPMVersion20()
		if err != nil {
			log.Warn("device is not TPM 2.0. disk encryption proceeds without TPM", map[string]interface{}{
				"device":    d.tpmdev,
				log.FnError: err,
			})
		}

		kek, err = t.getKEKFromTPM()
		if err != nil {
			log.Info("no TPM key encryption key was found", map[string]interface{}{
				log.FnError: err,
			})
			err := t.allocateNVRAM()
			if err != nil {
				return err
			}
			kek, err = t.getKEKFromTPM()
			if err != nil {
				panic(err)
			}
		}
	}

	for _, disk := range d.disks {
		err := d.setupDisk(ctx, disk, kek)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Driver) setupDisk(ctx context.Context, disk Disk, tpmKek []byte) error {
	log.Info("setting up full disk encryption", map[string]interface{}{
		"disk":           disk.Name(),
		"tpm_kek_length": len(tpmKek),
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
		return d.formatDisk(ctx, disk, f, tpmKek)
	}
	if err != nil {
		return err
	}

	ek, err := d.sabakan.CryptsGet(ctx, d.serial, md.HexID())
	if err == nil {
		log.Info("encryption key is found", map[string]interface{}{
			"disk": disk.Name(),
		})
		return Cryptsetup(disk, md, ek, tpmKek)
	}
	if sabakan.IsNotFound(err) {
		log.Info("encryption key is not found in sabakan", map[string]interface{}{
			"disk": disk.Name(),
		})
		return d.formatDisk(ctx, disk, f, tpmKek)
	}
	return err
}

func (d *Driver) formatDisk(ctx context.Context, disk Disk, f *os.File, tpmKek []byte) error {
	md, err := NewMetadata(d.cipher, d.keySize)
	if err != nil {
		return err
	}

	key := make([]byte, d.keySize)
	_, err = rand.Read(key)
	if err != nil {
		return err
	}

	ek, err := md.EncryptKey(key, tpmKek)
	if err != nil {
		return err
	}

	err = Cryptsetup(disk, md, ek, tpmKek)
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
