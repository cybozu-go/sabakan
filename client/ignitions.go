package client

import (
	"context"
	"io"
	"path"
)

// IgnitionsGet gets ignition template ID list of the specified role
func IgnitionsGet(ctx context.Context, role string) ([]string, *Status) {
	var ids []string
	err := client.getJSON(ctx, "ignitions/"+role, nil, &ids)
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// IgnitionsCat gets an ignition template for the role an id
func IgnitionsCat(ctx context.Context, role, id string, w io.Writer) *Status {
	req := client.NewRequest(ctx, "GET", path.Join("ignitions", role, id), nil)
	resp, status := client.Do(req)
	if status != nil {
		return status
	}
	defer resp.Body.Close()

	_, err := io.Copy(w, resp.Body)
	if err != nil {
		return ErrorStatus(err)
	}
	return nil
}

// IgnitionsSet posts an ignition template file
func IgnitionsSet(ctx context.Context, role string, fname string) *Status {
	tmpl, err := generateIgnitionYAML(fname)
	if err != nil {
		return ErrorStatus(err)
	}
	return client.sendRequest(ctx, "POST", "ignitions/"+role, tmpl)
}

// IgnitionsDelete deletes an ignition template specified by role and id
func IgnitionsDelete(ctx context.Context, role, id string) *Status {
	return client.sendRequest(ctx, "DELETE", path.Join("ignitions", role, id), nil)
}
