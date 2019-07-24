package cmd

import (
	"context"
	"crypto/rand"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cybozu-go/log"
	sabakan "github.com/cybozu-go/sabakan/v2/client"
	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"
)

// The value "0x105" represents Manufacturer of a TPM Properties defined below:
// https://github.com/google/go-tpm/blob/d6d17943421ff5e8991df2cea58480079d3a3c36/tpm2/constants.go#L168
const manufacturer = 0x105

const (
	tpmKeyLength = 64
	tpmOffsetHex = 0x01000000
)

var tpmOffset = tpmutil.Handle(tpmOffsetHex)

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
	err := d.allocateNVRAM(ctx)
	if err != nil {
		return err
	}

	for _, disk := range d.disks {
		err := d.setupDisk(ctx, disk)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Driver) allocateNVRAM(ctx context.Context) error {
	rw, err := tpm2.OpenTPM(d.tpmdev)
	if err != nil {
		return err
	}
	defer rw.Close()

	// Make sure this is a TPM 2.0
	// https://github.com/google/go-tpm/blob/30f8389f7afbbd553e969bf7c59c54e0a83a3373/tpm2/open_other.go#L35-L40
	caps, _, err := tpm2.GetCapability(rw, tpm2.CapabilityTPMProperties, 1, uint32(manufacturer))
	if err != nil {
		return err
	}

	prop := caps[0].(tpm2.TaggedProperty)
	_, err = tpmutil.Pack(prop.Value)
	if err != nil {
		return err
	}

	// Prepare encryption key
	kek := make([]byte, tpmKeyLength)
	_, err = rand.Read(kek)
	if err != nil {
		return err
	}
	err = defineNVSpace(rw, kek)
	if err != nil {
		return err
	}

	err = tpm2.NVWrite(rw, tpm2.HandleOwner, tpmOffset, "", kek, 0)
	if err != nil {
		e, ok := err.(tpm2.Error)
		if !ok {
			return err
		}
		if e.Code != tpm2.RCNVRange {
			err := undefineNVSpace(rw)
			if err != nil {
				return err
			}
			err = defineNVSpace(rw, kek)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func defineNVSpace(rw io.ReadWriteCloser, kek []byte) error {
	attr := tpm2.AttrOwnerWrite | tpm2.AttrOwnerRead
	err := tpm2.NVDefineSpace(
		rw,
		tpm2.HandleOwner,
		tpmOffset,
		"",
		"",
		nil,
		attr,
		uint16(len(kek)),
	)
	if err != nil {
		e, ok := err.(tpm2.Error)
		if !ok {
			return err
		}
		if e.Code != tpm2.RCNVDefined {
			return err
		}
	}
	return nil
}

func undefineNVSpace(rw io.ReadWriteCloser) error {
	return tpm2.NVUndefineSpace(rw, "", tpm2.HandleOwner, tpmOffset)
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
