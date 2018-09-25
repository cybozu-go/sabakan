# Makefile for integration testing using ginkgo

# configuration variables
BRIDGE_ADDRESS = 10.0.0.1
HOST1 = 10.0.0.101
HOST2 = 10.0.0.102
HOST3 = 10.0.0.103
WORKER = 10.0.0.104
PLACEMAT = $(GOPATH)/bin/placemat
GINKGO = $(GOPATH)/bin/ginkgo
ETCD_VERSION = 3.3.8
PLACEMAT_DATADIR = /var/scratch/placemat

COREOS_VERSION := 1745.5.0
COREOS_URL := http://stable.release.core-os.net/amd64-usr
COREOS_KERNEL := coreos_production_pxe.vmlinuz
COREOS_INITRD := coreos_production_pxe_image.cpio.gz

export BRIDGE_ADDRESS HOST1 HOST2 HOST3 WORKER PLACEMAT GINKGO
export COREOS_VERSION COREOS_KERNEL COREOS_INITRD

# non-configuration variables
ETCD_ARCHIVE := etcd-v$(ETCD_VERSION)-linux-amd64.tar.gz
SSH_PRIVKEY := ./mtest_key
OUTPUT := ./output
UBUNTU_IMAGE := ubuntu-18.04-server-cloudimg-amd64.img
SABACTL := $(OUTPUT)/sabactl
ETCDCTL := $(OUTPUT)/etcdctl
MACHINES_JSON := $(OUTPUT)/machines.json
IPAM_JSON := $(OUTPUT)/ipam.json
DHCP_JSON := $(OUTPUT)/dhcp.json
IGNITIONS := $(OUTPUT)/ignitions

export SSH_PRIVKEY SABACTL ETCDCTL MACHINES_JSON IPAM_JSON DHCP_JSON IGNITIONS


GENERATED_FILES = $(OUTPUT)/etcd  $(OUTPUT)/etcdctl \
	$(OUTPUT)/sabakan $(OUTPUT)/sabactl \
	$(OUTPUT)/cluster.yml \
	$(OUTPUT)/user-data1.yml \
	$(OUTPUT)/user-data2.yml \
	$(OUTPUT)/user-data3.yml \
	$(OUTPUT)/network1.yml \
	$(OUTPUT)/network2.yml \
	$(OUTPUT)/network3.yml \
	$(OUTPUT)/dhcp.json \
	$(OUTPUT)/ipam.json \
	$(OUTPUT)/machines.json \
	$(OUTPUT)/ignitions \
	$(UBUNTU_IMAGE)

all: test

$(ETCD_ARCHIVE):
	curl -sSLf -o $@ https://github.com/coreos/etcd/releases/download/v$(ETCD_VERSION)/etcd-v$(ETCD_VERSION)-linux-amd64.tar.gz

$(UBUNTU_IMAGE):
	curl -sSLf -o $@ https://cloud-images.ubuntu.com/releases/18.04/release/$@

$(COREOS_KERNEL) $(COREOS_INITRD):
	curl -fSL -o $@ $(COREOS_URL)/$(COREOS_VERSION)/$@

$(OUTPUT)/etcd $(OUTPUT)/etcdctl: $(ETCD_ARCHIVE)
	mkdir -p $(OUTPUT)
	tar -C $(OUTPUT) -x -z --wildcards --strip-components 1 -f $< '*/$(notdir $@)'
	touch $@

$(OUTPUT)/sabakan $(OUTPUT)/sabactl:
	mkdir -p $(OUTPUT)
	cd ..; \
		go get -d ./... && \
		GOBIN=$(realpath $(OUTPUT)) go install ./cmd/$(notdir $@)

$(OUTPUT)/cluster.yml: cluster.yml
	mkdir -p $(OUTPUT)
	sed -e 's/@BRIDGE_ADDRESS@/$(BRIDGE_ADDRESS)/' \
		-e 's/@BMC_ADDRESS@/$(BMC_ADDRESS)/' \
		-e 's/@COREOS_KERNEL@/$(COREOS_KERNEL)/' \
		-e 's/@COREOS_INITRD@/$(COREOS_INITRD)/' $< > $@

$(OUTPUT)/user-data1.yml: user-data.yml
	mkdir -p $(OUTPUT)
	sed -e 's/@HOST1@/$(HOST1)/' \
			-e 's/@HOST@/$(HOST1)/' $< > $@

$(OUTPUT)/user-data2.yml: user-data.yml
	mkdir -p $(OUTPUT)
	sed -e 's/@HOST1@/$(HOST1)/' \
			-e 's/@HOST@/$(HOST2)/' $< > $@

$(OUTPUT)/user-data3.yml: user-data.yml
	mkdir -p $(OUTPUT)
	sed -e 's/@HOST1@/$(HOST1)/' \
			-e 's/@HOST@/$(HOST3)/' $< > $@

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

test: $(GENERATED_FILES) $(COREOS_KERNEL) $(COREOS_INITRD)
	sudo rm -rf $(PLACEMAT_DATADIR)/volumes/host1
	sudo rm -rf $(PLACEMAT_DATADIR)/volumes/host2
	sudo rm -rf $(PLACEMAT_DATADIR)/volumes/host3
	sudo rm -rf $(PLACEMAT_DATADIR)/volumes/worker
	./test.sh

placemat: $(GENERATED_FILES) $(COREOS_KERNEL) $(COREOS_INITRD)
	sudo rm -rf $(PLACEMAT_DATADIR)/volumes/host1
	sudo rm -rf $(PLACEMAT_DATADIR)/volumes/host2
	sudo rm -rf $(PLACEMAT_DATADIR)/volumes/host3
	sudo rm -rf $(PLACEMAT_DATADIR)/volumes/worker
	sudo -b $(PLACEMAT) output/cluster.yml
	@echo 'Placemat is running.  To kill it, do "sudo pkill placemat".'

test-light:
	$(GINKGO)

clean:
	rm -rf $(OUTPUT)

setup:
	go get -t -d ./...
	go get -u github.com/onsi/ginkgo/ginkgo
	go get -u github.com/cybozu-go/placemat/...

.PHONY:	all clean setup test
