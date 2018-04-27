package sabakan

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"path"
	"reflect"
	"strings"
	"testing"
)

func testCryptsGet(t *testing.T) {
	m := mock.NewModel()
	handler := Server{m}

	serial := "1"
	diskPath := "exists-path"
	key := "aaa"

	ctx := context.Background()
	err := m.Storage.PutEncryptionKey(ctx, serial, diskPath, []byte(key))
	if err != nil {
		t.Fatal(err)
	}

	testData := []struct {
		path   string
		status int
		key    string
	}{
		{diskPath, 200, key},
		{"not-exist", 404, ""},
	}

	for _, td := range testData {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", path.Join("/api/v1/crypts", serial, td.path), nil)
		handler.ServeHTTP(w, r)
		resp := w.Result()

		if resp.StatusCode != td.status {
			t.Error("wrong status code, expects:", td.status, ", actual:", resp.StatusCode)
		}
		data, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		respKey := string(data)
		if len(td.key) > 0 && td.key != respKey {
			t.Error("wrong key, expects:", td.key, ", actual:", respKey)
		}
	}
}

func testCryptsPut(t *testing.T) {
	m := mock.NewModel()
	handler := Server{m}

	serial := "1"

	testData := []struct {
		path   string
		status int
		key    string
	}{
		{"put-path", 201, "aaa"},
		{"put-path", 409, "bbb"},
		{"another-path", 201, string([]byte{0, 1, 2, 100, 50, 200})},
	}

	for _, td := range testData {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("PUT", path.Join("/api/v1/crypts", serial, td.path),
			strings.NewReader(td.key))
		handler.ServeHTTP(w, r)

		resp := w.Result()
		if resp.StatusCode != td.status {
			t.Error("wrong status code, expects:", td.status, ", actual:", resp.StatusCode)
		}
		data, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != 201 {
			continue
		}

		var respJSON struct {
			Status int    `json:"status"`
			Path   string `json:"path"`
		}
		err = json.Unmarshal(data, &respJSON)
		if err != nil {
			t.Error("invalid JSON:", string(data))
			continue
		}
		if respJSON.Status != 201 {
			t.Error("invalid status in JSON:", respJSON.Status)
		}
		if respJSON.Path != td.path {
			t.Error("invalid path in JSON:", respJSON.Path)
		}

		stored, err := m.Storage.GetEncryptionKey(context.Background(), serial, td.path)
		if err != nil {
			t.Fatal(err)
		}
		storedKey := string(stored)
		if td.key != storedKey {
			t.Error("stored key is wrong, expect:", td.key, ", actual:", storedKey)
		}
	}
}

func testCryptsDelete(t *testing.T) {
	m := mock.NewModel()
	handler := Server{m}

	ctx := context.Background()
	expected := make(map[string]struct{})
	serial := "abc"
	key := "aaa"
	for i := 0; i < 5; i++ {
		diskPath := fmt.Sprintf("path%d", i)
		expected[diskPath] = struct{}{}
		err := m.Storage.PutEncryptionKey(ctx, serial, diskPath, []byte(key))
		if err != nil {
			t.Fatal(err)
		}
	}

	// dummy data to test bug in delete logic.
	serial2 := "abcd"
	err := m.Storage.PutEncryptionKey(ctx, serial2, "path1", []byte(key))
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("DELETE", path.Join("/api/v1/crypts", serial), nil)
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Fatal("expected: 200, actual:", resp.StatusCode)
	}

	var deletedPaths []string
	err = json.NewDecoder(resp.Body).Decode(&deletedPaths)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	actual := make(map[string]struct{})
	for _, p := range deletedPaths {
		actual[p] = struct{}{}
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatal("unexpected response:", deletedPaths)
	}

	for i := 0; i < 5; i++ {
		diskPath := fmt.Sprintf("path%d", i)
		data, err := m.Storage.GetEncryptionKey(ctx, serial, diskPath)
		if err != nil {
			t.Fatal(err)
		}
		if data != nil {
			t.Error(diskPath + " was not deleted")
		}
	}

	data, err := m.Storage.GetEncryptionKey(ctx, serial2, "path1")
	if err != nil {
		t.Fatal(err)
	}
	if data == nil {
		t.Error(serial2 + " must not be deleted")
	}
}

func TestCrypts(t *testing.T) {
	t.Run("Get", testCryptsGet)
	t.Run("Put", testCryptsPut)
	t.Run("Delete", testCryptsDelete)
}
