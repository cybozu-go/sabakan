package web

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/cybozu-go/sabakan/v3"
	"github.com/cybozu-go/sabakan/v3/models/mock"
)

func testHandleAssetsGetIndex(t *testing.T) {
	t.Parallel()

	m := mock.NewModel()
	handler := newTestServer(m)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/assets", nil)
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("resp.StatusCode != http.StatusOK:", resp.StatusCode)
	}

	var data []string
	err := json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 0 {
		t.Error("len(data) != 0:", len(data))
	}

	_, err = m.Asset.Put(context.Background(), "foo", "text/plain", nil, nil, strings.NewReader("bar"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = m.Asset.Put(context.Background(), "abc", "text/plain", nil, nil, strings.NewReader("bar"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = m.Asset.Put(context.Background(), "xyz", "text/plain", nil, nil, strings.NewReader("bar"))
	if err != nil {
		t.Fatal(err)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/api/v1/assets", nil)
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("resp.StatusCode != http.StatusOK:", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(data, []string{"abc", "foo", "xyz"}) {
		t.Error("data is not valid:", data)
	}
}

func testHandleAssetsGetInfo(t *testing.T) {
	t.Parallel()

	m := mock.NewModel()
	handler := newTestServer(m)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/assets/foo/meta", nil)
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatal("resp.StatusCode != http.StatusNotFound:", resp.StatusCode)
	}

	options := map[string]string{"version": "1.0.0"}

	_, err := m.Asset.Put(context.Background(), "foo", "text/plain", nil, options, strings.NewReader("bar"))
	if err != nil {
		t.Fatal(err)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/api/v1/assets/foo/meta", nil)
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("resp.StatusCode != http.StatusOK:", resp.StatusCode)
	}

	var data sabakan.Asset
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		t.Fatal(err)
	}
	if data.Name != "foo" {
		t.Error("data.Name != foo:", data.Name)
	}
	if data.ContentType != "text/plain" {
		t.Error("data.ContentType != text/plain:", data.ContentType)
	}
	if data.Sha256 != "fcde2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9" {
		t.Error("data.Sha256 is not valid:", data.Sha256)
	}
	if data.Options["version"] != "1.0.0" {
		t.Error("invalid version:", data.Options)
	}
}

func testHandleAssetsGet(t *testing.T) {
	t.Parallel()

	m := mock.NewModel()
	handler := newTestServer(m)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/assets/foo", nil)
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatal("resp.StatusCode != http.StatusNotFound:", resp.StatusCode)
	}

	_, err := m.Asset.Put(context.Background(), "foo", "text/plain", nil, nil, strings.NewReader("bar"))
	if err != nil {
		t.Fatal(err)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/api/v1/assets/foo", nil)
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("resp.StatusCode != http.StatusOK:", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "bar" {
		t.Error("string(data) != bar", string(data))
	}
	if resp.Header.Get("content-type") != "text/plain" {
		t.Error("content-type != text/plain", resp.Header.Get("content-type"))
	}
	if len(resp.Header.Get("X-Sabakan-Asset-ID")) == 0 {
		t.Error("X-Sabakan-Asset-ID is not set")
	}
	if resp.Header.Get("X-Sabakan-Asset-SHA256") != "fcde2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9" {
		t.Error("X-Sabakan-Asset-SHA256 is not valid:", resp.Header.Get("X-Sabakan-Asset-SHA256"))
	}
}

func testHandleAssetsPut(t *testing.T) {
	t.Parallel()
	m := mock.NewModel()
	handler := newTestServer(m)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/api/v1/assets/foo", strings.NewReader("bar"))
	r.Header.Set("content-length", "3")
	r.Header.Set("content-type", "text/plain")
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Fatal("resp.StatusCode != http.StatusCreated:", resp.StatusCode)
	}
	var status sabakan.AssetStatus
	err := json.NewDecoder(resp.Body).Decode(&status)
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != http.StatusCreated {
		t.Error("status.Status != http.StatusCreated:", status.Status)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", "/api/v1/assets/foo", strings.NewReader("bar"))
	r.Header.Set("content-length", "3")
	r.Header.Set("content-type", "text/plain")
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("resp.StatusCode != http.StatusOK:", resp.StatusCode)
	}
	err = json.NewDecoder(resp.Body).Decode(&status)
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != http.StatusOK {
		t.Error("status.Status != httpStatusOK:", status.Status)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", "/api/v1/assets/foo", strings.NewReader("bar"))
	r.Header.Set("content-length", "3")
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("resp.StatusCode != http.StatusBadRequest:", resp.StatusCode)
	}

	pr, pw := io.Pipe()
	go func() {
		_, _ = pw.Write([]byte("bar"))
		pw.Close()
	}()
	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", "/api/v1/assets/foo", pr)
	r.Header.Set("content-type", "text/plain")
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusLengthRequired {
		t.Fatal("resp.StatusCode != http.StatusLengthRequired:", resp.StatusCode)
	}

	pr2, pw2 := io.Pipe()
	go func() {
		_, _ = pw2.Write([]byte("bar"))
		pw2.Close()
	}()
	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", "/api/v1/assets/foo", pr2)
	r.Header.Set("content-length", strconv.FormatInt(5<<30, 10))
	r.ContentLength = 5 << 30
	r.Header.Set("content-type", "text/plain")
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatal("resp.StatusCode != http.StatusRequestEntityTooLarge:", resp.StatusCode)
	}

	// checksum validation
	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", "/api/v1/assets/foo", strings.NewReader("bar"))
	r.Header.Set("content-length", "3")
	r.Header.Set("content-type", "text/plain")
	r.Header.Set("X-Sabakan-Asset-SHA256", "FCDE2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9")
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Error("resp.StatusCode != http.StatusOK:", resp.StatusCode)
	}

	// checksum mismatch
	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", "/api/v1/assets/foo", strings.NewReader("bar"))
	r.Header.Set("content-length", "3")
	r.Header.Set("content-type", "text/plain")
	r.Header.Set("X-Sabakan-Asset-SHA256", "0cde2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9")
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode == http.StatusOK {
		t.Error("resp.StatusCode == http.StatusOK")
	}

	// invalid option key
	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", "/api/v1/assets/foo", strings.NewReader("bar"))
	r.Header.Set("content-length", "3")
	r.Header.Set("content-type", "text/plain")
	r.Header.Set("X-Sabakan-Asset-SHA256", "FCDE2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9")
	r.Header.Set("X-Sabakan-Asset-Options-Version?", "1.0.0")
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode == http.StatusOK {
		t.Error("resp.StatusCode == http.StatusOK")
	}
}

func testHandleAssetsGetAllowedRoles(t *testing.T) {
	t.Parallel()

	m := mock.NewModel()
	handler := newTestServer(m)
	ctx := context.Background()

	err := m.Machine.Register(ctx, []*sabakan.Machine{
		sabakan.NewMachine(sabakan.MachineSpec{
			Serial: "cs1",
			Role:   "cs",
			IPv4:   []string{"10.0.0.1", "10.0.1.1"},
			BMC:    sabakan.MachineBMC{IPv4: "10.1.0.1"},
		}),
		sabakan.NewMachine(sabakan.MachineSpec{
			Serial: "boot1",
			Role:   "boot",
			IPv4:   []string{"10.0.0.2"},
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = m.Asset.Put(ctx, "secret", "text/plain", nil, map[string]string{"allowed-roles": "cs,ss"}, strings.NewReader("bar"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = m.Asset.Put(ctx, "public", "text/plain", nil, nil, strings.NewReader("bar"))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		method     string
		asset      string
		remoteAddr string
		wantStatus int
	}{
		{"allowed role", "GET", "secret", "10.0.0.1:1234", http.StatusOK},
		{"allowed role with another address", "GET", "secret", "10.0.1.1:1234", http.StatusOK},
		{"allowed role with HEAD", "HEAD", "secret", "10.0.0.1:1234", http.StatusOK},
		{"allowed role with IPv4-mapped IPv6", "GET", "secret", "[::ffff:10.0.0.1]:1234", http.StatusOK},
		{"denied role", "GET", "secret", "10.0.0.2:1234", http.StatusForbidden},
		{"denied role with HEAD", "HEAD", "secret", "10.0.0.2:1234", http.StatusForbidden},
		{"BMC address", "GET", "secret", "10.1.0.1:1234", http.StatusForbidden},
		{"unknown address", "GET", "secret", "10.0.0.9:1234", http.StatusForbidden},
		{"localhost", "GET", "secret", "127.0.0.1:1234", http.StatusForbidden},
		{"no restriction", "GET", "public", "10.0.0.9:1234", http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.method, "/api/v1/assets/"+tt.asset, nil)
			r.RemoteAddr = tt.remoteAddr
			handler.ServeHTTP(w, r)

			resp := w.Result()
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
			if tt.wantStatus != http.StatusOK {
				return
			}
			wantCacheControl := ""
			if tt.asset == "secret" {
				wantCacheControl = "no-store"
			}
			if got := resp.Header.Get("Cache-Control"); got != wantCacheControl {
				t.Errorf("Cache-Control = %q, want %q", got, wantCacheControl)
			}
			if tt.method != "GET" {
				return
			}
			data, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != "bar" {
				t.Errorf("body = %q, want %q", data, "bar")
			}
		})
	}
}

func testHandleAssetsPutAllowedRoles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		header     string
		value      string
		wantStatus int
	}{
		{"single role", "X-Sabakan-Asset-Options-Allowed-Roles", "cs", http.StatusCreated},
		{"multiple roles", "X-Sabakan-Asset-Options-Allowed-Roles", "cs,ss,ss2", http.StatusCreated},
		{"empty", "X-Sabakan-Asset-Options-Allowed-Roles", "", http.StatusBadRequest},
		{"empty item", "X-Sabakan-Asset-Options-Allowed-Roles", "cs,,ss", http.StatusBadRequest},
		{"trailing comma", "X-Sabakan-Asset-Options-Allowed-Roles", "cs,", http.StatusBadRequest},
		{"invalid role", "X-Sabakan-Asset-Options-Allowed-Roles", "cs;ss", http.StatusBadRequest},
		{"comma in another option", "X-Sabakan-Asset-Options-Version", "1,2", http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := mock.NewModel()
			handler := newTestServer(m)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("PUT", "/api/v1/assets/foo", strings.NewReader("bar"))
			r.Header.Set("content-length", "3")
			r.Header.Set("content-type", "text/plain")
			r.Header.Set(tt.header, tt.value)
			handler.ServeHTTP(w, r)

			resp := w.Result()
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
			if tt.wantStatus != http.StatusCreated {
				return
			}
			asset, err := m.Asset.GetInfo(context.Background(), "foo")
			if err != nil {
				t.Fatal(err)
			}
			if asset.Options["allowed-roles"] != tt.value {
				t.Errorf("allowed-roles = %q, want %q", asset.Options["allowed-roles"], tt.value)
			}
		})
	}
}

func testHandleAssetsDelete(t *testing.T) {
	t.Parallel()

	m := mock.NewModel()
	handler := newTestServer(m)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("DELETE", "/api/v1/assets/foo", nil)
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatal("resp.StatusCode != http.StatusNotFound:", resp.StatusCode)
	}

	_, err := m.Asset.Put(context.Background(), "foo", "text/plain", nil, nil, strings.NewReader("bar"))
	if err != nil {
		t.Fatal(err)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("DELETE", "/api/v1/assets/foo", nil)
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("resp.StatusCode != http.StatusOK:", resp.StatusCode)
	}
}

func TestHandleAssets(t *testing.T) {
	t.Run("GetIndex", testHandleAssetsGetIndex)
	t.Run("GetInfo", testHandleAssetsGetInfo)
	t.Run("Get", testHandleAssetsGet)
	t.Run("GetAllowedRoles", testHandleAssetsGetAllowedRoles)
	t.Run("Put", testHandleAssetsPut)
	t.Run("PutAllowedRoles", testHandleAssetsPutAllowedRoles)
	t.Run("Delete", testHandleAssetsDelete)
}
