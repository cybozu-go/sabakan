package sabakan

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/cybozu-go/cmd"
)

func TestRemoteConfigGet(t *testing.T) {
	var method, path string
	conf := &Config{
		NodeIPv4Offset: "10.0.0.0",
		NodeRackShift:  4,
		BMCIPv4Offset:  "10.1.0.0",
		BMCRackShift:   2,
		NodeIPPerNode:  3,
		BMCIPPerNode:   1,
	}
	s1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		json.NewEncoder(w).Encode(conf)
	}))
	c := Client{endpoint: s1.URL, http: &cmd.HTTPClient{Client: &http.Client{}}}

	got, err := c.RemoteConfigGet(context.Background())
	if err != nil {
		t.Error("err == nil")
	}
	if method != "GET" || path != "/api/v1/config" {
		t.Errorf("%s != GET, nor %s != /api/v1/config", method, path)
	}
	if !reflect.DeepEqual(got, conf) {
		t.Errorf("%v != %v", got, conf)
	}

}

func TestRemoteConfigPost(t *testing.T) {
	var method, path string
	var record Config
	conf := Config{
		NodeIPv4Offset: "10.0.0.0",
		NodeRackShift:  4,
		BMCIPv4Offset:  "10.1.0.0",
		BMCRackShift:   2,
		NodeIPPerNode:  3,
		BMCIPPerNode:   1,
	}
	s1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		json.NewDecoder(r.Body).Decode(&record)
	}))
	c := Client{endpoint: s1.URL, http: &cmd.HTTPClient{Client: &http.Client{}}}

	err := c.RemoteConfigSet(context.Background(), &conf)
	if err != nil {
		t.Error("err == nil")
	}
	if method != "POST" || path != "/api/v1/config" {
		t.Errorf("%s != GET, nor %s != /api/v1/config", method, path)
	}
	if !reflect.DeepEqual(record, conf) {
		t.Errorf("%v != %v", record, conf)
	}

}

func TestJSONGet(t *testing.T) {
	s1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{ "apple": "red", "banana": "yellow"}`)
	}))
	defer s1.Close()

	var data = make(map[string]string)

	c := Client{endpoint: s1.URL, http: &cmd.HTTPClient{Client: &http.Client{}}}
	err := c.jsonGet(context.Background(), "/", nil, &data)
	if err != nil {
		t.Error(err)
	}
	if data["apple"] != "red" || data["banana"] != "yellow" {
		t.Error("unexpected data: ", data)
	}

	s2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, `{ "message": "404 not found" }`)
	}))
	defer s2.Close()

	c = Client{endpoint: s2.URL, http: &cmd.HTTPClient{Client: &http.Client{}}}

	err = c.jsonGet(context.Background(), "/", nil, &data)
	if err == nil {
		t.Errorf("%v != nil", err)
	}
}

func TestJSONPost(t *testing.T) {
	var record []byte
	s1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		record, _ = ioutil.ReadAll(r.Body)
	}))
	defer s1.Close()

	var data = make(map[string]string)

	c := Client{endpoint: s1.URL, http: &cmd.HTTPClient{Client: &http.Client{}}}
	err := c.jsonPost(context.Background(), "/", map[string]string{"apple": "red", "banana": "yellow"})
	if err != nil {
		t.Error(err)
	}
	if strings.TrimSpace(string(record)) != `{"apple":"red","banana":"yellow"}` {
		t.Error("unexpected recorded data: " + string(record))
	}

	s2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, `{ "message": "404 not found" }`)
	}))
	defer s2.Close()

	c = Client{endpoint: s2.URL, http: &cmd.HTTPClient{Client: &http.Client{}}}

	err = c.jsonGet(context.Background(), "/", nil, &data)
	if err == nil {
		t.Errorf("%v != nil", err)
	}
}
