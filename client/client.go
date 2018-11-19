package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/user"
	"path"

	"github.com/cybozu-go/well"
)

// Client is a sabakan client
type Client struct {
	url  *url.URL
	http *well.HTTPClient
}

var (
	client   *Client
	username string
)

// Setup initializes client package.
func Setup(endpoint string, http *well.HTTPClient) error {
	u, err := url.Parse(endpoint)
	if err != nil {
		return err
	}

	user, err := user.Current()
	if err != nil {
		return err
	}
	username = user.Username

	client = &Client{
		url:  u,
		http: http,
	}
	return nil
}

// NewRequest creates a new http.Request whose context is set to ctx.
// path will be prefixed by "/api/v1".
func (c *Client) NewRequest(ctx context.Context, method, p string, body io.Reader) *http.Request {
	u := *c.url
	u.Path = path.Join(u.Path, "/api/v1", p)
	r, _ := http.NewRequest(method, u.String(), body)
	r.Header.Set("X-Sabakan-User", username)
	return r.WithContext(ctx)
}

// Do calls http.Client.Do and processes errors.
// This returns non-nil *http.Response only when the server returns 2xx status code.
func (c *Client) Do(req *http.Request) (*http.Response, *Status) {
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, ErrorStatus(err)
	}

	errorStatus := ErrorHTTPStatus(resp)
	if errorStatus != nil {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
		return nil, errorStatus
	}

	return resp, nil
}

func (c *Client) getJSON(ctx context.Context, p string, params map[string]string, data interface{}) *Status {
	req := c.NewRequest(ctx, "GET", p, nil)
	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	resp, status := c.Do(req)
	if status != nil {
		return status
	}
	defer resp.Body.Close()

	err := json.NewDecoder(resp.Body).Decode(data)
	if err != nil {
		return ErrorStatus(err)
	}

	return nil
}

func (c *Client) getBytes(ctx context.Context, p string) ([]byte, *Status) {
	req := c.NewRequest(ctx, "GET", p, nil)
	resp, status := c.Do(req)
	if status != nil {
		return nil, status
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrorStatus(err)
	}
	return body, nil
}

func (c *Client) sendRequestWithJSON(ctx context.Context, method, p string, data interface{}) *Status {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(data)
	if err != nil {
		return ErrorStatus(err)
	}

	req := c.NewRequest(ctx, method, p, b)
	resp, status := c.Do(req)
	if status != nil {
		return status
	}
	resp.Body.Close()

	return nil
}

func (c *Client) sendRequest(ctx context.Context, method, p string, r io.Reader) *Status {
	req := c.NewRequest(ctx, method, p, r)
	resp, status := c.Do(req)
	if status != nil {
		return status
	}
	resp.Body.Close()

	return nil
}
