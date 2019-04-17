# Makefile for integration testing using ginkgo

# configuration variables
BRIDGE_ADDRESS = 10.0.0.1
HOST1 = 10.0.0.101
HOST2 = 10.0.0.102
HOST3 = 10.0.0.103
WORKER = 10.0.0.104
PLACEMAT = /usr/bin/placemat
GINKGO = env GOFLAGS=-mod=vendor $(GOPATH)/bin/ginkgo
ETCD_VERSION = 3.3.12
PLACEMAT_DATADIR = /var/scratch/placemat
COREOS_VERSION := 2023.5.0
COREOS_KERNEL := $(abspath .)/coreos_production_pxe.vmlinuz
COREOS_INITRD := $(abspath .)/coreos_production_pxe_image.cpio.gz
SUDO = sudo
SUITE = ./functions
ifeq ($(findstring /,$(SUITE)),)
  SUITE_PACKAGE = ./$(SUITE)
else
  SUITE_PACKAGE = $(SUITE)
endif

export BRIDGE_ADDRESS HOST1 HOST2 HOST3 WORKER PLACEMAT GINKGO SUITE_PACKAGE
export COREOS_VERSION COREOS_KERNEL COREOS_INITRD

### for Go
GOFLAGS = -mod=vendor
export GOFLAGS

# non-configuration variables
ETCD_ARCHIVE := etcd-v$(ETCD_VERSION)-linux-amd64.tar.gz
SSH_PRIVKEY = $(realpath ./mtest_key)
OUTPUT := ./output
COREOS_IMAGE := coreos_production_qemu_image.img
COREOS_IMAGE_DISK_SIZE = 5G
SABAKAN_IMAGE := $(abspath $(OUTPUT))/sabakan.img
ETCD := $(abspath $(OUTPUT))/etcd
ETCDCTL := $(abspath $(OUTPUT))/etcdctl
MACHINES_JSON := $(abspath $(OUTPUT))/machines.json
IPAM_JSON := $(abspath $(OUTPUT))/ipam.json
DHCP_JSON := $(abspath $(OUTPUT))/dhcp.json
IGNITIONS := $(abspath $(OUTPUT))/ignitions

export SSH_PRIVKEY SABAKAN_IMAGE ETCD ETCDCTL MACHINES_JSON IPAM_JSON DHCP_JSON IGNITIONS

TEST_DEPS = \
	$(COREOS_INITRD) \
	$(COREOS_KERNEL) \
	$(OUTPUT)/dhcp.json \
	$(OUTPUT)/ipam.json \
	$(OUTPUT)/machines.json \
	$(OUTPUT)/sabakan.img

PLACEMAT_DEPS = \
	$(OUTPUT)/cluster.yml \
	$(OUTPUT)/etcd \
	$(OUTPUT)/etcdctl \
	$(OUTPUT)/ignitions \
	$(OUTPUT)/network1.yml \
	$(OUTPUT)/network2.yml \
	$(OUTPUT)/network3.yml \
	$(OUTPUT)/user-data1.yml \
	$(OUTPUT)/user-data2.yml \
	$(OUTPUT)/user-data3.yml \
	$(OUTPUT)/$(COREOS_IMAGE)

all:
	@echo "Specify one of these targets:"
	@echo
	@echo "    placemat - run VMs in placemat."
	@echo "    stop     - stop placemat."
	@echo "    test     - run test in placemat env."
	@echo "    setup    - install dependencies."
	@echo
	@echo "'make test' accepts SUITE."

$(ETCD_ARCHIVE):
	curl -sSLf -o $@ https://github.com/coreos/etcd/releases/download/v$(ETCD_VERSION)/etcd-v$(ETCD_VERSION)-linux-amd64.tar.gz

$(COREOS_IMAGE):
	curl -sSLf -o $@.bz2 https://stable.release.core-os.net/amd64-usr/current/$@.bz2
	bunzip2 -q $@.bz2

$(OUTPUT)/$(COREOS_IMAGE): $(COREOS_IMAGE)
	mkdir -p $(OUTPUT)
	rm -f $@
	cp $< $@
	qemu-img resize $@ $(COREOS_IMAGE_DISK_SIZE)
	touch $@

$(COREOS_KERNEL) $(COREOS_INITRD):
	curl -fSL -o $@ http://stable.release.core-os.net/amd64-usr/$(COREOS_VERSION)/$(notdir $@)

