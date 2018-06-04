package client

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path"
)

// IgnitionsGet get ignition template ID list of the specified role
func (c *Client) IgnitionsGet(ctx context.Context, role string) ([]string, *Status) {
	var ids []string
	err := c.getJSON(ctx, path.Join("/ignitions", role), nil, &ids)
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// IgnitionsSet post a ignition template file
func (c *Client) IgnitionsSet(ctx context.Context, role string, fname string) (map[string]interface{}, *Status) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, ErrorStatus(err)
	}
	defer f.Close()

	req, err := http.NewRequest("POST", c.endpoint+path.Join("/api/v1/ignitions", role), f)
	if err != nil {
		return nil, ErrorStatus(err)
	}
	req = req.WithContext(ctx)
	res, err := c.http.Do(req)
	if err != nil {
		return nil, ErrorStatus(err)
	}
	defer res.Body.Close()

	var data map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return nil, ErrorStatus(err)
	}

	return data, nil
}
