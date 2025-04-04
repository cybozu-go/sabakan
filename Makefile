# Makefile for sabakan

# configuration variables
ETCD_VERSION = 3.5.19
GO_FILES=$(shell find -name '*.go' -not -name '*_test.go')
BUILT_TARGET=sabakan sabactl sabakan-cryptsetup
IMAGE ?= ghcr.io/cybozu-go/sabakan
TAG ?= latest
CFSSL_VER = 1.6.5
CFSSL = /usr/local/bin/cfssl
CFSSLJSON = /usr/local/bin/cfssljson
E2E_OUTPUT=./e2e/output

.PHONY: all
all: build

.PHONY: build
build: $(BUILT_TARGET)
$(BUILT_TARGET): $(GO_FILES)
	CGO_ENABLED=0 go build -ldflags="-s -w" ./pkg/$@

.PHONY: check-generate
check-generate:
	go mod tidy
	git diff --exit-code --name-only

.PHONY: code-check
code-check:
	test -z "$$(gofmt -s -l . | tee /dev/stderr)"
	staticcheck ./...
	test -z "$$(custom-checker -restrictpkg.packages=html/template,log $$(go list -tags='$(GOTAGS)' ./...) 2>&1 | tee /dev/stderr)"
	go vet ./...

.PHONY: test
test:
	go test -race -v ./...

.PHONY: e2e
e2e: build
	cd e2e/certs && ./gencerts.sh
	RUN_E2E=1 go test -v -count=1 ./e2e

.PHONY: clean
clean:
	rm -f $(BUILT_TARGET)
	rm -rf $(E2E_OUTPUT)

.PHONY: test-tools
test-tools: custom-checker staticcheck etcd

.PHONY: custom-checker
custom-checker:
	if ! which custom-checker >/dev/null; then \
		env GOFLAGS= go install github.com/cybozu-go/golang-custom-analyzer/cmd/custom-checker@latest; \
	fi

.PHONY: staticcheck
staticcheck:
	if ! which staticcheck >/dev/null; then \
		env GOFLAGS= go install honnef.co/go/tools/cmd/staticcheck@latest; \
	fi

.PHONY: etcd
etcd:
	if ! which etcd >/dev/null; then \
		curl -L https://github.com/etcd-io/etcd/releases/download/v${ETCD_VERSION}/etcd-v${ETCD_VERSION}-linux-amd64.tar.gz -o /tmp/etcd-v${ETCD_VERSION}-linux-amd64.tar.gz; \
		mkdir /tmp/etcd; \
		tar xzvf /tmp/etcd-v${ETCD_VERSION}-linux-amd64.tar.gz -C /tmp/etcd --strip-components=1; \
		$(SUDO) mv /tmp/etcd/etcd /usr/local/bin/; \
		rm -rf /tmp/etcd-v${ETCD_VERSION}-linux-amd64.tar.gz /tmp/etcd; \
	fi

.PHONY: docker-build
docker-build: build
	cp LICENSE ./docker
	cp ./sabakan ./sabactl ./sabakan-cryptsetup ./docker
	docker build --no-cache -t $(IMAGE):$(TAG) ./docker
	rm ./docker/sabactl ./docker/sabakan ./docker/sabakan-cryptsetup ./docker/LICENSE

.PHONY: setup-cfssl
setup-cfssl:
	if ! [ -f $(CFSSL) -a -f $(CFSSLJSON) ]; then \
		curl -sSLf -o cfssl https://github.com/cloudflare/cfssl/releases/download/v$(CFSSL_VER)/cfssl_$(CFSSL_VER)_linux_amd64; \
		curl -sSLf -o cfssljson https://github.com/cloudflare/cfssl/releases/download/v$(CFSSL_VER)/cfssljson_$(CFSSL_VER)_linux_amd64; \
		chmod +x cfssl cfssljson; \
		$(SUDO) mv cfssl cfssljson /usr/local/bin/; \
	fi
