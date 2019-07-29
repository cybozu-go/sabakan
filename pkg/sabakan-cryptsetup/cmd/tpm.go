package cmd

import (
	"crypto/rand"
	"io"

	"github.com/cybozu-go/log"
	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"
)

// The value "0x105" represents Manufacturer of a TPM Properties defined below:
// https://github.com/google/go-tpm/blob/d6d17943421ff5e8991df2cea58480079d3a3c36/tpm2/constants.go#L168
const manufacturer = 0x105

const (
	tpmKekLength = 64
	tpmOffsetHex = 0x01000000
)

var tpmOffset = tpmutil.Handle(tpmOffsetHex)

type tpmDriver struct {
	device io.ReadWriteCloser
}

func newTPMDriver(device string) (*tpmDriver, error) {
	rw, err := tpm2.OpenTPM(device)
	if err != nil {
		return nil, err
	}

	return &tpmDriver{
		device: rw,
	}, nil
}

func (t *tpmDriver) checkTPMVersion20() error {
	// Make sure this is a TPM 2.0
	// https://github.com/google/go-tpm/blob/30f8389f7afbbd553e969bf7c59c54e0a83a3373/tpm2/open_other.go#L35-L40
	caps, _, err := tpm2.GetCapability(t.device, tpm2.CapabilityTPMProperties, 1, uint32(manufacturer))
	if err != nil {
		return err
	}

	prop := caps[0].(tpm2.TaggedProperty)
	_, err = tpmutil.Pack(prop.Value)
	if err != nil {
		return err
	}

	return nil
}

func (t *tpmDriver) readKEKFromTPM() ([]byte, error) {
	return tpm2.NVReadEx(t.device, tpmOffset, tpm2.HandleOwner, "", 0)
}

func (t *tpmDriver) allocateNVRAM() error {
	err := t.defineNVSpace()
	if err != nil {
		e, ok := err.(tpm2.Error)
		if !ok {
			return err
		}
		if e.Code != tpm2.RCNVRange {
			log.Warn("out of range key encryption key, so re-define NV space", map[string]interface{}{
				log.FnError: err,
			})
			err := t.undefineNVSpace()
			if err != nil {
				return err
			}
			err = t.defineNVSpace()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *tpmDriver) defineNVSpace() error {
	attr := tpm2.AttrOwnerWrite | tpm2.AttrOwnerRead
	err := tpm2.NVDefineSpace(
		t.device,
		tpm2.HandleOwner,
		tpmOffset,
		"",
		"",
		nil,
		attr,
		uint16(tpmKekLength),
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

	// Prepare encryption key
	kek := make([]byte, tpmKekLength)
	_, err = rand.Read(kek)
	if err != nil {
		return err
	}

	return tpm2.NVWrite(t.device, tpm2.HandleOwner, tpmOffset, "", kek, 0)
}

func (t *tpmDriver) undefineNVSpace() error {
	return tpm2.NVUndefineSpace(t.device, "", tpm2.HandleOwner, tpmOffset)
}
