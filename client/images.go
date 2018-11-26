package client

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"os"
	"path"

	"github.com/cybozu-go/sabakan"
	"github.com/pkg/errors"
)

// ImagesIndex get index of images.
func ImagesIndex(ctx context.Context, os string) (sabakan.ImageIndex, error) {
	var index sabakan.ImageIndex
	err := client.getJSON(ctx, "images/"+os, nil, &index)
	if err != nil {
		return nil, err
	}
	return index, nil
}

// ImagesUpload upload image file.
func ImagesUpload(ctx context.Context, os, id, kernel, initrd string) error {
	reader, err := createImageArchive(kernel, initrd)
	if err != nil {
		return err
	}

	return client.sendRequest(ctx, "PUT", path.Join("images", os, id), reader)
}

func addFileToTar(tw *tar.Writer, name, p string) error {
	fi, err := os.Stat(p)
	if err != nil {
		return err
	}
	f, err := os.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()
	hdr := &tar.Header{
		Name: name,
		Mode: 0644,
		Size: fi.Size(),
	}
	err = tw.WriteHeader(hdr)
	if err != nil {
		return err
	}
	n, err := io.Copy(tw, f)
	if err != nil {
		return err
	}
	if n != fi.Size() {
		return errors.New("written size mismatch")
	}
	return nil
}

func createImageArchive(kernelPath, initrdPath string) (io.Reader, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	err := addFileToTar(tw, sabakan.ImageKernelFilename, kernelPath)
	if err != nil {
		return nil, err
	}
	err = addFileToTar(tw, sabakan.ImageInitrdFilename, initrdPath)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// ImagesDelete deletes image file.
func ImagesDelete(ctx context.Context, os, id string) error {
	return client.sendRequest(ctx, "DELETE", path.Join("images", os, id), nil)
}
