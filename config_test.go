package sabakan

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"

	"github.com/coreos/etcd/clientv3"
)

// TODO move this into to non-test
type store struct {
	etcd   *clientv3.Client
	prefix string
}

func (s *store) getConfig(ctx context.Context) (*Config, error) {
	key := path.Join(s.prefix, EtcdKeyConfig)
	resp, err := s.etcd.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, errors.New("no values")
	}

	var conf Config
	err = json.Unmarshal(resp.Kvs[0].Value, &conf)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}

func (s *store) putConfig(ctx context.Context, c *Config) error {
	key := path.Join(s.prefix, EtcdKeyConfig)
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	_, err = s.etcd.Put(ctx, key, string(b))
	return err
}

func TestHandleGetConfig(t *testing.T) {
	etcd, err := newEtcdClient()
	if err != nil {
		t.Fatal(err)
	}
	defer etcd.Close()
	client := etcdClient{client: etcd, prefix: *flagEtcdPrefix + t.Name()}
	store := store{etcd: etcd, prefix: *flagEtcdPrefix + t.Name()}

	r := httptest.NewRequest("GET", "localhost:8888/api/v1/crypts", nil)
	w := httptest.NewRecorder()
	client.handleGetConfig(w, r)
	resp := w.Result()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatal("resp.StatusCode == http.StatusNotFound")
	}

	value := Config{
		NodeIPv4Offset: "10.0.0.0",
		NodeRackShift:  4,
		BMCIPv4Offset:  "10.10.0.0",
		BMCRackShift:   2,
		NodeIPPerNode:  3,
		BMCIPPerNode:   1,
	}
	store.putConfig(context.Background(), &value)

	w = httptest.NewRecorder()
	client.handleGetConfig(w, r)
	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("resp.StatusCode == http.StatusOK")
	}
}

func TestHandlePostConfig(t *testing.T) {
	etcd, err := newEtcdClient()
	if err != nil {
		t.Fatal(err)
	}
	defer etcd.Close()
	client := etcdClient{client: etcd, prefix: *flagEtcdPrefix + t.Name()}
	store := store{etcd: etcd, prefix: *flagEtcdPrefix + t.Name()}

	b := new(bytes.Buffer)
	b.WriteString("{}")
	r := httptest.NewRequest("POST", "localhost:8888/api/v1/crypts", b)
	w := httptest.NewRecorder()
	client.handlePostConfig(w, r)
	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatal("resp.StatusCode == http.StatusBadRequest")
	}

	b = new(bytes.Buffer)
	b.WriteString(`{
		"node-ipv4-offset": "10.0.0.0",
		"node-rack-shift": 5,
		"bmc-ipv4-offset": "10.1.0.0",
		"bmc-rack-shift": 2,
		"node-ip-per-node": 3,
		"bmc-ip-per-node": 1
	}`)
	r = httptest.NewRequest("POST", "localhost:8888/api/v1/crypts", b)
	w = httptest.NewRecorder()
	client.handlePostConfig(w, r)
	resp = w.Result()
	conf, err := store.getConfig(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if conf.NodeIPv4Offset != "10.0.0.0" || conf.NodeRackShift != 5 || conf.BMCIPv4Offset != "10.1.0.0" ||
		conf.BMCRackShift != 2 || conf.NodeIPPerNode != 3 || conf.BMCIPPerNode != 1 {
		t.Fatalf("unexpected config: %v", conf)
	}
}
