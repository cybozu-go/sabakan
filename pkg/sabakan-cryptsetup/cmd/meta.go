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
	magicBytes    = "\x80sabakan-cryptsetup2"
	maxCipherName = 106
	idLength      = 16
	metadataSize  = 2 * 1024 * 1024
)

// Pre-defined errors
var (
	ErrNotFound = errors.New("not found")
)

// Metadata represents metadata block at the head of disk.
type Metadata struct {
	cipher string
	id     string
	kek    string
}

// ReadMetadata read metadata from f.
// If metadata does not exist, this returns ErrNotFound.
func ReadMetadata(f *os.File) (*Metadata, error) {
	data := make([]byte, metadataSize)
	_, err := f.ReadAt(data, 0)
	if err != nil {
		return nil, err
	}

	if string(data[0:len(magicBytes)]) != magicBytes {
		return nil, ErrNotFound
	}

	keySize := int(data[20])
	cnl := int(data[21])
	if cnl > maxCipherName {
		return nil, fmt.Errorf("cipher name too long: %d", cnl)
	}

	md := &Metadata{
		cipher: string(data[22:(22 + cnl)]),
		id:     string(data[128:144]),
		kek:    string(data[144:(144 + keySize)]),
	}
	return md, nil
}

// NewMetadata initializes a new Metadata.
func NewMetadata(cipher string, keySize int) (*Metadata, error) {
	if len(cipher) > maxCipherName {
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
		cipher: cipher,
		id:     string(id),
		kek:    string(kek),
	}
	return md, nil
}

// Write writes metadata to f.
func (m *Metadata) Write(f *os.File) error {
	if len(m.cipher) > maxCipherName {
		return errors.New("too long cipher name: " + m.cipher)
	}
	if len(m.id) != idLength {
		return errors.New("invalid id length")
	}

	data := bytes.Repeat([]byte{'\x88'}, metadataSize)
	copy(data, magicBytes)
	data[20] = byte(len(m.kek))
	data[21] = byte(len(m.cipher))
	copy(data[22:(22+len(m.cipher))], m.cipher)
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

// DecryptKey decrypts encrypted key.
func (m *Metadata) DecryptKey(kek, tpmKek []byte) ([]byte, error) {
	if len(kek) != len(m.kek) {
		return nil, fmt.Errorf("key length mismatch: expected=%d, actual=%d", len(m.kek), len(kek))
	}

	key := make([]byte, len(kek))
	for i := range kek {
		if len(tpmKek) != 0 {
			key[i] = m.kek[i] ^ kek[i] ^ tpmKek[i]
		} else {
			key[i] = m.kek[i] ^ kek[i]
		}
	}
	return key, nil
}

// EncryptKey encrypts key.
func (m *Metadata) EncryptKey(key, tpmKek []byte) ([]byte, error) {
	return m.DecryptKey(key, tpmKek)
}
