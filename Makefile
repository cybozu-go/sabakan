
start:
	sudo systemd-run --slice=machine rkt run --insecure-options=image --port=2379-tcp:2379 --port=2380-tcp:2380 docker://quay.io/cybozu/etcd:3.2 -- --name etcd-1 --advertise-client-urls http://0.0.0.0:2379 --listen-client-urls http://0.0.0.0:2379 --data-dir /var/lib/etcd --initial-cluster etcd-1=http://0.0.0.0:2380 --initial-cluster-state new --listen-peer-urls http://0.0.0.0:2380 --initial-advertise-peer-urls http://0.0.0.0:2380 2>&1 | cut -d " " -f 4 > /tmp/etcd.service

	go build ./cmd/sabakan && sudo ./sabakan
	sudo systemctl stop `cat /tmp/etcd.service`
	rm /tmp/etcd.service
