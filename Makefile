
LDFLAGS      := -w -s
MODULE       := github.com/figment-networks/indexer-manager
VERSION_FILE ?= ./VERSION


# Git Status
GIT_SHA ?= $(shell git rev-parse --short HEAD)

ifneq (,$(wildcard $(VERSION_FILE)))
VERSION ?= $(shell head -n 1 $(VERSION_FILE))
else
VERSION ?= n/a
endif

all: generate build-proto build

.PHONY: generate
generate:
	go generate ./...

.PHONY: build
build: LDFLAGS += -X $(MODULE)/cmd/manager/config.Timestamp=$(shell date +%s)
build: LDFLAGS += -X $(MODULE)/cmd/manager/config.Version=$(VERSION)
build: LDFLAGS += -X $(MODULE)/cmd/manager/config.GitSHA=$(GIT_SHA)
build:
	$(info building manager binary as ./manager_bin)
	go build -o manager_bin -ldflags '$(LDFLAGS)' ./cmd/manager

.PHONY: build-proto
build-proto:
	@protoc -I ./ --go_opt=paths=source_relative --go_out=plugins=grpc:. ./manager/proto/indexer.proto
	@mkdir -p ./manager/worker/transport/grpc/indexer
	@mkdir -p ./manager/transport/grpc/indexer
	@cp ./manager/proto/indexer.pb.go ./manager/worker/transport/grpc/indexer/
	@cp ./manager/proto/indexer.pb.go ./manager/transport/grpc/indexer/
