package client

import (
	"context"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"os"
	"path"

	"github.com/cybozu-go/sabakan"
)

// AssetsIndex retrieves index of assets
func AssetsIndex(ctx context.Context) ([]string, *Status) {
	var index []string
	err := client.getJSON(ctx, "assets", nil, &index)
	if err != nil {
		return nil, err
	}
	return index, nil
}

// AssetsInfo retrieves meta data of an asset
func AssetsInfo(ctx context.Context, name string) (*sabakan.Asset, *Status) {
	var asset sabakan.Asset
	err := client.getJSON(ctx, path.Join("assets", name, "meta"), nil, &asset)
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

func detectContentTypeFromFile(file *os.File) (string, error) {
	// from filename extension; this may return ""
	contentType := mime.TypeByExtension(path.Ext(file.Name()))
	if len(contentType) != 0 {
		return contentType, nil
	}

	// from first 512 bytes; this returns "application/octec-stream" as fallback
	buf := make([]byte, 512)
	_, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		return "", err
	}
	return http.DetectContentType(buf), nil
}

// AssetsUpload stores a file as an asset
func AssetsUpload(ctx context.Context, name, filename string) (*sabakan.AssetStatus, *Status) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, ErrorStatus(err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, ErrorStatus(err)
	}
	size := fileInfo.Size()

	contentType, err := detectContentTypeFromFile(file)
	if err != nil {
		return nil, ErrorStatus(err)
	}

	req := client.NewRequest(ctx, "PUT", "assets/"+name, file)
	req.ContentLength = size
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Expect", "100-continue")

	resp, status := client.Do(req)
	if status != nil {
		return nil, status
	}
	defer resp.Body.Close()

	var result sabakan.AssetStatus
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, ErrorStatus(err)
	}

	return &result, nil
}

// AssetsDelete deletes an asset
func AssetsDelete(ctx context.Context, name string) *Status {
	return client.sendRequest(ctx, "DELETE", path.Join("assets", name), nil)
}
