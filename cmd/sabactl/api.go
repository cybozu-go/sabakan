package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cybozu-go/cmd"
	"github.com/cybozu-go/sabakan"
)

type client struct {
	endpoint string
	http     *cmd.HTTPClient
}

func (c *client) remoteConfigGet(ctx context.Context) (*sabakan.Config, error) {
	var conf sabakan.Config
	err := c.jsonGet(ctx, "/config", &conf)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}

func (c *client) remoteConfigSet(ctx context.Context, conf *sabakan.Config) error {
	// TODO psot configure
	return nil
}

func (c *client) jsonGet(ctx context.Context, path string, data interface{}) error {
	req, err := http.NewRequest("GET", c.endpoint+"/api/v1"+path, nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		// TODO: return a message from a server"
		return fmt.Errorf("server returns failure code: " + res.Status)
	}
	return json.NewDecoder(res.Body).Decode(data)
}
