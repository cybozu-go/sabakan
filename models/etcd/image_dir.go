package etcd

import (
	"archive/tar"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cybozu-go/sabakan"
)

/*
Directory structure:

/var/lib/sabakan/
    - OS (coreos) /
        - ID1/
            - kernel
            - initrd.gz
        - ID2/
            - kernel
            - initrd.gz
        - ...

While extracting the image archive, a temporary directory is created
under `var/lib/sabakan/OS/`.  The temporary directory will be renamed
to `IDxx` once extraction successfully completed.
*/

// ImageDir is a struct to manage an image directory.
type ImageDir struct {
	// Dir is an absolute path to point image directory of an OS.
	Dir string
}

func writeToFile(p string, r io.Reader) error {
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()

	err = f.Chmod(0644)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, r)
	if err != nil {
		return err
	}

	return f.Sync()
}

func (d ImageDir) Extract(r io.Reader, id string, members []string) error {
	defer func() {
		io.Copy(ioutil.Discard, r)
	}()

	tmpdir, err := ioutil.TempDir(d.Dir, "_tmp")
	if err != nil {
		return err
	}
	defer func() {
		if tmpdir == "" {
			return
		}
		os.RemoveAll(tmpdir)
	}()

	memberMap := make(map[string]bool)
	for _, m := range members {
		memberMap[m] = true
	}

	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if !memberMap[hdr.Name] {
			return sabakan.ErrBadRequest
		}
		delete(memberMap, hdr.Name)

		err = writeToFile(filepath.Join(tmpdir, hdr.Name), tr)
		if err != nil {
			return err
		}
	}

	if len(memberMap) > 0 {
		return sabakan.ErrBadRequest
	}

	err = os.Rename(tmpdir, filepath.Join(d.Dir, id))
	if err != nil {
		return err
	}
	tmpdir = ""
	return nil
}

func (d ImageDir) GC(ids []string) error {
	for _, id := range ids {
		p := filepath.Join(d.Dir, id)
		err := os.RemoveAll(p)
		if err != nil {
			return err
		}
	}
	return nil
}
