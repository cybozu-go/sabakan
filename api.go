package sabakan

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cybozu-go/cmd"
)

// Client is a sabakan client
type Client struct {
	endpoint string
	http     *cmd.HTTPClient
}

// NewClient creates a new sabakan client
func NewClient(endpoint string, http *cmd.HTTPClient) *Client {
	return &Client{
		endpoint: endpoint,
		http:     http,
	}
}

// RemoteConfigGet gets a remote config
func (c *Client) RemoteConfigGet(ctx context.Context) (*Config, error) {
	var conf Config
	err := c.jsonGet(ctx, "/config", nil, &conf)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}

// RemoteConfigSet sets a remote config
func (c *Client) RemoteConfigSet(ctx context.Context, conf *Config) error {
	return c.jsonPost(ctx, "/config", conf)
}

// MachinesGet get machine information from sabakan server
func (c *Client) MachinesGet(ctx context.Context, params map[string]string) ([]Machine, error) {
	var machines []Machine
	err := c.jsonGet(ctx, "/machines", params, &machines)
	if err != nil {
		return nil, err
	}
	return machines, nil
}

// MachinesCreate create machines information to sabakan server
func (c *Client) MachinesCreate(ctx context.Context, machines []Machine) error {
	return c.jsonPost(ctx, "/machines", machines)
}

// MachinesUpdate update machines information on sabakan server
func (c *Client) MachinesUpdate(ctx context.Context, machines []Machine) error {
	return nil
}

func (c *Client) jsonGet(ctx context.Context, path string, params map[string]string, data interface{}) error {
	req, err := http.NewRequest("GET", c.endpoint+"/api/v1"+path, nil)
	if err != nil {
		return err
	}
	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
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

func (c *Client) jsonPost(ctx context.Context, path string, data interface{}) error {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.endpoint+"/api/v1"+path, b)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		// TODO: return a message from a server"
		return fmt.Errorf("server returns failure code: " + res.Status)
	}
	return nil
}