$(OUTPUT)/etcd $(OUTPUT)/etcdctl: $(ETCD_ARCHIVE)
	mkdir -p $(OUTPUT)
	tar -C $(OUTPUT) -x -z --wildcards --strip-components 1 -f $< '*/$(notdir $@)'
	touch $@

$(OUTPUT)/sabakan.img:
	GOBIN=$(shell pwd)/../docker go install ../pkg/sabakan
	GOBIN=$(shell pwd)/../docker go install ../pkg/sabakan-cryptsetup
	GOBIN=$(shell pwd)/../docker go install ../pkg/sabactl
	cp ../LICENSE $(shell pwd)/../docker
	sudo podman build --no-cache --rm=false -t quay.io/cybozu/sabakan:dev ../docker/
	mkdir -p $(OUTPUT)
	sudo podman save -o $@ quay.io/cybozu/sabakan:dev

$(OUTPUT)/cluster.yml: cluster.yml
	mkdir -p $(OUTPUT)
	sed -e 's/@BRIDGE_ADDRESS@/$(BRIDGE_ADDRESS)/' \
		-e 's/@BMC_ADDRESS@/$(BMC_ADDRESS)/' \
		-e 's/@COREOS_KERNEL@/$(notdir $(COREOS_KERNEL))/' \
		-e 's/@COREOS_INITRD@/$(notdir $(COREOS_INITRD))/' $< > $@

$(OUTPUT)/network1.yml: network.yml
	mkdir -p $(OUTPUT)
	sed -e 's/@BRIDGE_ADDRESS@/$(BRIDGE_ADDRESS)/' \
            -e 's/@HOST@/$(HOST1)/' $< > $@

$(OUTPUT)/network2.yml: network.yml
	mkdir -p $(OUTPUT)
	sed -e 's/@BRIDGE_ADDRESS@/$(BRIDGE_ADDRESS)/' \
            -e 's/@HOST@/$(HOST2)/' $< > $@

$(OUTPUT)/network3.yml: network.yml
	mkdir -p $(OUTPUT)
	sed -e 's/@BRIDGE_ADDRESS@/$(BRIDGE_ADDRESS)/' \
            -e 's/@HOST@/$(HOST3)/' $< > $@

$(OUTPUT)/dhcp.json $(OUTPUT)/ipam.json $(OUTPUT)/machines.json:
	ln -f $(notdir $@) $@

$(OUTPUT)/ignitions:
	cp -r ignitions $@
	sed -i -e 's/@BRIDGE_ADDRESS@/$(BRIDGE_ADDRESS)/' \
            -e 's/@WORKER@/$(WORKER)/' $@/networkd/10-eth0.network

$(CT):
	curl -fsSL -o ct https://github.com/coreos/container-linux-config-transpiler/releases/download/v$(CT_VERSION)/ct-v$(CT_VERSION)-x86_64-unknown-linux-gnu
	chmod +x ct
	$(SUDO) mv ct $@

$(OUTPUT)/host.ign: host-ign.yml $(CT)
	mkdir -p $(OUTPUT)
	$(CT) -in-file=$< -pretty -out-file=$@

$(OUTPUT)/node1.ign: $(OUTPUT)/node.ign
	sed -e 's/__HOST1__/$(HOST1)/' \
		-e 's/__HOST__/$(HOST1)/' $< > $@

$(OUTPUT)/node2.ign: $(OUTPUT)/node.ign
	sed -e 's/__HOST1__/$(HOST1)/' \
		-e 's/__HOST__/$(HOST2)/' $< > $@

$(OUTPUT)/node3.ign: $(OUTPUT)/node.ign
	sed -e 's/__HOST1__/$(HOST1)/' \
		-e 's/__HOST__/$(HOST3)/' $< > $@

# Targets
test: $(TEST_DEPS)
	./test.sh "$(TARGET)"

placemat: $(PLACEMAT_DEPS) $(TEST_DEPS)
	sudo rm -rf $(PLACEMAT_DATADIR)
	sudo rm -rf $(VM_SHARED_DIR)
	sudo systemd-run --unit=placemat.service $(PLACEMAT) -enable-virtfs $(abspath output/cluster.yml)
	@echo 'Placemat is running.  To kill it, do "make stop".'

stop:
	sudo systemctl stop placemat.service
	sudo systemctl disable placemat.service
	sudo rm /run/systemd/transient/placemat.service
	sudo systemctl daemon-reload
	sudo systemctl reset-failed placemat.service

clean:
	rm -rf $(OUTPUT)

setup:
	go install github.com/onsi/ginkgo/ginkgo

.PHONY: all test placemat stop clean setup
