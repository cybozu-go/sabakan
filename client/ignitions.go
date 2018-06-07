package client

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path"
)

// IgnitionsGet gets ignition template ID list of the specified role
func IgnitionsGet(ctx context.Context, role string) ([]string, *Status) {
	var ids []string
	err := client.getJSON(ctx, path.Join("/ignitions", role), nil, &ids)
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// IgnitionsCat gets an ignition template for the role an id
func IgnitionsCat(ctx context.Context, role, id string) (string, *Status) {
	req, err := http.NewRequest("GET", client.endpoint+path.Join("/api/v1/ignitions", role, id), nil)
	if err != nil {
		return "", ErrorStatus(err)
	}
	req = req.WithContext(ctx)
	res, err := client.http.Do(req)
	if err != nil {
		return "", ErrorStatus(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", ErrorStatus(err)
	}
	return string(body), nil
}

// IgnitionsSet posts an ignition template file
func IgnitionsSet(ctx context.Context, role string, fname string) (map[string]interface{}, *Status) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, ErrorStatus(err)
	}
	defer f.Close()

	req, err := http.NewRequest("POST", client.endpoint+path.Join("/api/v1/ignitions", role), f)
	if err != nil {
		return nil, ErrorStatus(err)
	}
	req = req.WithContext(ctx)
	res, err := client.http.Do(req)
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

// IgnitionsDelete deletes an ignition template specified by role and id
func IgnitionsDelete(ctx context.Context, role, id string) *Status {
	req, err := http.NewRequest("DELETE", client.endpoint+path.Join("/api/v1/ignitions", role, id), nil)
	if err != nil {
		return ErrorStatus(err)
	}
	req = req.WithContext(ctx)
	res, err := client.http.Do(req)
	if err != nil {
		return ErrorStatus(err)
	}
	res.Body.Close()
	return nil
}
