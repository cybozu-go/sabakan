# Makefile for integration testing using ginkgo

# configuration variables
BRIDGE_ADDRESS = 10.0.0.1
HOST1 = 10.0.0.101
HOST2 = 10.0.0.102
HOST3 = 10.0.0.103
WORKER1 = 10.0.0.104
WORKER2 = 10.0.0.105
PLACEMAT = /usr/bin/placemat2
GINKGO = $(GOPATH)/bin/ginkgo
ETCD_VERSION = 3.5.5
PLACEMAT_DATADIR = /var/scratch/placemat
COREOS_VERSION := 3227.2.3
COREOS_KERNEL := $(abspath .)/flatcar_production_pxe.vmlinuz
COREOS_INITRD := $(abspath .)/flatcar_production_pxe_image.cpio.gz
CT_VERSION = 0.9.3
SUDO = sudo

export HOST1 HOST2 HOST3 WORKER1 WORKER2 PLACEMAT GINKGO
export COREOS_VERSION COREOS_KERNEL COREOS_INITRD

# non-configuration variables
ETCD_ARCHIVE := etcd-v$(ETCD_VERSION)-linux-amd64.tar.gz
SSH_PRIVKEY = $(realpath ./mtest_key)
OUTPUT := ./output
COREOS_IMAGE := flatcar_production_qemu_image.img
SABAKAN_IMAGE := $(abspath $(OUTPUT))/sabakan.img
SABAKAN_IMAGE_URL := quay.io/cybozu/sabakan:dev
CT := $(abspath $(OUTPUT))/ct
ETCD := $(abspath $(OUTPUT))/etcd
ETCDCTL := $(abspath $(OUTPUT))/etcdctl
MACHINES_JSON := $(abspath $(OUTPUT))/machines.json
IPAM_JSON := $(abspath $(OUTPUT))/ipam.json
DHCP_JSON := $(abspath $(OUTPUT))/dhcp.json
IGNITIONS := $(abspath $(OUTPUT))/ignitions
READ_NVRAM := $(abspath $(OUTPUT))/readnvram

export SSH_PRIVKEY SABAKAN_IMAGE SABAKAN_IMAGE_URL ETCD ETCDCTL MACHINES_JSON IPAM_JSON DHCP_JSON IGNITIONS READ_NVRAM

TEST_DEPS = \
	$(COREOS_INITRD) \
	$(COREOS_KERNEL) \
	$(OUTPUT)/dhcp.json \
	$(OUTPUT)/ipam.json \
	$(OUTPUT)/machines.json \
	$(OUTPUT)/sabakan.img \
	$(OUTPUT)/readnvram

PLACEMAT_DEPS = \
	$(COREOS_IMAGE) \
	$(OUTPUT)/cluster.yml \
	$(OUTPUT)/etcd \
	$(OUTPUT)/etcdctl \
	$(OUTPUT)/host1.ign \
	$(OUTPUT)/host2.ign \
	$(OUTPUT)/host3.ign \
	$(OUTPUT)/ignitions

.PHONY: all
all:
	@echo "Specify one of these targets:"
	@echo
	@echo "    placemat - run VMs in placemat."
	@echo "    stop     - stop placemat."
	@echo "    test     - run test in placemat env."
	@echo "    setup    - install dependencies."
	@echo "    clean    - clean temporary files."

$(ETCD_ARCHIVE):
	curl -sSLf -o $@ https://github.com/etcd-io/etcd/releases/download/v$(ETCD_VERSION)/etcd-v$(ETCD_VERSION)-linux-amd64.tar.gz

$(COREOS_IMAGE):
	curl -sSLf -o $@.bz2 https://stable.release.flatcar-linux.net/amd64-usr/current/$@.bz2
	bunzip2 -q $@.bz2

$(COREOS_KERNEL) $(COREOS_INITRD):
	curl -fSL -o $@ http://stable.release.flatcar-linux.net/amd64-usr/$(COREOS_VERSION)/$(notdir $@)

$(OUTPUT)/etcd $(OUTPUT)/etcdctl: $(ETCD_ARCHIVE)
	mkdir -p $(OUTPUT)
	tar -C $(OUTPUT) -x -z --wildcards --strip-components 1 -f $< '*/$(notdir $@)'
	touch $@

