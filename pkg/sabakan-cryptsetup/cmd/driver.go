package cmd

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cybozu-go/log"
	sabakan "github.com/cybozu-go/sabakan/v3/client"
)

const maxRetry = 10

// Driver setup crypt devices.
type Driver struct {
	serial  string
	sabakan *sabakan.Client
	disks   []Disk
	cipher  string
	keySize int
	tpmdev  string

	// status variables
	tpmVersion TpmVersionID
}

// NewDriver creates Driver.
//
// It may return nil when the serial code of the machine cannot be identified,
// or sabakanURL is not valid.
func NewDriver(sabakanURL, cipher string, keySize int, tpmdev string, disks []Disk) (*Driver, error) {
	crt, err := os.ReadFile(opts.caCert)
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(crt)

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
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}
	saba, err := sabakan.NewClient(sabakanURL, hc)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile("/sys/devices/virtual/dmi/id/product_serial")
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

		tpmVersion: TpmNone,
	}, nil
}

// Setup setup crypt devices.
func (d *Driver) Setup(ctx context.Context) error {
	var kek []byte

	_, err := os.Stat(d.tpmdev)
	switch {
	case err == nil:
		log.Info("TPM is found. disk encryption proceeds with TPM", map[string]interface{}{
			"device": d.tpmdev,
		})
		kek, d.tpmVersion, err = readKeyFromTPM(d.tpmdev)
		if err != nil {
			return err
		}
	case os.IsNotExist(err):
		log.Info("no TPM is found. disk encryption proceeds without TPM", map[string]interface{}{
			"device":    d.tpmdev,
			log.FnError: err,
		})
	default:
		return err
	}

	for _, disk := range d.disks {
		err := d.setupDisk(ctx, disk, kek)
		if err != nil {
			return err
		}
		log.Info("encrypted disk is ready", map[string]interface{}{
			"path": filepath.Join("/dev/mapper", disk.CryptName()),
		})
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
		log.Info("disk is not formatted. format disk", map[string]interface{}{
			"disk": disk.Name(),
		})
		return d.formatDisk(ctx, disk, f, tpmKek)
	}
	if err != nil {
		return err
	}
	if md.tpmVersion != d.tpmVersion {
		if d.tpmVersion == 0 {
			log.Error("TPM becomes unavailable", map[string]interface{}{
				"tpmversion": md.tpmVersion.String(),
			})
			return errors.New("TPM unavailable")
		}
		log.Info("reformat disk because TPM is now available", map[string]interface{}{
			"disk":       disk.Name(),
			"tpmversion": d.tpmVersion.String(),
		})
		return d.formatDisk(ctx, disk, f, tpmKek)
	}

	var retries int
RETRY:
	ek, err := d.sabakan.CryptsGet(ctx, d.serial, md.HexID())
	if err == nil {
		log.Info("encryption key is found. run cryptsetup", map[string]interface{}{
			"disk": disk.Name(),
		})
		return Cryptsetup(disk, md, ek, tpmKek)
	}
	if sabakan.IsNotFound(err) {
		log.Info("encryption key is not found in sabakan. format disk", map[string]interface{}{
			"disk": disk.Name(),
		})
		return d.formatDisk(ctx, disk, f, tpmKek)
	}

	log.Error("failed to retrieve key from sabakan", map[string]interface{}{
		log.FnError: err,
		"disk":      disk.Name(),
		"try":       retries + 1,
	})
	if retries == maxRetry {
		return err
	}
	retries++
	time.Sleep(time.Duration(retries) * time.Second * 2)
	goto RETRY
}

func (d *Driver) formatDisk(ctx context.Context, disk Disk, f *os.File, tpmKek []byte) error {
	md, err := NewMetadata(d.cipher, d.keySize, d.tpmVersion)
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

	var retries int
RETRY:
	err = d.sabakan.CryptsPut(ctx, d.serial, md.HexID(), ek)
	if err == nil {
		return nil
	}
	log.Error("failed to send key to sabakan", map[string]interface{}{
		log.FnError: err,
		"disk":      disk.Name(),
		"try":       retries + 1,
	})
	if retries == maxRetry {
		return err
	}
	retries++
	time.Sleep(time.Duration(retries) * time.Second * 2)
	goto RETRY
}
