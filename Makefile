
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

all: generate build

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

