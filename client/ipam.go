package client

import (
	"context"

	"github.com/cybozu-go/sabakan"
)

// IPAMConfigGet retrieves IPAM configurations
func (c *Client) IPAMConfigGet(ctx context.Context) (*sabakan.IPAMConfig, *Status) {
	var conf sabakan.IPAMConfig
	err := c.getJSON(ctx, "/config/ipam", nil, &conf)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}

// IPAMConfigSet sets IPAM configurations
func (c *Client) IPAMConfigSet(ctx context.Context, conf *sabakan.IPAMConfig) *Status {
	return c.sendRequestWithJSON(ctx, "PUT", "/config/ipam", conf)
}
