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
ETCD_VERSION = 3.6.9
ETCD_SHA256SUM = 633136f13fcadac52e5c0ddcb97912643af6fcb9cb362e75774d0f96f7666396
PLACEMAT_DATADIR = /var/scratch/placemat
COREOS_VERSION := 4459.2.4
COREOS_IMAGE_SHA512SUM := c53e24f5bcf0fa10b4ecf50942267d34ee8c05657eb1750fb031fde541395043829818418905c8c63cf5d335dd592d06bc2478f768819a051dad1c4e41fbe94c
COREOS_KERNEL_SHA512SUM := d95d658bb11ab4a5593d8cba5d6cdd6dcf9e7dcf58ed4ad6b6ec9aedaee3420a6eb64cd0672ea79fcda2a8e5defd9fab0c612c1b604b36e5d4ab7698ef24a778
COREOS_INITRD_SHA512SUM := c4c2d973237d4c22c35b18593821e8660591b9a47a4402cb46d90e1ad565cf5d9c5c1ca1120b851a35c0e0aa50730e80750ab60f9ff037cd725c08436c40c7c2
# The latest CT is 0.9.4, but it is not distributed as a pre-built binary.
CT_VERSION = 0.9.3
CT_SHA256SUM = bb5ccbaddd26e0cfd241309de72ef3011aa718095bd2f94ca19096d98efb6ab8
SUDO = sudo

export HOST1 HOST2 HOST3 WORKER1 WORKER2 PLACEMAT GINKGO
export COREOS_VERSION COREOS_KERNEL COREOS_INITRD

# non-configuration variables
ETCD_ARCHIVE := etcd-v$(ETCD_VERSION)-linux-amd64.tar.gz
SSH_PRIVKEY = $(realpath ./mtest_key)
OUTPUT := ./output
COREOS_IMAGE := flatcar_production_qemu_image.img
COREOS_KERNEL := $(abspath .)/flatcar_production_pxe.vmlinuz
COREOS_INITRD := $(abspath .)/flatcar_production_pxe_image.cpio.gz
SABAKAN_IMAGE := $(abspath $(OUTPUT))/sabakan.img
SABAKAN_IMAGE_NAME := ghcr.io/cybozu-go/sabakan
SABAKAN_IMAGE_TAG := dev
SABAKAN_IMAGE_URL := $(SABAKAN_IMAGE_NAME):$(SABAKAN_IMAGE_TAG)
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
	$(OUTPUT)/ignitions \
	$(IGNITIONS)/files/etc/sabakan/sabakan-tls-ca.crt
	
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
	echo "$(ETCD_SHA256SUM) $@" | sha256sum -c -

$(COREOS_IMAGE):
	curl -sSLf -o $@.bz2 https://stable.release.flatcar-linux.net/amd64-usr/$(COREOS_VERSION)/$@.bz2
	echo "$(COREOS_IMAGE_SHA512SUM) $@.bz2" | sha512sum -c -
	bunzip2 -q $@.bz2

$(COREOS_KERNEL):
	curl -fSL -o $@ http://stable.release.flatcar-linux.net/amd64-usr/$(COREOS_VERSION)/$(notdir $@)
	echo "$(COREOS_KERNEL_SHA512SUM) $@" | sha512sum -c -

$(COREOS_INITRD):
	curl -fSL -o $@ http://stable.release.flatcar-linux.net/amd64-usr/$(COREOS_VERSION)/$(notdir $@)
	echo "$(COREOS_INITRD_SHA512SUM) $@" | sha512sum -c -

$(OUTPUT)/etcd $(OUTPUT)/etcdctl: $(ETCD_ARCHIVE)
	mkdir -p $(OUTPUT)
	tar -C $(OUTPUT) -x -z --wildcards --strip-components 1 -f $< '*/$(notdir $@)'
	touch $@

$(OUTPUT)/sabakan.img:
	make -C ../ IMG=$(SABAKAN_IMAGE_NAME) TAG=$(SABAKAN_IMAGE_TAG) docker-build
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
	echo "$(CT_SHA256SUM) $@" | sha256sum -c -
	chmod +x $@

$(OUTPUT)/host.ign: host-ign.yml $(CT)
	mkdir -p $(OUTPUT)
	$(CT) -strict -in-file=host-ign.yml --files-dir=. -pretty -out-file=$@

$(OUTPUT)/host1.ign: $(OUTPUT)/host.ign
	sed -e 's/__BRIDGE_ADDRESS__/$(BRIDGE_ADDRESS)/' \
		-e 's/__HOST_NAME__/host1/' \
		-e 's/__HOST1__/$(HOST1)/' \
		-e 's/__HOST__/$(HOST1)/g' \
		-e 's/__HOST_ADDRESS__/$(HOST1)/' $< > $@

$(OUTPUT)/host2.ign: $(OUTPUT)/host.ign
	sed -e 's/__BRIDGE_ADDRESS__/$(BRIDGE_ADDRESS)/' \
		-e 's/__HOST_NAME__/host2/' \
		-e 's/__HOST1__/$(HOST1)/' \
		-e 's/__HOST__/$(HOST2)/g' \
		-e 's/__HOST_ADDRESS__/$(HOST2)/' $< > $@

$(OUTPUT)/host3.ign: $(OUTPUT)/host.ign
	sed -e 's/__BRIDGE_ADDRESS__/$(BRIDGE_ADDRESS)/' \
		-e 's/__HOST_NAME__/host3/' \
		-e 's/__HOST1__/$(HOST1)/' \
		-e 's/__HOST__/$(HOST3)/g' \
		-e 's/__HOST_ADDRESS__/$(HOST3)/' $< > $@

$(OUTPUT)/readnvram:
	go build -o $@ ./readnvram/main.go

$(IGNITIONS)/files/etc/sabakan/sabakan-tls-ca.crt:
	mkdir -p $(IGNITIONS)/files/etc/sabakan
	cp $(OUTPUT)/certs/ca.crt $(IGNITIONS)/files/etc/sabakan/sabakan-tls-ca.crt

# Targets
.PHONY: test
test: $(TEST_DEPS)
	./test.sh "$(TARGET)"

.PHONY: placemat
placemat: gencerts $(PLACEMAT_DEPS)
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
	go install github.com/onsi/ginkgo/v2/ginkgo
	sudo make -C ../ setup-cfssl

.PHONY: gencerts
gencerts:
	cd ./certs && ./gencerts.sh
