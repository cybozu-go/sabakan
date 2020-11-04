# Makefile for sabakan

GO_FILES=$(shell find -name '*.go' -not -name '*_test.go')
BUILT_TARGET=sabakan sabactl sabakan-cryptsetup
ETCD_DIR = /tmp/neco-etcd

.PHONY: all
all: test

.PHONY: build
build: $(BUILT_TARGET)
$(BUILT_TARGET): $(GO_FILES)
	go build ./pkg/$@

.PHONY: start-etcd
start-etcd:
	systemd-run --user --unit neco-etcd.service etcd --data-dir $(ETCD_DIR)

.PHONY: stop-etcd
stop-etcd:
	systemctl --user stop neco-etcd.service

.PHONY: code-check
code-check:
	test -z "$$(gofmt -s -l . | tee /dev/stderr)"
	staticcheck ./...
	test -z "$$(nilerr ./... 2>&1 | tee /dev/stderr)"
	test -z "$$(custom-checker -restrictpkg.packages=html/template,log $$(go list -tags='$(GOTAGS)' ./...) 2>&1 | tee /dev/stderr)"
	ineffassign .
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
	git add go.mod

.PHONY: clean
clean:
	rm -f $(BUILT_TARGET)

.PHONY: test-tools
test-tools: custom-checker staticcheck nilerr ineffassign

.PHONY: custom-checker
custom-checker:
	if ! which custom-checker >/dev/null; then \
		cd /tmp; env GOFLAGS= GO111MODULE=on go get github.com/cybozu/neco-containers/golang/analyzer/cmd/custom-checker; \
	fi

.PHONY: staticcheck
staticcheck:
	if ! which staticcheck >/dev/null; then \
		cd /tmp; env GOFLAGS= GO111MODULE=on go get honnef.co/go/tools/cmd/staticcheck; \
	fi

.PHONY: nilerr
nilerr:
	if ! which nilerr >/dev/null; then \
		cd /tmp; env GOFLAGS= GO111MODULE=on go get github.com/gostaticanalysis/nilerr/cmd/nilerr; \
	fi

.PHONY: ineffassign
ineffassign:
	if ! which ineffassign >/dev/null; then \
		cd /tmp; env GOFLAGS= GO111MODULE=on go get github.com/gordonklaus/ineffassign; \
	fi
