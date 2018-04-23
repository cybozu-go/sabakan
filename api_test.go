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

func TestMachinesGet(t *testing.T) {
	var method, path, rawQuery string
	machines := []Machine{{Serial: "123abc"}}

	s1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		rawQuery = r.URL.RawQuery
		json.NewEncoder(w).Encode(machines)
	}))
	c := Client{endpoint: s1.URL, http: &cmd.HTTPClient{Client: &http.Client{}}}

	got, err := c.MachinesGet(context.Background(), map[string]string{"serial": "123abc"})
	if err != nil {
		t.Error("err != nil", err)
	}

	if method != "GET" {
		t.Errorf("%s != GET", method)
	}
	expectedPath := "/api/v1/machines"
	if path != expectedPath {
		t.Errorf("%s != %s", path, expectedPath)
	}
	expectedRawQuery := "serial=123abc"
	if rawQuery != expectedRawQuery {
		t.Errorf("%s != %s", rawQuery, expectedRawQuery)
	}
	if !reflect.DeepEqual(got, machines) {
		t.Errorf("%v != %v", got, machines)
	}
}

func TestMachinesCreate(t *testing.T) {
	var method, path string

	s1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
	}))
	c := Client{endpoint: s1.URL, http: &cmd.HTTPClient{Client: &http.Client{}}}

	err := c.MachinesCreate(context.Background(), []Machine{})
	if err != nil {
		t.Error("err == nil")
	}
	if method != "POST" || path != "/api/v1/machines" {
		t.Errorf("%s != POST, nor %s != /api/v1/machines", method, path)
	}
}

func TestMachinesUpdate(t *testing.T) {
	var method, path string

	s1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
	}))
	c := Client{endpoint: s1.URL, http: &cmd.HTTPClient{Client: &http.Client{}}}

	err := c.MachinesUpdate(context.Background(), []Machine{})
	if err != nil {
		t.Error("err == nil")
	}
	if method != "PUT" || path != "/api/v1/machines" {
		t.Errorf("%s != PUT, nor %s != /api/v1/machines", method, path)
	}
}

func TestGetJSON(t *testing.T) {
	s1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{ "apple": "red", "banana": "yellow"}`)
	}))
	defer s1.Close()

	var data = make(map[string]string)

	c := Client{endpoint: s1.URL, http: &cmd.HTTPClient{Client: &http.Client{}}}
	err := c.getJSON(context.Background(), "/", nil, &data)
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

	err = c.getJSON(context.Background(), "/", nil, &data)
	if err == nil {
		t.Errorf("%v != nil", err)
	}
}

func TestSendRequestWithJSON(t *testing.T) {
	var record []byte
	var method string
	s1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		record, _ = ioutil.ReadAll(r.Body)
		method = r.Method
	}))
	defer s1.Close()

	var data = make(map[string]string)

	c := Client{endpoint: s1.URL, http: &cmd.HTTPClient{Client: &http.Client{}}}
	err := c.sendRequestWithJSON(context.Background(), "POST", "/", map[string]string{"apple": "red", "banana": "yellow"})
	if err != nil {
		t.Error(err)
	}
	if strings.TrimSpace(string(record)) != `{"apple":"red","banana":"yellow"}` {
		t.Error("unexpected recorded data: " + string(record))
	}
	if method != "POST" {
		t.Error("unexpected request method:", method)
	}

	s2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, `{ "message": "404 not found" }`)
	}))
	defer s2.Close()

	c = Client{endpoint: s2.URL, http: &cmd.HTTPClient{Client: &http.Client{}}}

	err = c.getJSON(context.Background(), "/", nil, &data)
	if err == nil {
		t.Errorf("%v != nil", err)
	}
}
