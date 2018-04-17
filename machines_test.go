package sabakan

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"
)

func PostConfig(etcdClient EtcdClient) {
	config := Config{
		NodeIPv4Offset: "10.0.0.0/16",
		NodeRackShift:  4,
		BMCIPv4Offset:  "10.1.0.0/16",
		BMCRackShift:   2,
		NodeIPPerNode:  3,
		BMCIPPerNode:   1,
	}
	val, _ := json.Marshal(config)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "localhost:8888/api/v1/config", bytes.NewBuffer(val))
	etcdClient.handlePostConfig(w, r)
}

func PostMachines(etcdClient EtcdClient) ([]Machine, *http.Response) {
	mcs := make([]Machine, 2)
	mcs[0] = Machine{
		Serial:           "1234abcd",
		Product:          "R740",
		Datacenter:       "dc1",
		Rack:             2,
		NodeNumberOfRack: 1,
		Role:             "boot",
		Cluster:          "apac",
	}
	mcs[1] = Machine{
		Serial:           "5678efgh",
		Product:          "R640",
		Datacenter:       "dc1",
		Rack:             2,
		NodeNumberOfRack: 2,
		Role:             "node",
		Cluster:          "emea",
	}
	val, _ := json.Marshal(mcs)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "localhost:8888/api/v1/machines", bytes.NewBuffer(val))
	etcdClient.handlePostMachines(w, r)
	return mcs, w.Result()
}

func TestHandlePostMachines(t *testing.T) {
	etcd, _ := newEtcdClient()
	defer etcd.Close()
	prefix := path.Join(*flagEtcdPrefix, t.Name())
	ctx := context.Background()
	mi, _ := Indexing(ctx, etcd, prefix)
	etcdClient := EtcdClient{etcd, prefix, mi}

	PostConfig(etcdClient)
	mcs, resp := PostMachines(etcdClient)

	var savedMachine Machine
	for i := 0; i < 2; i++ {
		etcdResp, _ := etcd.Get(context.Background(), path.Join(prefix, EtcdKeyMachines, mcs[i].Serial))
		json.Unmarshal(etcdResp.Kvs[0].Value, &savedMachine)
		if resp.StatusCode != 200 {
			t.Fatal("expected: 200, actual:", resp.StatusCode)
		}
		if savedMachine.Serial != mcs[i].Serial {
			t.Fatal("saved machine value not found, ", mcs[i].Serial)
		}
		if savedMachine.Product != mcs[i].Product {
			t.Fatal("saved machine value not found, ", mcs[i].Product)
		}
		if savedMachine.Datacenter != mcs[i].Datacenter {
			t.Fatal("saved machine value not found, ", mcs[i].Datacenter)
		}
		if savedMachine.Rack != mcs[i].Rack {
			t.Fatal("saved machine value not found, ", mcs[i].Rack)
		}
		if savedMachine.NodeNumberOfRack != mcs[i].NodeNumberOfRack {
			t.Fatal("saved machine value not found, ", mcs[i].NodeNumberOfRack)
		}
		if savedMachine.Role != mcs[i].Role {
			t.Fatal("saved machine value not found, ", mcs[i].Role)
		}
		if savedMachine.Cluster != mcs[i].Cluster {
			t.Fatal("saved machine value not found, ", mcs[i].Cluster)
		}
	}
}

