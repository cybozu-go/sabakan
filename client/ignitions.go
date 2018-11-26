package client

import (
	"context"
	"fmt"
	"io"
	"path"
)

// IgnitionsGet gets ignition template metadata list of the specified role
func IgnitionsGet(ctx context.Context, role string) ([]map[string]string, error) {
	var metadata []map[string]string
	err := client.getJSON(ctx, "ignitions/"+role, nil, &metadata)
	if err != nil {
		return nil, err
	}
	return metadata, nil
}

// IgnitionsCat gets an ignition template for the role an id
func IgnitionsCat(ctx context.Context, role, id string, w io.Writer) error {
	req := client.NewRequest(ctx, "GET", path.Join("ignitions", role, id), nil)
	resp, status := client.Do(req)
	if status != nil {
		return status
	}
	defer resp.Body.Close()

	_, err := io.Copy(w, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

// IgnitionsSet posts an ignition template file
func IgnitionsSet(ctx context.Context, role string, fname string, meta map[string]string) error {
	tmpl, err := generateIgnitionYAML(fname)
	if err != nil {
		return err
	}
	req := client.NewRequest(ctx, "POST", "ignitions/"+role, tmpl)
	for k, v := range meta {
		req.Header.Set(fmt.Sprintf("X-Sabakan-Ignitions-%s", k), v)
	}
	resp, status := client.Do(req)
	if status != nil {
		return status
	}
	resp.Body.Close()

	return nil
}

// IgnitionsDelete deletes an ignition template specified by role and id
func IgnitionsDelete(ctx context.Context, role, id string) error {
	return client.sendRequest(ctx, "DELETE", path.Join("ignitions", role, id), nil)
}
