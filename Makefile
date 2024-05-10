ifndef VERSION
	VERSION = ${shell git rev-parse --abbrev-ref HEAD}-${shell git rev-parse --short HEAD}

endif

ifeq (,$(shell go env GOBIN))
	GOBIN=$(shell go env GOPATH)/bin
else
	GOBIN=$(shell go env GOBIN)
endif
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)
GOLANG_CI_LINT = $(LOCALBIN)/golangci-lint
GOLANG_CI_LINT_VERSION ?= v1.54.2

.PHONY: lint
lint:
	GOBIN=$(LOCALBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANG_CI_LINT_VERSION)
	$(LOCALBIN)/golangci-lint run -v

FLAGS = -ldflags '-s -w -X github.com/kyma-project/modulectl/cmd/modulectl/version.Version=$(VERSION)'

.PHONY: resolve
resolve:
	go mod tidy

.PHONY: validate
validate:
	./hack/verify-generated-docs.sh

.PHONY: build
build: build-windows build-linux build-darwin build-windows-arm build-linux-arm build-darwin-arm

# AMD based chipsets
.PHONY: build-windows
build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ./bin/modulectl.exe $(FLAGS) ./cmd

.PHONY: build-darwin
build-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ./bin/modulectl-darwin $(FLAGS) ./cmd

.PHONY: build-linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/modulectl-linux $(FLAGS) ./cmd

# ARM based chipsets
.PHONY: build-windows-arm
build-windows-arm:
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -o ./bin/modulectl-arm.exe $(FLAGS) ./cmd

.PHONY: build-darwin-arm
build-darwin-arm:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o ./bin/modulectl-darwin-arm $(FLAGS) ./cmd

.PHONY: build-linux-arm
build-linux-arm:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ./bin/modulectl-linux-arm $(FLAGS) ./cmd

.PHONY: docs
docs:
	rm -f ./docs/gen-docs/*
	go run ./cmd/gendocs/gendocs.go
