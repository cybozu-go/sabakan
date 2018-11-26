package client

import (
	"bytes"
	"context"
	"path"
)

// CryptsGet gets an encryption key from sabakan server.
func CryptsGet(ctx context.Context, serial, device string) ([]byte, error) {
	return client.getBytes(ctx, path.Join("crypts", serial, device))
}

// CryptsPut puts an encryption key to sabakan server.
func CryptsPut(ctx context.Context, serial, device string, key []byte) error {
	r := bytes.NewReader(key)
	return client.sendRequest(ctx, "PUT", path.Join("crypts", serial, device), r)
}

// CryptsDelete removes all encryption keys of the machine specified by serial.
func CryptsDelete(ctx context.Context, serial string) error {
	return client.sendRequest(ctx, "DELETE", path.Join("crypts", serial), nil)
}
