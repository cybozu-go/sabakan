package etcd

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/sabakan/v2"
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
under `/var/lib/sabakan/OS/`.  The temporary directory will be renamed
to `IDxx` once extraction successfully completed.
If the rename fails due to the existing `IDxx` directory, the contents of
two directories are compared.  If the two are not equivalent, a conflict
error will be returned.
*/

// ImageDir is a struct to manage an image directory.
type ImageDir struct {
	// Dir is an absolute path to point image directory of an OS.
	Dir string
}

// Exists returns true if image files referenced by "id"
// is stored in the directory.
func (d ImageDir) Exists(id string) bool {
	p := filepath.Join(d.Dir, id)
	_, err := os.Stat(p)
	return err == nil
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

func equalFileContent(filename0, filename1 string) (bool, error) {
	f0, err := os.Open(filename0)
	if err != nil {
		return false, err
	}
	defer f0.Close()

	f1, err := os.Open(filename1)
	if err != nil {
		return false, err
	}
	defer f1.Close()

	fi0, err := f0.Stat()
	if err != nil {
		return false, err
	}
	fi1, err := f1.Stat()
	if err != nil {
		return false, err
	}
	if fi0.Size() != fi1.Size() {
		return false, nil
	}

	buf0 := make([]byte, 16*1024)
	buf1 := make([]byte, 16*1024)
	for {
		n0, err := f0.Read(buf0)
		if err != nil && err != io.EOF {
			return false, err
		}
		buf0 = buf0[0:n0]

		n1, err := f1.Read(buf1)
		if err != nil && err != io.EOF {
			return false, err
		}
		buf1 = buf1[0:n1]

		if n0 == 0 {
			break
		}

		if !bytes.Equal(buf0, buf1) {
			return false, nil
		}
	}

	return true, nil
}

// Extract reads tar archive from "r" to extract files shown in "members".
//
// Extracted files are finally stored under "id" directory.
// If the tar archive contains a file not in "members", or if the tar
// archive lacks a file in "members", this function returns sabakan.ErrBadRequest.
func (d ImageDir) Extract(r io.Reader, id string, members []string) error {
	defer func() {
		io.Copy(io.Discard, r)
	}()

	err := os.MkdirAll(d.Dir, 0755)
	if err != nil {
		return err
	}

	dstdir := filepath.Join(d.Dir, id)

	tmpdir, err := os.MkdirTemp(d.Dir, "_tmp")
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

	err = os.Rename(tmpdir, dstdir)
	if err == nil {
		tmpdir = ""
		return nil
	}

	if !os.IsExist(err) {
		return err
	}

	for _, m := range members {
		eq, err := equalFileContent(filepath.Join(tmpdir, m), filepath.Join(dstdir, m))
		if err != nil || !eq {
			return fmt.Errorf("different content")
		}
	}

	return nil
}

func copyFile(w io.Writer, p string) error {
	f, err := os.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(w, f)
	return err
}

// Download writes files in a directory specified by "id" as a tar archive.
func (d ImageDir) Download(w io.Writer, id string) error {
	files, err := os.ReadDir(filepath.Join(d.Dir, id))
	if err != nil {
		return err
	}

	tw := tar.NewWriter(w)

	for _, fi := range files {
		if !fi.Type().IsRegular() {
			log.Warn("non-regular file in image dir", map[string]interface{}{
				"dir":  filepath.Join(d.Dir, id),
				"name": fi.Name(),
			})
			continue
		}

		info, err := fi.Info()
		if err != nil {
			return err
		}

		hdr := &tar.Header{
			Name: fi.Name(),
			Size: info.Size(),
			Mode: 0644,
		}
		err = tw.WriteHeader(hdr)
		if err != nil {
			return err
		}
		err = copyFile(tw, filepath.Join(d.Dir, id, fi.Name()))
		if err != nil {
			return err
		}
	}

	return tw.Close()
}

// ServeFile opens filename in "id" directory then calls "f" with the opened file.
func (d ImageDir) ServeFile(id, filename string, f func(content io.ReadSeeker)) error {
	p := filepath.Join(d.Dir, id, filename)
	g, err := os.Open(p)
	if err != nil {
		return err
	}
	defer g.Close()

	f(g)
	return nil
}

// GC removes images listed in "ids".
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

// Size returns byte size of the image
func (d ImageDir) Size(id string) (int64, error) {
	var imageSize int64
	files, err := os.ReadDir(filepath.Join(d.Dir, id))
	if err != nil {
		return 0, err
	}
	for _, file := range files {
		fi, err := file.Info()
		if err != nil {
			return 0, err
		}
		imageSize += fi.Size()
	}
	return imageSize, nil
}
