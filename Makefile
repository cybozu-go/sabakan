BRIDGE_NAME=sabakan
BRIDGE_IP="192.168.0.1"
TAP_NAME=sabakan_client
MAC_ADDRESS="52:54:00:11:22:33"

OVMF_CODE_PATH=/usr/share/OVMF/OVMF_CODE.fd
OVMF_VARS_PATH=/tmp/sabakan-dhcp-debug-vm.fd
ORIGINAL_OVMF_VARS_PATH=/usr/share/OVMF/OVMF_VARS.fd

GO_FILES=$(shell find -name '*.go' -not -name '*_test.go')
BUILT_TARGET=sabakan

.DEFAULT_GOAL := build

etcd-server:
	sudo rkt run --insecure-options=image \
	  --port=2379-tcp:2379 --port=2380-tcp:2380 \
	  docker://quay.io/cybozu/etcd:3.2 \
	  --interactive \
	  -- \
	  --name etcd-1 --data-dir /var/lib/etcd \
	  --advertise-client-urls http://0.0.0.0:2379 --listen-client-urls http://0.0.0.0:2379 \
	  --initial-cluster etcd-1=http://0.0.0.0:2380 --initial-cluster-state new \
	  --listen-peer-urls http://0.0.0.0:2380 --initial-advertise-peer-urls http://0.0.0.0:2380 &

dhcp-debug-network: clean-dhcp-debug-network
	sudo ip link add $(BRIDGE_NAME) type bridge
	sudo ip link set $(BRIDGE_NAME) up
	sudo ip addr add $(BRIDGE_IP)/24 dev $(BRIDGE_NAME)
	sudo ip tuntap add $(TAP_NAME) mode tap
	sudo ip link set $(TAP_NAME) master $(BRIDGE_NAME)
	sudo ip link set $(TAP_NAME) up

build: $(BUILT_TARGET)
$(BUILT_TARGET): $(GO_FILES)
	go build ./cmd/sabakan

$(OVMF_VARS_PATH): $(ORIGINAL_OVMF_VARS_PATH)
	cp $(ORIGINAL_OVMF_VARS_PATH) $(OVMF_VARS_PATH)

dhcp-debug-vm: dhcp-debug-network $(OVMF_CODE_PATH) $(OVMF_VARS_PATH)
	kvm \
	  -drive if=pflash,file=$(OVMF_CODE_PATH),format=raw,readonly \
	  -drive if=pflash,file=$(OVMF_VARS_PATH),format=raw,readonly \
	  -netdev tap,id=net0,ifname=$(TAP_NAME),script=no,downscript=no \
	  -device virtio-net-pci,netdev=net0,romfile=,mac=$(MAC_ADDRESS) &

debug: build etcd-server dhcp-debug-vm
	sh -c "trap '$(MAKE) clean-dhcp-debug-network' EXIT; \
	  sudo ./sabakan -loglevel debug -dhcp-bind 0.0.0.0:67 -etcd-timeout 5s -dhcp-interface $(BRIDGE_NAME)"

clean-dhcp-debug-network:
	sudo ip link del $(TAP_NAME) type bridge || true
	sudo ip link del $(BRIDGE_NAME) type bridge || true

clean: clean-dhcp-debug-network
	rm -rf sabakan

.PHONY: etcd-server dhcp-debug-network build dhcp-debug-vm debug clean-dhcp-debug-network clean
