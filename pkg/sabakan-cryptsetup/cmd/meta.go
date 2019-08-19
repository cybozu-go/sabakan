package cmd

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
)

const (
	magicBytes3    = "\x80sabakan-cryptsetup3"
	magicBytes2    = "\x80sabakan-cryptsetup2"
	maxCipherName3 = 105
	maxCipherName2 = 106
	idLength       = 16
	metadataSize   = 2 * 1024 * 1024
)

// Pre-defined errors
var (
	ErrNotFound = errors.New("not found")
)

// TpmVersionID represents TPM versions.
type TpmVersionID int

// TPM versions.
const (
	TpmNone TpmVersionID = 0
	Tpm12                = 1
	Tpm20                = 2
)

func (v TpmVersionID) String() string {
	switch v {
	case TpmNone:
		return "None"
	case Tpm12:
		return "1.2"
	case Tpm20:
		return "2.0"
	}
	return ""
}

// Metadata represents metadata block at the head of disk.
type Metadata struct {
	cipher     string
	id         string
	kek        string
	tpmVersion TpmVersionID
}

// ReadMetadata read metadata from f.
// If metadata does not exist, this returns ErrNotFound.
func ReadMetadata(f *os.File) (*Metadata, error) {
	data := make([]byte, metadataSize)
	_, err := f.ReadAt(data, 0)
	if err != nil {
		return nil, err
	}

	if bytes.HasPrefix(data, []byte(magicBytes2)) {
		return convertMetadata2(f, data)
	}

	if !bytes.HasPrefix(data, []byte(magicBytes3)) {
		return nil, ErrNotFound
	}

	keySize := int(data[20])
	tpmVersion := TpmVersionID(data[21])
	cnl := int(data[22])
	if cnl > maxCipherName3 {
		return nil, fmt.Errorf("cipher name too long: %d", cnl)
	}

	md := &Metadata{
		cipher:     string(data[23:(23 + cnl)]),
		id:         string(data[128:144]),
		kek:        string(data[144:(144 + keySize)]),
		tpmVersion: tpmVersion,
	}
	return md, nil
}

// NewMetadata initializes a new Metadata.
func NewMetadata(cipher string, keySize int, tpmVersion TpmVersionID) (*Metadata, error) {
	if len(cipher) > maxCipherName3 {
		return nil, errors.New("too long cipher name")
	}
	if keySize > 255 {
		return nil, errors.New("too large key size")
	}

	id := make([]byte, idLength)
	_, err := rand.Read(id)
	if err != nil {
		return nil, err
	}
	kek := make([]byte, keySize)
	_, err = rand.Read(kek)
	if err != nil {
		return nil, err
	}

	md := &Metadata{
		cipher:     cipher,
		id:         string(id),
		kek:        string(kek),
		tpmVersion: tpmVersion,
	}
	return md, nil
}

// Write writes metadata to f.
func (m *Metadata) Write(f *os.File) error {
	if len(m.cipher) > maxCipherName3 {
		return errors.New("too long cipher name: " + m.cipher)
	}
	if len(m.id) != idLength {
		return errors.New("invalid id length")
	}

	data := bytes.Repeat([]byte{'\x88'}, metadataSize)
	copy(data, magicBytes3)
	data[20] = byte(len(m.kek))
	data[21] = byte(m.tpmVersion)
	data[22] = byte(len(m.cipher))
	copy(data[23:(23+len(m.cipher))], m.cipher)
	copy(data[128:144], m.id)
	copy(data[144:], m.kek)

	_, err := f.WriteAt(data, 0)
	if err != nil {
		return err
	}
	return f.Sync()
}

// Cipher returns cipher suite for this disk.
func (m *Metadata) Cipher() string {
	return m.cipher
}

// ID returns randomly assigned ID of this disk.
func (m *Metadata) ID() string {
	return m.id
}

// HexID returns hexadecimal encoded ID.
func (m *Metadata) HexID() string {
	return hex.EncodeToString([]byte(m.id))
}

// Kek returns key encryption key.
func (m *Metadata) Kek() string {
	return m.kek
}

// TpmVersion returns TPM version ID.
func (m *Metadata) TpmVersion() TpmVersionID {
	return m.tpmVersion
}

// DecryptKey decrypts encrypted key.
func (m *Metadata) DecryptKey(ek, tpmKek []byte) ([]byte, error) {
	if len(ek) != len(m.kek) {
		return nil, fmt.Errorf("key length mismatch: expected=%d, actual=%d", len(m.kek), len(ek))
	}
	if len(tpmKek) != 0 && len(ek) > len(tpmKek) {
		return nil, fmt.Errorf("TPM key is too short: required=%d, actual=%d", len(ek), len(tpmKek))
	}

	key := make([]byte, len(ek))
	for i := range ek {
		if len(tpmKek) != 0 {
			key[i] = m.kek[i] ^ ek[i] ^ tpmKek[i]
		} else {
			key[i] = m.kek[i] ^ ek[i]
		}
	}
	return key, nil
}

// EncryptKey encrypts key.
func (m *Metadata) EncryptKey(key, tpmKek []byte) ([]byte, error) {
	return m.DecryptKey(key, tpmKek)
}

func convertMetadata2(f *os.File, data []byte) (*Metadata, error) {
	keySize := int(data[20])
	cnl := int(data[21])
	if cnl > maxCipherName3 {
		return nil, fmt.Errorf("cipher name too long: %d", cnl)
	}

	md := &Metadata{
		cipher:     string(data[22:(22 + cnl)]),
		id:         string(data[128:144]),
		kek:        string(data[144:(144 + keySize)]),
		tpmVersion: TpmNone,
	}
	if err := md.Write(f); err != nil {
		return nil, err
	}
	return md, nil
}
