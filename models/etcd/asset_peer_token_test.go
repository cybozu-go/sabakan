package etcd

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/cybozu-go/sabakan/v3"
)

func testAssetPeerTokenCreate(t *testing.T) {
	t.Parallel()

	d, _ := testNewDriver(t)
	ctx := context.Background()

	valid, err := d.verifyAssetPeerToken(ctx, "foo")
	if err != nil {
		t.Fatal(err)
	}
	if valid {
		t.Error("token should be invalid before it is created")
	}

	token, err := d.assetPeerToken(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Fatal("token is empty")
	}
	token2, err := d.assetPeerToken(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if token2 != token {
		t.Errorf("token changed: %q -> %q", token, token2)
	}

	tests := []struct {
		name  string
		token string
		want  bool
	}{
		{"same token", token, true},
		{"wrong token", token + "x", false},
		{"empty token", "", false},
	}
	for _, tt := range tests {
		got, err := d.verifyAssetPeerToken(ctx, tt.token)
		if err != nil {
			t.Fatal(err)
		}
		if got != tt.want {
			t.Errorf("%s: verifyAssetPeerToken = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func testAssetPeerTokenDownload(t *testing.T) {
	t.Parallel()

	d, _ := testNewDriver(t)
	d.dataDir = t.TempDir()
	ctx := context.Background()

	token, err := d.assetPeerToken(ctx)
	if err != nil {
		t.Fatal(err)
	}

	content := []byte("bar")
	sum := sha256.Sum256(content)
	tests := []struct {
		name      string
		id        int
		options   map[string]string
		wantToken string
	}{
		{"restricted asset", 1, map[string]string{sabakan.AssetOptionAllowedRoles: "cs"}, token},
		{"other asset", 2, nil, ""},
	}
	for _, tt := range tests {
		var gotToken string
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotToken = r.Header.Get(sabakan.AssetPeerTokenHeader)
			w.Header().Set("X-Sabakan-Asset-ID", strconv.Itoa(tt.id))
			w.Header().Set("X-Sabakan-Asset-SHA256", hex.EncodeToString(sum[:]))
			_, _ = w.Write(content)
		}))

		asset := &sabakan.Asset{
			Name:    "foo",
			ID:      tt.id,
			URLs:    []string{ts.URL + "/api/v1/assets/foo"},
			Options: tt.options,
		}
		err := d.downloadAsset(ctx, asset)
		ts.Close()
		if err != nil {
			t.Fatalf("%s: %v", tt.name, err)
		}
		if gotToken != tt.wantToken {
			t.Errorf("%s: token = %q, want %q", tt.name, gotToken, tt.wantToken)
		}
	}
}

func TestAssetPeerToken(t *testing.T) {
	t.Run("Create", testAssetPeerTokenCreate)
	t.Run("Download", testAssetPeerTokenDownload)
}
