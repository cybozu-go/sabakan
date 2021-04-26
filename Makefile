# Makefile for sabakan

# configuration variables
ETCD_VERSION = 3.3.25
GO_FILES=$(shell find -name '*.go' -not -name '*_test.go')
BUILT_TARGET=sabakan sabactl sabakan-cryptsetup

.PHONY: all
all: test

.PHONY: build
build: $(BUILT_TARGET)
$(BUILT_TARGET): $(GO_FILES)
	go build ./pkg/$@

.PHONY: code-check
code-check:
	test -z "$$(gofmt -s -l . | tee /dev/stderr)"
	staticcheck ./...
	test -z "$$(nilerr ./... 2>&1 | tee /dev/stderr)"
	test -z "$$(custom-checker -restrictpkg.packages=html/template,log $$(go list -tags='$(GOTAGS)' ./...) 2>&1 | tee /dev/stderr)"
	go vet ./...

.PHONY: test
test: build code-check
	go test -race -v ./...

.PHONY: e2e
e2e: build
	RUN_E2E=1 go test -v -count=1 ./e2e

.PHONY: mod
mod:
	go mod tidy
	git add go.mod go.sum

.PHONY: clean
clean:
	rm -f $(BUILT_TARGET)

.PHONY: test-tools
test-tools: custom-checker staticcheck nilerr etcd

.PHONY: custom-checker
custom-checker:
	if ! which custom-checker >/dev/null; then \
		env GOFLAGS= go install github.com/cybozu/neco-containers/golang/analyzer/cmd/custom-checker@latest; \
	fi

.PHONY: staticcheck
staticcheck:
	if ! which staticcheck >/dev/null; then \
		env GOFLAGS= go install honnef.co/go/tools/cmd/staticcheck@latest; \
	fi

.PHONY: nilerr
nilerr:
	if ! which nilerr >/dev/null; then \
		env GOFLAGS= go install github.com/gostaticanalysis/nilerr/cmd/nilerr@latest; \
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
