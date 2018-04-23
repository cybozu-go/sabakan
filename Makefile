
start:
	sudo systemd-run --slice=machine rkt run quay.io/cybozu/etcd:3.2 -- --name etcd-1 --advertise-client-urls http://0.0.0.0:2379 --listen-client-urls http://0.0.0.0:2379 --data-dir /var/lib/etcd --initial-cluster etcd-1=http://0.0.0.0:2380 --initial-cluster-state new --listen-peer-urls http://0.0.0.0:2380 --initial-advertise-peer-urls http://0.0.0.0:2380 2>&1 | cut -d " " -f 4 > /tmp/etcd.service
	sleep 5

	etcd_ip=`sudo rkt list --format=json | jq -r '.[] | select(.app_names[0] == "etcd" and .state == "running") | .networks[0].ip'`; \
	go run cmd/sabakan/main.go -etcd-servers http://$${etcd_ip}:2379

stop:
	sudo systemctl stop `cat /tmp/etcd.service`
	rm /tmp/etcd.service

