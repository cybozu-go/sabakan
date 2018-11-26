package client

import (
	"context"
	"path"
	"strings"

	"github.com/cybozu-go/sabakan"
)

// KernelParamsGet retrieves kernel parameters
func KernelParamsGet(ctx context.Context, os string) (sabakan.KernelParams, error) {
	body, status := client.getBytes(ctx, path.Join("kernel_params", os))
	if status != nil {
		return "", status
	}

	return sabakan.KernelParams(body), status
}

// KernelParamsSet sets kernel parameters
func KernelParamsSet(ctx context.Context, os string, params sabakan.KernelParams) error {
	r := strings.NewReader(string(params))
	return client.sendRequest(ctx, "PUT", path.Join("kernel_params", os), r)
}
