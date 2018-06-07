package client

import (
	"context"

	"github.com/cybozu-go/sabakan"
)

// IPAMConfigGet retrieves IPAM configurations
func IPAMConfigGet(ctx context.Context) (*sabakan.IPAMConfig, *Status) {
	var conf sabakan.IPAMConfig
	err := client.getJSON(ctx, "/config/ipam", nil, &conf)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}

// IPAMConfigSet sets IPAM configurations
func IPAMConfigSet(ctx context.Context, conf *sabakan.IPAMConfig) *Status {
	return client.sendRequestWithJSON(ctx, "PUT", "/config/ipam", conf)
}
