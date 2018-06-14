package web

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/cybozu-go/sabakan"
	"github.com/cybozu-go/sabakan/models/mock"
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

	_, err = m.Asset.Put(context.Background(), "foo", "text/plain", nil, strings.NewReader("bar"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = m.Asset.Put(context.Background(), "abc", "text/plain", nil, strings.NewReader("bar"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = m.Asset.Put(context.Background(), "xyz", "text/plain", nil, strings.NewReader("bar"))
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

	_, err := m.Asset.Put(context.Background(), "foo", "text/plain", nil, strings.NewReader("bar"))
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

	_, err := m.Asset.Put(context.Background(), "foo", "text/plain", nil, strings.NewReader("bar"))
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

	data, err := ioutil.ReadAll(resp.Body)
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
		pw.Write([]byte("bar"))
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
		pw2.Write([]byte("bar"))
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

	_, err := m.Asset.Put(context.Background(), "foo", "text/plain", nil, strings.NewReader("bar"))
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
	t.Run("Put", testHandleAssetsPut)
	t.Run("Delete", testHandleAssetsDelete)
}
