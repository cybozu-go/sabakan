package cmd

import (
	"crypto/rand"
	"io"

	"github.com/cybozu-go/log"
	"github.com/google/go-tpm/legacy/tpm2"
	"github.com/google/go-tpm/tpm"
	"github.com/google/go-tpm/tpmutil"
)

const (
	tpmKekLength = 256
	tpmOffsetHex = 0x01000000
)

var tpmOffset = tpmutil.Handle(tpmOffsetHex)

type tpmDriver struct {
	io.ReadWriteCloser
}

func (t *tpmDriver) readKEKFromTPM() ([]byte, error) {
	return tpm2.NVReadEx(t, tpmOffset, tpm2.HandleOwner, "", 0)
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
		t,
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

	return tpm2.NVWrite(t, tpm2.HandleOwner, tpmOffset, "", kek, 0)
}

func (t *tpmDriver) undefineNVSpace() error {
	return tpm2.NVUndefineSpace(t, "", tpm2.HandleOwner, tpmOffset)
}

func readKeyFromTPM(device string) ([]byte, TpmVersionID, error) {
	rw, err := tpm2.OpenTPM(device)
	if err != nil {
		t2, err2 := tpm.OpenTPM(device)
		if err2 == nil {
			t2.Close()
			log.Warn("device is not TPM 2.0. disk encryption proceeds without TPM", map[string]interface{}{
				"device": device,
			})
			return nil, TpmNone, nil
		}
		return nil, TpmNone, err
	}

	t := &tpmDriver{rw}
	defer t.Close()

	kek, err := t.readKEKFromTPM()
	if err == nil {
		return kek, Tpm20, nil
	}

	log.Info("TPM key encryption key was not found", map[string]interface{}{
		log.FnError: err,
	})
	err = t.allocateNVRAM()
	if err != nil {
		return nil, TpmNone, err
	}
	kek, err = t.readKEKFromTPM()
	if err != nil {
		panic(err)
	}
	return kek, Tpm20, nil
}
