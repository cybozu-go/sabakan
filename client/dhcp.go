package client

import (
	"context"

	"github.com/cybozu-go/sabakan"
)

// DHCPConfigGet retrieves DHCP configurations
func DHCPConfigGet(ctx context.Context) (*sabakan.DHCPConfig, error) {
	conf := new(sabakan.DHCPConfig)
	err := client.getJSON(ctx, "config/dhcp", nil, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

// DHCPConfigSet sets DHCP configurations
func DHCPConfigSet(ctx context.Context, conf *sabakan.DHCPConfig) error {
	return client.sendRequestWithJSON(ctx, "PUT", "config/dhcp", conf)
}