func TestHandleGetAndPutMachines(t *testing.T) {
	etcd, _ := newEtcdClient()
	defer etcd.Close()
	prefix := path.Join(*flagEtcdPrefix, t.Name())
	ctx := context.Background()
	mi, _ := Indexing(ctx, etcd, prefix)

	etcdClient := EtcdClient{etcd, prefix, mi}
	PostConfig(etcdClient)
	mcs, _ := PostMachines(etcdClient)
	for i := 0; i < 2; i++ {
		etcdResp, _ := etcd.Get(context.Background(), path.Join(prefix, EtcdKeyMachines, mcs[i].Serial))
		err := mi.AddIndex(etcdResp.Kvs[0].Value)
		if err != nil {
			t.Fatal("Failed to add index, ", err.Error())
		}
	}

	querySingleValue := map[int][]string{
		0: []string{
			"?serial=1234abcd",
			"?ipv4=10.0.0.97",
			"?ipv4=10.0.0.33",
			"?ipv4=10.0.0.65",
		},
		1: []string{
			"?serial=5678efgh",
			"?ipv4=10.0.0.98",
			"?ipv4=10.0.0.34",
			"?ipv4=10.0.0.66",
		},
	}

	// Test replying serial
	for i := 0; i < len(mcs); i++ {
		var respMachine Machine
		for n := 0; n < len(querySingleValue[i]); n++ {
			q := "?serial=" + mcs[i].Serial
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "localhost:8888/api/v1/machines"+q, nil)
			etcdClient.handleGetMachines(w, r)

			resp := w.Result()
			json.NewDecoder(resp.Body).Decode(&respMachine)

			if resp.StatusCode != 200 {
				t.Fatal("expected: 200, actual:", resp.StatusCode)
			}
			if respMachine.Serial != mcs[i].Serial {
				t.Fatal("machine value not found, ", mcs[i].Serial)
			}
			if respMachine.Product != mcs[i].Product {
				t.Fatal("machine value not found, ", mcs[i].Product)
			}
			if respMachine.Datacenter != mcs[i].Datacenter {
				t.Fatal("machine value not found, ", mcs[i].Datacenter)
			}
			if respMachine.Rack != mcs[i].Rack {
				t.Fatal("machine value not found, ", mcs[i].Rack)
			}
			if respMachine.NodeNumberOfRack != mcs[i].NodeNumberOfRack {
				t.Fatal("machine value not found, ", mcs[i].NodeNumberOfRack)
			}
			if respMachine.Role != mcs[i].Role {
				t.Fatal("machine value not found, ", mcs[i].Role)
			}
			if respMachine.Cluster != mcs[i].Cluster {
				t.Fatal("machine value not found, ", mcs[i].Cluster)
			}
		}
	}

	// Test replying single value array
	querySingleValueArray := map[int][]string{
		0: []string{
			"?product=R740",
			"?role=boot",
			"?cluster=apac",
			"?datacenter=dc1&product=R740",
			"?datacenter=dc1&role=boot",
			"?datacenter=dc1&cluster=apac",
			"?rack=2&product=R740",
			"?rack=2&role=boot",
			"?rack=2&cluster=apac",
		},
		1: []string{
			"?product=R640",
			"?role=node",
			"?cluster=emea",
		},
	}

	for i := 0; i < len(mcs); i++ {
		var respMachines []Machine
		for n := 0; n < len(querySingleValueArray[i]); n++ {
			q := querySingleValueArray[i][n]
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "localhost:8888/api/v1/machines"+q, nil)

			etcdClient.handleGetMachines(w, r)

			resp := w.Result()
			json.NewDecoder(resp.Body).Decode(&respMachines)

			if resp.StatusCode != 200 {
				t.Fatal("expected: 200, actual:", resp.StatusCode)
			}
			if respMachines[0].Serial != mcs[i].Serial {
				t.Fatal("machine value not found, ", mcs[i].Serial)
			}
			if respMachines[0].Product != mcs[i].Product {
				t.Fatal("machine value not found, ", mcs[i].Product)
			}
			if respMachines[0].Datacenter != mcs[i].Datacenter {
				t.Fatal("machine value not found, ", mcs[i].Datacenter)
			}
			if respMachines[0].Rack != mcs[i].Rack {
				t.Fatal("machine value not found, ", mcs[i].Rack)
			}
			if respMachines[0].NodeNumberOfRack != mcs[i].NodeNumberOfRack {
				t.Fatal("machine value not found, ", mcs[i].NodeNumberOfRack)
			}
			if respMachines[0].Role != mcs[i].Role {
				t.Fatal("machine value not found, ", mcs[i].Role)
			}
			if respMachines[0].Cluster != mcs[i].Cluster {
				t.Fatal("machine value not found, ", mcs[i].Cluster)
			}
		}
	}

	// Test replying multiple values array
	queryMultipleValuesArray := []string{
		"?datacenter=dc1",
		"?rack=2",
	}

	for n := 0; n < len(queryMultipleValuesArray); n++ {
		var respMachines []Machine
		q := queryMultipleValuesArray[n]
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "localhost:8888/api/v1/machines"+q, nil)

		etcdClient.handleGetMachines(w, r)

		resp := w.Result()
		json.NewDecoder(resp.Body).Decode(&respMachines)

		if resp.StatusCode != 200 {
			t.Fatal("expected: 200, actual:", resp.StatusCode)
		}
		for i := 0; i < len(mcs); i++ {
			if respMachines[i].Serial != mcs[i].Serial {
				t.Fatal("machine value not found, ", mcs[i].Serial)
			}
			if respMachines[i].Product != mcs[i].Product {
				t.Fatal("machine value not found, ", mcs[i].Product)
			}
			if respMachines[i].Datacenter != mcs[i].Datacenter {
				t.Fatal("machine value not found, ", mcs[i].Datacenter)
			}
			if respMachines[i].Rack != mcs[i].Rack {
				t.Fatal("machine value not found, ", mcs[i].Rack)
			}
			if respMachines[i].NodeNumberOfRack != mcs[i].NodeNumberOfRack {
				t.Fatal("machine value not found, ", mcs[i].NodeNumberOfRack)
			}
			if respMachines[i].Role != mcs[i].Role {
				t.Fatal("machine value not found, ", mcs[i].Role)
			}
			if respMachines[i].Cluster != mcs[i].Cluster {
				t.Fatal("machine value not found, ", mcs[i].Cluster)
			}
		}
	}

	// Test replying status code 404
	queryNotFound := []string{
		"?serial=blahblah",
		"?ipv4=192.168.1.1",
	}

	for n := 0; n < len(queryNotFound); n++ {
		var respMachines []Machine
		q := queryNotFound[n]
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "localhost:8888/api/v1/machines"+q, nil)

		etcdClient.handleGetMachines(w, r)

		resp := w.Result()
		json.NewDecoder(resp.Body).Decode(&respMachines)

		if resp.StatusCode != 404 {
			t.Fatal("expected: 404, actual:", resp.StatusCode)
		}
		if len(respMachines) != 0 {
			t.Fatal("expected: 0 actual:", len(respMachines))
		}
	}

	// Test replying empty array with status code 200
	queryEmptyArray := []string{
		"?product=EPYC",
		"?datacenter=abc",
		"?rack=100",
		"?role=misc",
		"?cluster=uswest",
	}

	for n := 0; n < len(queryEmptyArray); n++ {
		var respMachines []Machine
		q := queryEmptyArray[n]
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "localhost:8888/api/v1/machines"+q, nil)

		etcdClient.handleGetMachines(w, r)

		resp := w.Result()
		json.NewDecoder(resp.Body).Decode(&respMachines)

		if resp.StatusCode != 200 {
			t.Fatal("expected: 200, actual:", resp.StatusCode)
		}
		if len(respMachines) != 0 {
			t.Fatal("expected: 0 actual:", len(respMachines))
		}
	}

	val, _ := json.Marshal([]map[string]interface{}{
		{
			"serial":              "1234abcd",
			"datacenter":          "ny1",
			"node-number-of-rack": 5,
		},
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "localhost:8888/api/v1/machines", bytes.NewBuffer(val))

	etcdClient.handlePutMachines(w, r)

	q := "?serial=1234abcd"
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "localhost:8888/api/v1/machines"+q, nil)

	etcdClient.handleGetMachines(w, r)

	var respMachine Machine
	resp := w.Result()
	json.NewDecoder(resp.Body).Decode(&respMachine)

	if resp.StatusCode != 200 {
		t.Fatal("expected: 200, actual:", resp.StatusCode)
	}
	if respMachine.Serial != mcs[0].Serial {
		t.Fatal("machine value not found, ", mcs[0].Serial)
	}
	if respMachine.Product != mcs[0].Product {
		t.Fatal("machine value not found, ", mcs[0].Product)
	}
	if respMachine.Datacenter != "ny1" {
		t.Fatal("machine value not found, ", mcs[0].Datacenter)
	}
	if respMachine.Rack != mcs[0].Rack {
		t.Fatal("machine value not found, ", mcs[0].Rack)
	}
	if respMachine.NodeNumberOfRack != 5 {
		t.Fatal("machine value not found, ", mcs[0].NodeNumberOfRack)
	}
	if respMachine.Role != mcs[0].Role {
		t.Fatal("machine value not found, ", mcs[0].Role)
	}
	if respMachine.Cluster != mcs[0].Cluster {
		t.Fatal("machine value not found, ", mcs[0].Cluster)
	}
}
