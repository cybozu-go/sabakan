GO_FILES=$(shell find -name '*.go' -not -name '*_test.go')
BUILT_TARGET=sabakan sabactl

.DEFAULT_GOAL := build

build: $(BUILT_TARGET)
$(BUILT_TARGET): $(GO_FILES)
	go build ./cmd/$@

e2e: $(BUILT_TARGET)
	RUN_E2E=1 go test -v -count=1 ./e2e

clean:
	rm -f $(BUILT_TARGET)

.PHONY: build clean e2e