$(OUTPUT)/sabakan.img:
	GOBIN=$(shell pwd)/../docker go install ../pkg/sabakan
	GOBIN=$(shell pwd)/../docker go install ../pkg/sabakan-cryptsetup
	GOBIN=$(shell pwd)/../docker go install ../pkg/sabactl
	cp ../LICENSE $(shell pwd)/../docker
	sudo docker build --no-cache --rm=false -t $(SABAKAN_IMAGE_URL) ../docker/
	mkdir -p $(OUTPUT)
	sudo docker save $(SABAKAN_IMAGE_URL) > $@

$(OUTPUT)/cluster.yml: cluster.yml
	mkdir -p $(OUTPUT)
	sed -e 's/__BRIDGE_ADDRESS__/$(BRIDGE_ADDRESS)/' \
		-e 's#__IGNITIONS_FOLDER__#$(IGNITIONS)#' \
		-e 's/__COREOS_IMAGE__/$(COREOS_IMAGE)/' $< > $@

$(OUTPUT)/dhcp.json $(OUTPUT)/ipam.json $(OUTPUT)/machines.json:
	ln -f $(notdir $@) $@

$(OUTPUT)/ignitions:
	cp -r ignitions $@
	sed -i -e 's/@BRIDGE_ADDRESS@/$(BRIDGE_ADDRESS)/' $@/networkd/10-eth0.network

$(CT):
	curl -fsSL -o $@ https://github.com/flatcar/container-linux-config-transpiler/releases/download/v$(CT_VERSION)/ct-v$(CT_VERSION)-x86_64-unknown-linux-gnu
	chmod +x $@

$(OUTPUT)/host.ign: host-ign.yml $(CT)
	mkdir -p $(OUTPUT)
	$(CT) -strict -in-file=$< -pretty -out-file=$@

$(OUTPUT)/host1.ign: $(OUTPUT)/host.ign
	sed -e 's/__BRIDGE_ADDRESS__/$(BRIDGE_ADDRESS)/' \
		-e 's/__HOST_NAME__/host1/' \
		-e 's/__HOST1__/$(HOST1)/' \
		-e 's/__HOST__/$(HOST1)/' \
		-e 's/__HOST_ADDRESS__/$(HOST1)/' $< > $@

$(OUTPUT)/host2.ign: $(OUTPUT)/host.ign
	sed -e 's/__BRIDGE_ADDRESS__/$(BRIDGE_ADDRESS)/' \
		-e 's/__HOST_NAME__/host2/' \
		-e 's/__HOST1__/$(HOST1)/' \
		-e 's/__HOST__/$(HOST2)/' \
		-e 's/__HOST_ADDRESS__/$(HOST2)/' $< > $@

$(OUTPUT)/host3.ign: $(OUTPUT)/host.ign
	sed -e 's/__BRIDGE_ADDRESS__/$(BRIDGE_ADDRESS)/' \
		-e 's/__HOST_NAME__/host3/' \
		-e 's/__HOST1__/$(HOST1)/' \
		-e 's/__HOST__/$(HOST3)/' \
		-e 's/__HOST_ADDRESS__/$(HOST3)/' $< > $@

$(OUTPUT)/readnvram:
	go build -o $@ ./readnvram/main.go

# Targets
.PHONY: test
test: $(TEST_DEPS)
	./test.sh "$(TARGET)"

.PHONY: placemat
placemat: $(PLACEMAT_DEPS)
	sudo rm -rf $(PLACEMAT_DATADIR)
	sudo rm -rf $(VM_SHARED_DIR)
	sudo systemd-run --unit=placemat.service $(PLACEMAT) $(abspath output/cluster.yml)
	@echo 'Placemat is running.  To kill it, do "make stop".'

.PHONY: stop
stop:
	sudo systemctl stop placemat.service
	sudo systemctl disable placemat.service
	sudo rm /run/systemd/transient/placemat.service
	sudo systemctl daemon-reload
	sudo systemctl reset-failed placemat.service

.PHONY: clean
clean:
	rm -rf $(OUTPUT)

.PHONY: setup
setup:
	go install github.com/onsi/ginkgo/ginkgo
