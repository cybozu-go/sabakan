package client

import (
	"context"
	"path"
)

// CryptsGet get crypt key bytes from sabakan server
func CryptsGet(ctx context.Context, serial, device string) ([]byte, *Status) {
	return client.getBytes(ctx, path.Join("crypts", serial, device))
}

// CryptsPut create the crypt key to sabakan server
func CryptsPut(ctx context.Context, serial, device string, key []byte) *Status {
	return client.sendRequestWithBytes(ctx, "PUT", path.Join("crypts", serial, device), key)
}

// CryptsDelete remove the all crypt keys specified serial
func CryptsDelete(ctx context.Context, serial string) *Status {
	return client.sendRequest(ctx, "DELETE", path.Join("crypts", serial))
}
