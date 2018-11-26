package client

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"path"

	"github.com/cybozu-go/sabakan"
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
func ImagesUpload(ctx context.Context, os, id string, kernel io.Reader, kernelSize int64, initrd io.Reader, initrdSize int64) error {
	reader, err := createImageArchive(kernel, kernelSize, initrd, initrdSize)
	if err != nil {
		return err
	}

	return client.sendRequest(ctx, "PUT", path.Join("images", os, id), reader)
}

func addFileToTar(tw *tar.Writer, name string, src io.Reader, size int64) error {
	hdr := &tar.Header{
		Name: name,
		Mode: 0644,
		Size: size,
	}
	err := tw.WriteHeader(hdr)
	if err != nil {
		return err
	}
	_, err = io.CopyN(tw, src, size)
	if err != nil {
		return err
	}
	return nil
}

func createImageArchive(kernel io.Reader, kernelSize int64, initrd io.Reader, initrdSize int64) (io.Reader, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	err := addFileToTar(tw, sabakan.ImageKernelFilename, kernel, kernelSize)
	if err != nil {
		return nil, err
	}
	err = addFileToTar(tw, sabakan.ImageInitrdFilename, initrd, initrdSize)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// ImagesDelete deletes image file.
func ImagesDelete(ctx context.Context, os, id string) error {
	return client.sendRequest(ctx, "DELETE", path.Join("images", os, id), nil)
}
