package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"strconv"

	"github.com/cybozu-go/log"
)

const (
	modprobeCmd   = "/sbin/modprobe"
	cryptsetupCmd = "/sbin/cryptsetup"
)

// InitModules load kernel modules for dm-crypt.
func InitModules() {
	err := exec.Command(modprobeCmd, "aesni-intel").Run()
	if err == nil {
		return
	}

	err = exec.Command(modprobeCmd, "aes-x86_64").Run()
	if err == nil {
		return
	}

	log.Warn("failed to load AES kernel modules", nil)
}

// Cryptsetup invokes cryptsetup to open crypt device.
// ek is the encrypted encryption key.
func Cryptsetup(d Disk, md *Metadata, ek []byte) error {
	key, err := md.DecryptKey(ek)
	if err != nil {
		return err
	}
	args := []string{
		"--hash=plain", "--key-file=-", "--cipher=" + md.Cipher(),
		"--key-size=" + strconv.Itoa(len(key)*8),
		"--offset=" + strconv.Itoa(metadataSize/512),
		// cryptsetup 2.0.0 and Linux 4.12 adds suport for larger sector sizes > 512.
		// https://www.saout.de/pipermail/dm-crypt/2017-December/005771.html
		// https://gitlab.com/cryptsetup/cryptsetup/wikis/DMCrypt
		//"--sector-size=" + strconv.Itoa(d.SectorSize()),
		"--allow-discards",
		"open", "--type=plain", d.Device(), d.CryptName(),
	}
	cmd := exec.Command(cryptsetupCmd, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = bytes.NewReader(key)
	return cmd.Run()
}
