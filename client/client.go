package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path"

	"github.com/cybozu-go/cmd"
)

// Client is a sabakan client
type Client struct {
	endpoint string
	http     *cmd.HTTPClient
}

var (
	client *Client
)

// Setup initializes client package.
func Setup(endpoint string, http *cmd.HTTPClient) {
	client = &Client{
		endpoint: endpoint,
		http:     http,
	}
}

func (c *Client) getJSON(ctx context.Context, path string, params map[string]string, data interface{}) *Status {
	req, err := http.NewRequest("GET", c.endpoint+"/api/v1"+path, nil)
	if err != nil {
		return ErrorStatus(err)
	}
	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	req = req.WithContext(ctx)
	res, err := c.http.Do(req)
	if err != nil {
		return ErrorStatus(err)
	}
	defer res.Body.Close()

	errorStatus := ErrorHTTPStatus(res)
	if errorStatus != nil {
		return errorStatus
	}

	err = json.NewDecoder(res.Body).Decode(data)
	if err != nil {
		return ErrorStatus(err)
	}

	return nil
}

func (c *Client) getBytes(ctx context.Context, path string) ([]byte, *Status) {
	req, err := http.NewRequest("GET", c.endpoint+"/api/v1"+path, nil)
	if err != nil {
		return nil, ErrorStatus(err)
	}
	req = req.WithContext(ctx)
	res, err := c.http.Do(req)
	if err != nil {
		return nil, ErrorStatus(err)
	}
	defer res.Body.Close()
	errorStatus := ErrorHTTPStatus(res)
	if errorStatus != nil {
		return nil, errorStatus
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, ErrorStatus(err)
	}
	return body, nil
}

func (c *Client) sendRequestWithJSON(ctx context.Context, method string, path string, data interface{}) *Status {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(data)
	if err != nil {
		return ErrorStatus(err)
	}

	req, err := http.NewRequest(method, c.endpoint+"/api/v1"+path, b)
	if err != nil {
		return ErrorStatus(err)
	}
	req = req.WithContext(ctx)
	res, err := c.http.Do(req)
	if err != nil {
		return ErrorStatus(err)
	}
	defer res.Body.Close()

	return ErrorHTTPStatus(res)
}

func (c *Client) sendRequestWithBytes(ctx context.Context, method, resource string, body []byte) *Status {
	req, err := http.NewRequest(method, c.endpoint+path.Join("/api/v1", resource), bytes.NewBuffer(body))
	if err != nil {
		return ErrorStatus(err)
	}
	req = req.WithContext(ctx)
	res, err := c.http.Do(req)
	if err != nil {
		return ErrorStatus(err)
	}
	defer res.Body.Close()

	return ErrorHTTPStatus(res)
}

func (c *Client) sendRequest(ctx context.Context, method, path string) *Status {
	return c.sendRequestWithBytes(ctx, method, path, nil)
}
