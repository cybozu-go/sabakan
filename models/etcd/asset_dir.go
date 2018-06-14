package etcd

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/cybozu-go/log"
)

// AssetDir is a struct to manage the assets directory.
type AssetDir struct {
	Dir string
}

// Path returns a path to an asset of id.
func (d AssetDir) Path(id int) string {
	return filepath.Join(d.Dir, strconv.Itoa(id))
}

// Exists returns true if a local copy of an asset of id exists.
func (d AssetDir) Exists(id int) bool {
	_, err := os.Stat(d.Path(id))
	return err == nil
}

// Remove removes an asset.
func (d AssetDir) Remove(id int) error {
	return os.Remove(d.Path(id))
}

// Save stores an asset.
// When successful, this returns SHA256 checksum of the contents.
func (d AssetDir) Save(id int, r io.Reader, csum []byte) ([]byte, error) {
	err := os.MkdirAll(d.Dir, 0755)
	if err != nil {
		return nil, err
	}

	f, err := ioutil.TempFile(d.Dir, ".tmp")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	h := sha256.New()
	w := io.MultiWriter(f, h)
	_, err = io.Copy(w, r)
	if err != nil {
		return nil, err
	}

	hsum := h.Sum(nil)
	if csum != nil && !bytes.Equal(csum, hsum) {
		os.Remove(f.Name())
		return nil, fmt.Errorf("checksum mismatch for id %d", id)
	}

	err = f.Sync()
	if err != nil {
		return nil, err
	}

	dest := d.Path(id)
	err = os.Rename(f.Name(), dest)
	if err != nil {
		return nil, err
	}

	return hsum, nil
}

// GC removes garbage, that is, files whose names are not in ids.
func (d AssetDir) GC(ids map[int]bool) error {
	fil, err := ioutil.ReadDir(d.Dir)
	if err != nil {
		return err
	}

	for _, fi := range fil {
		if !fi.Mode().IsRegular() {
			continue
		}

		id, err := strconv.Atoi(fi.Name())
		if err != nil || !ids[id] {
			log.Info("removing garbage asset", map[string]interface{}{
				"name": fi.Name(),
			})
			os.Remove(filepath.Join(d.Dir, fi.Name()))
		}
	}

	return nil
}
