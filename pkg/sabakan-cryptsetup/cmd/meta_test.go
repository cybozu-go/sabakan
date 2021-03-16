package cmd

import (
	"encoding/hex"
	"os"
	"testing"
)

func TestMetadata(t *testing.T) {
	t.Parallel()

	f, err := os.CreateTemp("", "gotest.")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		name := f.Name()
		f.Close()
		os.Remove(name)
	}()

	_, err = ReadMetadata(f)
	if err == nil {
		t.Error("metadata should not be found")
	}

	md, err := NewMetadata("aes-xts-plain64", 64, Tpm20)
	if err != nil {
		t.Fatal(err)
	}
	if md.Cipher() != "aes-xts-plain64" {
		t.Error(`md.Cipher() != "aes-xts-plain64"`, md.Cipher())
	}
	if len(md.ID()) != idLength {
		t.Error(`len(m.ID()) != idLength`, len(md.ID()))
	} else {
		t.Log(md.HexID())
	}
	if len(md.Kek()) != 64 {
		t.Error(`len(m.Kek()) != 64`, len(md.Kek()))
	} else {
		t.Log(hex.EncodeToString([]byte(md.Kek())))
	}
	if md.TpmVersion() != Tpm20 {
		t.Error(`md.TpmVersion() != Tpm20`, md.TpmVersion().String())
	}

	err = md.Write(f)
	if err != nil {
		t.Fatal(err)
	}

	md2, err := ReadMetadata(f)
	if err != nil {
		t.Fatal(err)
	}

	if md.cipher != md2.cipher {
		t.Error(`md.cipher != md2.cipher`, md2.cipher)
	}
	if md.id != md2.id {
		t.Error(`md.id != md2.id`, md2.id)
	}
	if md.kek != md2.kek {
		t.Error(`md.kek != md2.kek`, md2.kek)
	}
}
