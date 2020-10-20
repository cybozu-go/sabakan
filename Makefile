# Makefile for sabakan

GO_FILES=$(shell find -name '*.go' -not -name '*_test.go')
BUILT_TARGET=sabakan sabactl sabakan-cryptsetup
ETCD_DIR = /tmp/neco-etcd

all: test

build: $(BUILT_TARGET)
$(BUILT_TARGET): $(GO_FILES)
	go build ./pkg/$@

start-etcd:
	systemd-run --user --unit neco-etcd.service etcd --data-dir $(ETCD_DIR)

stop-etcd:
	systemctl --user stop neco-etcd.service

test: build
	test -z "$$(gofmt -s -l . | tee /dev/stderr)"
	staticcheck ./...
	test -z "$$(nilerr ./... 2>&1 | tee /dev/stderr)"
	test -z "$$(custom-checker -restrictpkg.packages=html/template,log $$(go list -tags='$(GOTAGS)' ./...) 2>&1 | tee /dev/stderr)"
	ineffassign .
	go test -race -v ./...
	go vet ./...

e2e: build
	RUN_E2E=1 go test -v -count=1 ./e2e

mod:
	go mod tidy
	git add go.mod

clean:
	rm -f $(BUILT_TARGET)

.PHONY: all build start-etcd stop-etcd test e2e mod clean
