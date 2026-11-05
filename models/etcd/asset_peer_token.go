package etcd

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"errors"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// assetPeerToken returns the token shared by sabakan servers, creating it if it does not exist yet.
func (d *driver) assetPeerToken(ctx context.Context) (string, error) {
	newToken := rand.Text()
	resp, err := d.client.Txn(ctx).
		If(clientv3.Compare(clientv3.CreateRevision(KeyAssetPeerToken), "=", 0)).
		Then(clientv3.OpPut(KeyAssetPeerToken, newToken)).
		Else(clientv3.OpGet(KeyAssetPeerToken)).
		Commit()
	if err != nil {
		return "", err
	}
	if resp.Succeeded {
		return newToken, nil
	}

	kvs := resp.Responses[0].GetResponseRange().Kvs
	if len(kvs) == 0 {
		return "", errors.New("asset peer token has been deleted")
	}
	return string(kvs[0].Value), nil
}

// verifyAssetPeerToken does not create the token, so no token is valid until a server pulls a restricted asset.
func (d *driver) verifyAssetPeerToken(ctx context.Context, token string) (bool, error) {
	if token == "" {
		return false, nil
	}
	resp, err := d.client.Get(ctx, KeyAssetPeerToken)
	if err != nil {
		return false, err
	}
	if len(resp.Kvs) == 0 {
		return false, nil
	}
	return subtle.ConstantTimeCompare(resp.Kvs[0].Value, []byte(token)) == 1, nil
}
